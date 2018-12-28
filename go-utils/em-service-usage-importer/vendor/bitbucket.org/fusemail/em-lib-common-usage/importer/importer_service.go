package importer

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"time"

	"bitbucket.org/fusemail/em-lib-common-usage/common"
	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/fm-lib-commons-golang/trace"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrInvalidStepTypeID error message for Step Type with different Source ID for the actual Source
	ErrInvalidStepTypeID = errors.New("Invalid Step Type ID for the actual Source")
	//ErrorNewInvalidParams error message for any invalid parameter passed in New method for a TransformerStep
	ErrorNewInvalidParams = errors.New("New Transformer Step Invalid parameters")
	//ErrorNewInvalidOptions error message for any invalid parameter passed in New method for a TransformerStep
	ErrorNewInvalidOptions = errors.New("New Transformer Step Invalid Options")
)

//ImporterStepService implements the DownloadStep Interface
type ImporterStepService struct {
	common.Options
	billingDB   *db.BillingDB
	stepService common.StepService
	trace.Tracers
}

/*
New creates a new Import Step Service with a common implementation for all vendors.
*/
func New(stepService common.StepService, options common.Options, billingDB *db.BillingDB) (*ImporterStepService, error) {

	logger := log.WithField("action", "New ImporterStepService")

	logger.Debug("Creating New ImporterStep")

	var err error

	if !validateOptions(options) {
		err = ErrorNewInvalidOptions
	}

	if !validateBillingDB(billingDB) {
		err = ErrorNewInvalidOptions
	}

	if err != nil {
		return nil, err
	}

	i := &ImporterStepService{
		Options:     options,
		billingDB:   billingDB,
		stepService: stepService,
	}

	logger.Debug("New DownloadStep Created")

	return i, nil

}

func validateOptions(options common.Options) bool {

	if len(options.User) == 0 {
		return false
	}
	if len(options.InputFolder) == 0 {
		return false
	}
	if options.SourceID < 1 {
		return false
	}

	return true
}

func validateBillingDB(billingDB *db.BillingDB) bool {

	if billingDB.BillingTransactionDB == nil {
		return false
	}
	if billingDB.StepRepository == nil {
		return false
	}
	if billingDB.StepTypeRepository == nil {
		return false
	}
	if billingDB.FileRepository == nil {
		return false
	}

	return true
}

/*
ExecuteImportStep executes all tasks in the Import Step. The tasks are:
1 - Retrieve a file to be proccessed
2 - Create or restart a Import Step
3 - Import the file into the proper vendor usage table
4 - Update the Import Step with the result status
*/
func (i *ImporterStepService) ExecuteImportStep(vendorMapper VendorMapper) error {

	t := &trace.Trace{
		Action: "Execute Importer Step",
		Start:  time.Now(),
		Logger: log.WithFields(log.Fields{
			"source_id":    i.Options.SourceID,
			"step_type_id": i.Options.StepTypeID,
			"input_folder": i.Options.InputFolder,
		}),
	}
	defer i.Trace(t)

	t.Logger.Debug("Starting Import Step Execution")

	var updateErr error

	//Get the Step Type based on Options.SourceID and Options.StepTypeID. The returned step from StepTypeID fetching
	//Must have same Source ID as the Options value
	stepType, err := i.getStepType()
	if err != nil {
		return err
	}

	/*
		Fetch the data to be transformed from the elegible file
	*/
	fileData, previousStep, err := i.stepService.FetchUnprocessedFileByStepType(stepType)
	if err != nil {
		if err != common.ErrNoFileToTransform {
			t.Error = err
		}
		return err
	}

	/*
		Start the Import Step (save/update register in the database)
	*/
	step, err := i.startImportStep(stepType, previousStep.CollectionJobID)
	if err != nil {
		t.Error = err
		return err
	}

	/*
		Import data
	*/
	err = i.executeImport(fileData, vendorMapper)
	if err != nil {
		t.Error = err
		return err
	}

	//Updates the Transformer Step to finished, creates a file in the database and adds the file info to the step.
	t.Logger.Debug("Updating Import Step status to Successfully Finished")
	_, err = i.stepService.UpdateStepToSuccessWithFileID(step, previousStep.FileID.Int64)
	if err != nil {
		_, updateErr = i.stepService.UpdateStepToError(step, err)
		if updateErr != nil {
			err = updateErr
		}
		t.Error = err
		return err
	}

	t.Logger.Debug("Import Step Execution Finished")

	return nil
}

