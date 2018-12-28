package common

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/em-lib-common-usage/fileutil"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrToUpdateNil error message for invalid erro passed as parameter
	ErrToUpdateNil = errors.New("Invalid error to update Step. Must be not nil")
	// ErrNoFileToTransform error message for when there is no available file to be transformed
	ErrNoFileToTransform = errors.New("There is no file available for Transform Step")
)

// Options holds common configuration variables to be used by the library components
type Options struct {
	SourceID     int64
	StepTypeID   int64
	InputFolder  string
	OutputFolder string
	User         string
	FileFormat   fileutil.FileFormat
}

// StepService defines the interface for common process applied to steps entities.
type StepService interface {

	/*
		RestartStep changes Step status to IN PROGRESS, set file ID to nil and start date to Now()
	*/
	RestartStep(step db.Step) (db.Step, error)

	/*
		UpdateStepToSuccessAndCreateFile creates a File entity and update the current Step to success
		Most of the method happens in a transaction. The last task is to fetch the new File entity and this task is
		taken after the transaction commit. Therefore this method might retrun a sql.ErrNoRows
	*/
	UpdateStepToSuccessAndCreateFile(step db.Step, fileMeta fileutil.FileMeta) (db.Step, db.File, error)

	/*
		UpdateStepToSuccessWithFileID updates the step to success with the given file ID
	*/
	UpdateStepToSuccessWithFileID(step db.Step, fileID int64) (db.Step, error)

	/*
		UpdateStepToError upadtes the given step to Error Status and set the error message with the Error String from
		the given error. If any error happens during this update, an new error is returned. If no error occurs, the updated
		step and a nil error are returned.
	*/
	UpdateStepToError(step db.Step, err error) (db.Step, error)

	/*
		FetchUnprocessedFileByStepType returns a file and the step that created this file, if the file is not processed by the
		given step type
	*/
	FetchUnprocessedFileByStepType(stepType db.StepType) (io.ReadCloser, db.Step, error)
}

// StepServiceImpl implements the StepService Interface
type StepServiceImpl struct {
	Options
	billingDB *db.BillingDB
}

// NewStepService creates a new StepService implementation
func NewStepService(billingDB *db.BillingDB, options Options) *StepServiceImpl {

	i := &StepServiceImpl{
		Options:   options,
		billingDB: billingDB,
	}

	return i
}

/*
RestartStep is the StepServiceImpl implementation for RestartStep method defined in StepService Inteface
*/
func (i *StepServiceImpl) RestartStep(step db.Step) (db.Step, error) {

	logger := log.WithField("step_to_restart", step)
	logger.Debug("Restarting step")

	stepPrevStatus := step.Status
	stepPrevError := step.Error
	stepPrevFileID := step.FileID
	stepPrevStartDate := step.StarteDate

	logger.Debug("Starting transaction to update Step")

	var err error
	var tx db.BillingTx

	//Start a new Transaction. Any Insert and Update is taken inside a transaction
	tx, err = i.billingDB.NewTransaction()
	if err != nil {
		return step, i.billingDB.HandleTx(tx, err)
	}

	step.Status = db.StepStatusInProgress.ToString()
	step.Error = ""
	step.FileID = sql.NullInt64{Int64: 0, Valid: false} //Set file id to null
	step.StarteDate = time.Now()

	//Updates step so it is restarted
	err = i.billingDB.StepRepository.Update(step, i.Options.User, tx)
	err = i.billingDB.HandleTx(tx, err)
	if err != nil {
		step.Status = stepPrevStatus
		step.Error = stepPrevError
		step.FileID = stepPrevFileID
		step.StarteDate = stepPrevStartDate
	}

	logger.Debug("Step Restarting action finished")

	return step, err

}

