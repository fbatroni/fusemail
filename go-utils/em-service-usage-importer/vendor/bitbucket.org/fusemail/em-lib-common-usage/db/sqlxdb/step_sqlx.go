package sqlxdb

import (
	"database/sql"
	"time"

	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/fm-lib-commons-golang/deps"
	"bitbucket.org/fusemail/fm-lib-commons-golang/trace"
)

//StepSqlxRepository implements the StepRepository interface
type StepSqlxRepository struct {
	trace.Tracers
	SQLXDB *deps.SQLXDB
}

//Create implements the db.StepRepository.Create method
func (i *StepSqlxRepository) Create(stepTypeID int64, collectionJobID int64, user string, tx db.BillingTx) (int64, error) {

	start := time.Now()
	var traceError error

	query := "INSERT INTO step (step_type_id, collection_job_id, created_by) VALUES(?, ?, ?)"
	args := []interface{}{stepTypeID, collectionJobID, user}

	var id int64
	id, traceError = tx.RunMustExec(query, args...).LastInsertId()

	i.SQLXDB.Trace(query, start, traceError)
	return id, traceError
}

//Fetch implements the db.StepRepository.Fetch method
func (i *StepSqlxRepository) Fetch(stepID int64) (db.Step, error) {

	start := time.Now()
	var traceError error

	query := `SELECT step_id, step_type_id, collection_job_id, start_date, end_date, file_id, status, error
			  FROM step
			  WHERE step_id = ?`
	args := []interface{}{stepID}

	step := db.Step{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&step)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return step, err
}

//ListByCollectionJobID implements the db.StepRepository.ListByCollectionJobID method
func (i *StepSqlxRepository) ListByCollectionJobID(collectionJobID int64) ([]db.Step, error) {

	start := time.Now()
	var traceError error

	query := `SELECT step_id, step_type_id, collection_job_id, start_date, end_date, file_id, status, error
			  FROM step
			  WHERE collection_job_id = ?`
	args := []interface{}{collectionJobID}

	steps := []db.Step{}

	err := i.SQLXDB.DB.Select(&steps, query, args...)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return steps, err
}

//FetchUnfinishedBySourceAndType implements the db.StepRepository.FetchUnfinishedBySourceAndType method
func (i *StepSqlxRepository) FetchUnfinishedBySourceAndType(sourceID int64, stepTypeID int64, user string) (db.Step, error) {

	start := time.Now()
	var traceError error

	query := `SELECT s.step_id, s.step_type_id, s.collection_job_id, s.start_date, s.end_date, s.file_id, s.status, s.error
			  FROM step s
			  JOIN step_type st USING(step_type_id)
			  WHERE s.step_type_id = ? and s.created_by = ? and s.status <> ?
			  and st.source_id = ?`
	args := []interface{}{stepTypeID, user, db.StepStatusFinished.ToString(), sourceID}

	step := db.Step{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&step)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return step, err
}

//FetchByFileIDAndNoNextStep implements the db.StepRepository.FetchByFileIDAndNoNextStep method
func (i *StepSqlxRepository) FetchByFileIDAndNoNextStep(fileID int64, stepTypeID int64, user string) (db.Step, error) {

	start := time.Now()
	var traceError error

	query := `SELECT prevStep.step_id, prevStep.step_type_id, prevStep.collection_job_id, prevStep.start_date, prevStep.end_date, prevStep.file_id, prevStep.status, prevStep.error
	FROM step prevStep
	LEFT JOIN (SELECT step_id, collection_job_id, step_type_id FROM step WHERE step_type_id = ? and status = ?) as nextStep
	USING(collection_job_id)
	WHERE nextStep.step_id is null
	AND prevStep.file_id = ?
	AND prevStep.status = ?
	LIMIT 1` //This method will return only one step that match the query conditions.

	args := []interface{}{stepTypeID, db.StepStatusFinished.ToString(), fileID, db.StepStatusFinished.ToString()}

	step := db.Step{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&step)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return step, err
}

//Update implements the db.StepRepository.Update method
func (i *StepSqlxRepository) Update(step db.Step, user string, tx db.BillingTx) error {

	start := time.Now()
	var traceError error

	query := `UPDATE step
			  SET start_date =?, end_date = ?, file_id = ?, status =?, error =?, modified_by =?
			  WHERE step_id= ?`

	args := []interface{}{
		step.StarteDate,
		time.Now(),
		step.FileID,
		step.Status,
		step.Error,
		user,
		step.ID,
	}

	_, traceError = tx.RunMustExec(query, args...).RowsAffected()

	i.SQLXDB.Trace(query, start, traceError)
	return traceError
}
