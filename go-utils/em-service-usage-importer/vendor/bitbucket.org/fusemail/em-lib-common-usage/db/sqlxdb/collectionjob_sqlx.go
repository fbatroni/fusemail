package sqlxdb

import (
	"database/sql"
	"time"

	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/fm-lib-commons-golang/deps"
	"bitbucket.org/fusemail/fm-lib-commons-golang/trace"
)

//CollectionJobSqlxRepository implements the CollectionJobRepository interface
type CollectionJobSqlxRepository struct {
	trace.Tracers
	SQLXDB *deps.SQLXDB
}

//Create implements the CollectionJobRepository.Create method
func (i *CollectionJobSqlxRepository) Create(sourceID int64, currentStepTypeID int64, user string, tx db.BillingTx) (int64, error) {

	start := time.Now()
	var traceError error

	query := "INSERT INTO collection_job (source_id, current_step_type_id, created_by) VALUES(?, ?, ?)"
	args := []interface{}{sourceID, currentStepTypeID, user}

	var id int64
	id, traceError = tx.RunMustExec(query, args...).LastInsertId()

	i.SQLXDB.Trace(query, start, traceError)
	return id, traceError
}

//Fetch implements the CollectionJobRepository.Fetch method
func (i *CollectionJobSqlxRepository) Fetch(collectionJobID int64) (db.CollectionJob, error) {

	start := time.Now()
	var traceError error

	query := `SELECT collection_job_id, source_id, status, current_step_type_id, 
			  start_date, end_date
			  FROM collection_job
			  WHERE collection_job_id = ?`
	args := []interface{}{collectionJobID}

	collectionJob := db.CollectionJob{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&collectionJob)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return collectionJob, err
}

//ListByStatus implements the CollectionJobRepository.ListByStatus method
func (i *CollectionJobSqlxRepository) ListByStatus(collectionJobStatus db.CollectionJobStatus) ([]db.CollectionJob, error) {

	start := time.Now()
	var traceError error

	query := `SELECT collection_job_id, source_id, status, current_step_type_id, 
			  start_date, end_date
			  FROM collection_job
			  WHERE status = ?`

	args := []interface{}{collectionJobStatus}

	collectionJobs := []db.CollectionJob{}

	err := i.SQLXDB.DB.Select(&collectionJobs, query, args...)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)

	return collectionJobs, err
}

//Update implements the CollectionJobRepository.Update method
func (i *CollectionJobSqlxRepository) Update(collectionJob db.CollectionJob, user string, tx db.BillingTx) error {

	start := time.Now()

	query := `UPDATE collection_job
			  SET status = ?, current_step_type_id = ?, end_date =?, modified_by =?
			  WHERE collection_job_id= ?`

	args := []interface{}{
		collectionJob.Status,
		collectionJob.CurrentStepTypeID,
		time.Now(),
		user,
		collectionJob.ID,
	}

	_, err := tx.RunMustExec(query, args...).RowsAffected()
	i.SQLXDB.Trace(query, start, err)
	return err
}