/*
UpdateStepToSuccessAndCreateFile is the StepServiceImpl implementation for UpdateStepToSuccessAndCreateFile method defined in StepService Inteface
*/
func (i *StepServiceImpl) UpdateStepToSuccessAndCreateFile(step db.Step, fileMeta fileutil.FileMeta) (db.Step, db.File, error) {

	logger := log.WithFields(log.Fields{
		"source_id": i.Options.SourceID,
		"step":      step,
		"file_info": fileMeta,
	})

	logger.Debug("Starting transaction to update Step to success")

	var file db.File
	var err error
	var tx db.BillingTx
	var fileID int64

	//Start a new Transaction. Any Insert and Update is taken inside a transaction
	tx, err = i.billingDB.NewTransaction()
	if err != nil {
		return step, file, i.billingDB.HandleTx(tx, err)
	}

	//Check if there is a file in DB with same checksum. If it has, use it
	file, err = i.billingDB.FileRepository.FetchByChecksum(fileMeta.Checksum)
	if err != nil && err != sql.ErrNoRows {
		return step, file, i.billingDB.HandleTx(tx, err)
	}

	//If there is no file with the actual checksum, creates a new one
	if err == sql.ErrNoRows {

		//Creates a new file in the Billing database
		fileID, err = i.billingDB.FileRepository.Create(fileMeta.Checksum, fileMeta.FileName, fileMeta.FilePath, i.Options.User, tx)
		if err != nil {
			return step, file, i.billingDB.HandleTx(tx, err)
		}

	} else {
		fileID = file.ID
	}

	//Save previous status for recovering in case of error
	prevStatus := step.Status
	prevError := step.Error
	prevFileID := step.FileID
	prevEndDate := step.EndDate

	//Updates the step with the new file metadata
	step.FileID = sql.NullInt64{Int64: fileID, Valid: true}
	step.Status = db.StepStatusFinished.ToString()
	step.Error = ""
	step.EndDate = time.Now()

	err = i.billingDB.StepRepository.Update(step, i.Options.User, tx)
	if err != nil {
		step.Status = prevStatus
		step.Error = prevError
		step.FileID = prevFileID
		step.EndDate = prevEndDate
		return step, file, i.billingDB.HandleTx(tx, err)
	}

	//Commits or Rollback if it has error
	err = i.billingDB.HandleTx(tx, err)
	if err != nil {
		step.Status = prevStatus
		step.Error = prevError
		step.FileID = prevFileID
		step.EndDate = prevEndDate
		return step, file, err
	}

	//If a new file was created, file.ID will be 0
	if file.ID < 1 {
		//Retrieves the new file to return it. It might return a sql.ErrNoRows. It is expected from this method
		//to return a Step and a File or an error
		file, err = i.billingDB.FileRepository.Fetch(fileID)
		if err != nil {
			return step, file, i.billingDB.HandleTx(tx, err)
		}

	}

	logger.Debug("Transaction to update Step finished")
	return step, file, err
}

//UpdateStepToSuccessWithFileID updates the step to success with the given file ID
func (i *StepServiceImpl) UpdateStepToSuccessWithFileID(step db.Step, fileID int64) (db.Step, error) {

	logger := log.WithFields(log.Fields{
		"source_id": i.Options.SourceID,
		"step":      step,
		"file_id":   fileID,
	})

	logger.Debug("Starting transaction to update Step to success with file ID")

	var err error
	var tx db.BillingTx

	//Start a new Transaction. Any Insert and Update is taken inside a transaction
	tx, err = i.billingDB.NewTransaction()
	if err != nil {
		return step, i.billingDB.HandleTx(tx, err)
	}

	//Save previous status for recovering in case of error
	prevStatus := step.Status
	prevError := step.Error
	prevFileID := step.FileID
	prevEndDate := step.EndDate

	//Updates the step with the new file metadata
	step.FileID = sql.NullInt64{Int64: fileID, Valid: true}
	step.Status = db.StepStatusFinished.ToString()
	step.Error = ""
	step.EndDate = time.Now()

	err = i.billingDB.StepRepository.Update(step, i.Options.User, tx)
	if err != nil {
		step.Status = prevStatus
		step.Error = prevError
		step.FileID = prevFileID
		step.EndDate = prevEndDate
		return step, i.billingDB.HandleTx(tx, err)
	}

	//Commits or Rollback if it has error
	err = i.billingDB.HandleTx(tx, err)
	if err != nil {
		step.Status = prevStatus
		step.Error = prevError
		step.FileID = prevFileID
		step.EndDate = prevEndDate
		return step, err
	}

	logger.Debug("Transaction to update Step with file ID finished")
	return step, err
}

