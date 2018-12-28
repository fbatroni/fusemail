// Package db has the interfaces definitions to crate Billing Database Repositories
package db

import (
	"database/sql"

	"bitbucket.org/fusemail/fm-lib-commons-golang/health"
)

//Interface Definitions

//CollectionJobRepository defines the methods to oparates SQL statements with Collection Job Table
type CollectionJobRepository interface {
	/*
		Create creates a new Collection job row in billing.collection_job and returns its ID
		Params:
			- sourceID int64: The source where this Collection Job will fetch the usage data
			- currentStepTypeID int64: The initial step type for the given source.
			- user: The user that is requesting the usage collection
			- *BillingTx - Database transaction in which this operation is inserted

		This method should be called in a transaction.
	*/
	Create(sourceID int64, currentStepTypeID int64, user string, tx BillingTx) (int64, error)
	/*
		Fetch returns a Collection Job identified by the collectionJobID parameter
		Params:
			- collectionJobID int64: The Collection Job Unique identify

		If no Collection Job is found, return sql.ErrNoRows
	*/
	Fetch(collectionJobID int64) (CollectionJob, error)
	/*
		ListByStatus returns all Collection Jobs with status equal to the status passed as parameter
		Params:
			- collectionJobStatus CollectionJobStatus: The Collection Job Status to be use as filter to fetch
			Collection Jobs

	*/
	ListByStatus(collectionJobStatus CollectionJobStatus) ([]CollectionJob, error)
	/*
		Update updates the given Collection Job in the database
		Params:
			- collectionJob CollectionJob: Collection Job to be updated.
			- *BillingTx - Database transaction in which this operation is inserted
	*/
	Update(collectionJob CollectionJob, user string, tx BillingTx) error
}

//StepRepository defines the methods to oparates SQL statements with Step Table
type StepRepository interface {
	/*
		Create creates a new Step row in billing.step and returns its ID
		Params:
			- StepTypeID int64: The Step type ID
			- CollectionJobID int64: The Collection Job that owns this step
			- user: The user that is requesting the usage collection
	*/
	Create(stepTypeID int64, collectionJobID int64, user string, tx BillingTx) (int64, error)
	/*
	   Fetch returns a Step identified by the stepID parameter
	   Params:
	   	- stepID int64: The Step Unique identify

	   If no Step is found, return sql.ErrNoRows
	*/
	Fetch(stepID int64) (Step, error)
	/*
		ListByCollectionJobID returns all Steps with status equal to the status passed as parameter
		Params:
			- stepStatus StepStatus: The Step Status to be use as filter to fetch
			Steps
	*/
	ListByCollectionJobID(collectionJobID int64) ([]Step, error)
	/*
		FetchUnfinishedBySourceAndType returns an unfinished step for with the same type as the given parameter and
		for a collection job for the given source. Step is also filtered by the user who created it.
		Params:
			- sourceID int64: The source where the data is being collected and to which the step is owned
			- stepTypeID int: The type of unfinished step
			- user string: User who created the step

		If no Step is found, return sql.ErroNoRows
	*/
	FetchUnfinishedBySourceAndType(sourceID int64, stepTypeID int64, user string) (Step, error)

	/*
		FetchByFileIDAndNoNextStep retrives the Step that created the given file IF there is no
		next step done for the Collection Job in which the steps are inserted. For example,
		A DownloadStep X (ID= 110, CollectionJob = 10, StepStypeID = 1, FileID = 100) created a file A with ID 100.
		If this method receives as parameter FileID = 100, nextStepStypeID = 2. It will return the
		DownloadStep X IF there is NO Step with StepType = 2, CollectionJob = 10 and Status FINISHED.

		Param nextStepTypeID is the step that must be NOT DONE for the given file.
	*/
	FetchByFileIDAndNoNextStep(fileID int64, nextStepTypeID int64, user string) (Step, error)

	/*
		Update updates the given Step in the database
		Params:
			- step Step: Step to be updated.
			- tx *BillingTx - Database transaction in which this operation is inserted
	*/
	Update(step Step, user string, tx BillingTx) error
}