/*
	getStepType retrieves an StepType with based on Options.StepTypeID and Options.SourceID.
*/
func (i *ImporterStepService) getStepType() (db.StepType, error) {

	Logger := log.WithFields(log.Fields{
		"step_type_id": i.Options.StepTypeID,
		"source_id":    i.Options.SourceID,
	})

	Logger.Debug("Fetching Step Type")

	stepType := db.StepType{}
	var err error

	stepType, err = i.billingDB.StepTypeRepository.Fetch(i.Options.StepTypeID)
	if err != nil {
		return stepType, err
	}
	if stepType.SourceID != i.Options.SourceID {
		return stepType, ErrInvalidStepTypeID
	}

	Logger.Debug("Fetching Step Type Finished")

	return stepType, err
}

/*
	startImportStep verifies if there is an unfinished Import Step (based on the stepType parameter)
	for the actual source. If there is an unfinished step, the step is restarted and returned. Otherwise,
	a new Import Step is created.
*/
func (i *ImporterStepService) startImportStep(stepType db.StepType, collectionJobID int64) (db.Step, error) {

	t := &trace.Trace{
		Action: "Create/Restart Import Step",
		Start:  time.Now(),
		Logger: log.WithFields(log.Fields{
			"step_type_id":      stepType.ID,
			"collection_job_id": collectionJobID,
			"source_id":         stepType.SourceID,
		}),
	}
	defer i.Trace(t)

	t.Logger.Debug("Creating/Updating a Import Step")

	step := db.Step{}
	var err error

	//Retrives unfinished Transform Step for the actual Job
	step, err = i.billingDB.FetchUnfinishedBySourceAndType(stepType.SourceID, stepType.ID, i.Options.User)
	if err != nil && err != sql.ErrNoRows {
		return step, err
	}

	//If has a unfinished step, restart it and return
	if err == nil && step.CollectionJobID == collectionJobID {

		t.Logger = t.Logger.WithFields(log.Fields{
			"step_id":           step.ID,
			"collection_job_id": step.CollectionJobID,
			"previous_status":   fmt.Sprintf("%s : %s", step.Status, step.Error),
		})

		t.Logger.Info("A Unfinished Import Step will be restarted")
		return i.stepService.RestartStep(step)

	}

	return i.createNewImportStep(stepType, collectionJobID)
}

func (i *ImporterStepService) createNewImportStep(stepType db.StepType, collectionJobID int64) (db.Step, error) {

	var step db.Step

	tx, err := i.billingDB.NewTransaction()
	if err != nil {
		return step, i.billingDB.HandleTx(tx, err)
	}

	stepID, err := i.billingDB.StepRepository.Create(stepType.ID, collectionJobID, i.Options.User, tx)
	if err != nil {
		return step, i.billingDB.HandleTx(tx, err)
	}

	//Fetch collection job to update its current step
	collectionJob, err := i.billingDB.CollectionJobRepository.Fetch(collectionJobID)
	if err != nil {
		return step, i.billingDB.HandleTx(tx, err)
	}

	//Updates the collection job to the current step
	collectionJob.CurrentStepTypeID = stepType.ID
	err = i.billingDB.CollectionJobRepository.Update(collectionJob, i.Options.User, tx)

	//Commits or Rollback if it has error
	err = i.billingDB.HandleTx(tx, err)
	if err != nil {
		return step, err
	}

	return i.billingDB.StepRepository.Fetch(stepID)
}

func (i *ImporterStepService) executeImport(data io.ReadCloser, vendorMapper VendorMapper) error {

	r := csv.NewReader(data)

	bulkFlush := 10
	bulk := [][]interface{}{}

	keys := make([]string, 0, len(vendorMapper.ColumnMappers))
	for k := range vendorMapper.ColumnMappers {
		keys = append(keys, k)
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		usage := make([]interface{}, len(keys)+1)

		//Prepare the usage to be inserted in the bulk
		for _, key := range keys {
			val := record[vendorMapper.ColumnMappers[key].CsvIndex]
			usage[vendorMapper.ColumnMappers[key].TableIndex] = val
		}
		usage[len(usage)-1] = i.Options.User

		bulk = append(bulk, usage)

		if len(bulk) == bulkFlush {

			err = i.billingDB.BulkInsert(vendorMapper.SqlBase, vendorMapper.SqlValues, bulk)
			if err != nil {
				return err
			}
			//Flush the bulk
			bulk = bulk[:0]
		}
	}

	if len(bulk) > 0 {
		err := i.billingDB.BulkInsert(vendorMapper.SqlBase, vendorMapper.SqlValues, bulk)
		if err != nil {
			return err
		}
	}

	bulk = nil

	return nil
}