/*
UpdateStepToError is the StepServiceImpl implementation for UpdateStepToError method defined in StepService Inteface
*/
func (i *StepServiceImpl) UpdateStepToError(step db.Step, err error) (db.Step, error) {

	logger := log.WithFields(log.Fields{
		"source_id": i.Options.SourceID,
		"step":      step,
		"error":     err,
	})

	logger.Debug("Starting transaction to update Step to error")

	if err == nil {
		return step, ErrToUpdateNil
	}

	prevStepStatus := step.Status
	prevStepError := step.Error
	prevEndDate := step.EndDate

	var retErr error
	var tx db.BillingTx

	//Start a new Transaction. Any Insert and Update is taken inside a transaction
	tx, retErr = i.billingDB.NewTransaction()

	if retErr != nil {
		return step, i.billingDB.HandleTx(tx, retErr)
	}

	step.Status = db.StepStatusError.ToString()
	step.Error = err.Error()
	step.EndDate = time.Now()

	//Update the step with the Error Message and Error status
	retErr = i.billingDB.StepRepository.Update(step, i.Options.User, tx)

	retErr = i.billingDB.HandleTx(tx, retErr)

	if retErr != nil {
		//Revert step changes
		step.Status = prevStepStatus
		step.Error = prevStepError
		step.EndDate = prevEndDate
	}

	logger.Debug("Transaction to update Step to error finished")
	return step, retErr

}

/*
	FetchUnprocessedFileByStepType return the file' data to be processed as the step that created this file.
*/
func (i *StepServiceImpl) FetchUnprocessedFileByStepType(stepType db.StepType) (io.ReadCloser, db.Step, error) {

	logger := log.WithFields(log.Fields{
		"step_type":    stepType,
		"input_folder": i.Options.InputFolder,
	})

	logger.Debug("Starting fetching file data from source folder")

	var fileData io.ReadCloser
	var previousStep db.Step
	var errRet = ErrNoFileToTransform

	//Listing all files
	filesInfo, err := fileutil.ListFilesInFolder(i.Options.InputFolder)
	if err != nil {
		return fileData, previousStep, err
	}

	filesInFolder := fmt.Sprintf("There are %d files in %s folder", len(filesInfo), i.Options.InputFolder)
	logger.Info(filesInFolder)
	//Sort the files info to put oldest ones first
	sort.Slice(filesInfo, func(i, j int) bool {

		return filesInfo[i].ModTime().Before(filesInfo[j].ModTime())

	})

	//Iterates through files info to check if an file is ready to process.
	for _, fileInfo := range filesInfo {

		fileFullname := filepath.Join(i.Options.InputFolder, filepath.Clean(fileInfo.Name()))

		//First, gets the current file checksum
		checksum, err := fileutil.GetChecksumByName(fileFullname)
		if err != nil {
			return fileData, previousStep, err
		}

		//Retrives file information in billing db
		dbFile, err := i.billingDB.FileRepository.FetchByChecksum(checksum)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			} else {
				return fileData, previousStep, err
			}
		}

		//Search for the step that created this file and this file AND that HAS NOT been processed by the actual step
		previousStep, err = i.billingDB.FetchByFileIDAndNoNextStep(dbFile.ID, stepType.ID, i.Options.User)
		if err != nil {
			if err == sql.ErrNoRows { //If found nothing, file is not elegible to be processed. Move to next one
				continue
			} else {
				return fileData, previousStep, err
			}
		}

		logger = logger.WithFields(log.Fields{
			"previous_step_id":   previousStep.ID,
			"previous_Step_type": previousStep.StepTypeID,
			"file_id":            dbFile.ID,
			"file_checksum":      dbFile.Checksum,
			"file_name":          dbFile.Name,
		})

		logger.Info("File elegible to be transformed")

		file, err := os.Open(fileFullname)
		if err != nil {
			return fileData, previousStep, err

		}

		//Actual file is elegible to process, create the reader for the file data
		fileData = ioutil.NopCloser(file)
		errRet = nil
		break
	}

	logger.Debug("Fetching file data from source folder finished")
	return fileData, previousStep, errRet
}