//StepTypeRepository defines the methods to oparates SQL statements with Step Type Table
type StepTypeRepository interface {
	/*
		Fetch gets a specific StepType based on the ID passed as parameter
		Params:
			- StepTypeID int64: The Step Type unique idendify

		If no Step is found, return sql.ErroNoRows
	*/
	Fetch(stepTypeID int64) (StepType, error)
	/*
		GetStepBySource returns all Steps Types owned by a given source
		Params:
			- sourceID int64: The specific source that owns the step types to be fetched
	*/
	ListBySource(sourceID int64) ([]StepType, error)
	/*
		GetStepTypeSourceAndByName returns a StepType for a specific Source based on the step type name
		passed as a parameter.
		Params:
			- sourceID int64: The specific source that owns the step type to be fetched
			- stepType StepTypeName: The step type name to be fetched

		If no Step is found, return sql.ErroNoRows
	*/
	FetchBySourceAndName(sourceID int64, stepType StepTypeName) (StepType, error)
	/*
		FetchBySourceAndOrder returns the step type that matches with the given source and step name.
		Params:
			- step Step: Step to be updated.

		If no Step is found, return sql.ErroNoRows
	*/
	FetchBySourceAndOrder(sourceID int64, order int) (StepType, error)
}

//FileRepository defines the methods to oparates SQL statements with File Table
type FileRepository interface {
	/*
		Create creates a new File row in billing.collection_job and returns its ID
			Params:
				- checksum string:
				- name string:
				- filePath string:
	*/
	Create(checksum string, name string, filePath string, user string, tx BillingTx) (int64, error)
	/*
		Fetch returns a File identified by the fileID parameter
			Params:
				- fileID int64: The File Unique identify

			If no File is found, return sql.ErrNoRows
	*/
	Fetch(fileID int64) (File, error)
	/*
		FetchByChecksum returns all File that checksum is equals to the checksum passed as parameter
			Params:
				- checksum string: The File checksum
			If no File is found, return sql.ErrNoRows
	*/
	FetchByChecksum(checksum string) (File, error)
}

// BillingDB holds all database interface repositories and a interface to start transactions.
type BillingDB struct {
	BillingTransactionDB
	CollectionJobRepository
	StepRepository
	StepTypeRepository
	FileRepository
}

//Check is the implementation for Health interface from fusemail go common lib
func (i *BillingDB) Check() (map[string]interface{}, error) {
	return i.BillingTransactionDB.Check()
}

//HandleTx process the transaction based on the given error. If the error is nil, the transaction is committed. Is there is an error
//the transactions is rolled back.
func (i *BillingDB) HandleTx(tx BillingTx, err error) error {

	var txErr, retErr error
	retErr = err

	if tx != nil {

		if err != nil {
			txErr = tx.RollbackTransaction()
		} else {
			txErr = tx.CommitTransaction()
		}

		if txErr != nil {
			retErr = txErr
		}
	}

	return retErr
}

//BillingTransactionDB defines a inferface for a Billing transactional database.
type BillingTransactionDB interface {
	/*
		NewTransaction creates a new transaction using the database implementation, wrap it up in a
		BillingTx interface and return it.
	*/
	NewTransaction() (BillingTx, error)

	/*
		BulkInsert insert all given entities in one single insert statment.
	*/
	BulkInsert(sqlBase string, sqlValues string, entities [][]interface{}) error

	// Connetc to the database
	Connect() error
	health.Depender
}

/*
BillingTx wraps an Sql Transaction Implementation to wrap and execute the following methods:
	-	Select
	-	QueryRow
	-	Exec
	-	MustExec
	- 	Commit
	-	Rollback

*/
type BillingTx interface {

	/*
		RunExec warps Select (for multiple returns) regardless the database implementation.
		Mapps the return to the destiny interface
	*/
	RunSelect(dest interface{}, query string, args ...interface{}) error
	/*
		RunExec warps QueryRow (for single return) regardless the database implementation.
		Mapps the return to the destiny interface
	*/
	RunQueryRow(dest interface{}, query string, args ...interface{}) error
	/*
		RunExec warps Exec functions (for inserts and updates) regardless the database implementation.
	*/
	RunExec(query string, args ...interface{}) (sql.Result, error)

	/*
		RunExec warps MustExec functions (for inserts and updates) regardless the database implementation.
	*/
	RunMustExec(query string, args ...interface{}) sql.Result

	/*
		Commit an closes a active transaction.
	*/
	CommitTransaction() error

	/*
		Rollback an closes a active transaction.
	*/
	RollbackTransaction() error
}
