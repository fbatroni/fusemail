package sqlxdb

import (
	"database/sql"
	"time"

	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/fm-lib-commons-golang/deps"
	"bitbucket.org/fusemail/fm-lib-commons-golang/trace"
)

// StepTypeSqlxRepository implements the db.StepTypeRepository
type StepTypeSqlxRepository struct {
	trace.Tracers
	SQLXDB *deps.SQLXDB
}

// Fetch implements the db.Fetch method
func (i *StepTypeSqlxRepository) Fetch(stepTypeID int64) (db.StepType, error) {

	start := time.Now()
	var traceError error

	query := `SELECT step_type_id, source_id, name, step_order FROM step_type WHERE step_type_id = ?`
	args := []interface{}{stepTypeID}

	stepType := db.StepType{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&stepType)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return stepType, err
}

// ListBySource implements the db.ListBySource method
func (i *StepTypeSqlxRepository) ListBySource(sourceID int64) ([]db.StepType, error) {

	start := time.Now()
	var traceError error

	query := `SELECT step_type_id, source_id, name, step_order  FROM step_type  WHERE source_id = ?`
	args := []interface{}{sourceID}

	stepTypes := []db.StepType{}

	err := i.SQLXDB.DB.Select(&stepTypes, query, args...)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return stepTypes, err
}

// FetchBySourceAndName implements the db.FetchBySourceAndName method
func (i *StepTypeSqlxRepository) FetchBySourceAndName(sourceID int64, stepTypeName db.StepTypeName) (db.StepType, error) {

	start := time.Now()
	var traceError error

	query := `SELECT step_type_id, source_id, name, step_order  FROM step_type  WHERE source_id = ? AND name = ?`
	args := []interface{}{sourceID, stepTypeName}

	stepType := db.StepType{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&stepType)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return stepType, err
}

// FetchBySourceAndOrder implements the db.FetchBySourceAndOrder method
func (i *StepTypeSqlxRepository) FetchBySourceAndOrder(sourceID int64, order int) (db.StepType, error) {

	start := time.Now()
	var traceError error

	query := `SELECT step_type_id, source_id, name, step_order FROM step_type  WHERE source_id = ? AND step_order = ?`
	args := []interface{}{sourceID, order}

	stepType := db.StepType{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&stepType)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return stepType, err
}
