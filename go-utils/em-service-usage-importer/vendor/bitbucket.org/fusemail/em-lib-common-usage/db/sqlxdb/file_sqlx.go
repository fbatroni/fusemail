package sqlxdb

import (
	"database/sql"
	"time"

	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/fm-lib-commons-golang/deps"
	"bitbucket.org/fusemail/fm-lib-commons-golang/trace"
)

//FileSqlxRepository implements the db.FileRepository interface
type FileSqlxRepository struct {
	trace.Tracers
	SQLXDB *deps.SQLXDB
}

// Create implements the db.FileRepository.Create method
func (i *FileSqlxRepository) Create(checksum string, name string, filePath string, user string, tx db.BillingTx) (int64, error) {

	start := time.Now()
	var traceError error

	query := "INSERT INTO file (checksum, name, file_path, created_by) VALUES(?, ?, ?, ?)"
	args := []interface{}{checksum, name, filePath, user}

	var id int64
	id, traceError = tx.RunMustExec(query, args...).LastInsertId()

	i.SQLXDB.Trace(query, start, traceError)
	return id, traceError
}

// Fetch implements the db.FileRepository.Fetch method
func (i *FileSqlxRepository) Fetch(fileID int64) (db.File, error) {

	start := time.Now()
	var traceError error

	query := `SELECT file_id, checksum, name, file_path
			  FROM file
			  WHERE file_id = ?`
	args := []interface{}{fileID}

	file := db.File{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&file)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return file, err
}

// FetchByChecksum implements the db.FileRepository.FetchByChecksum method
func (i *FileSqlxRepository) FetchByChecksum(checksum string) (db.File, error) {

	start := time.Now()
	var traceError error

	query := `SELECT file_id, checksum, name, file_path
			  FROM file
			  WHERE checksum = ?`
	args := []interface{}{checksum}

	file := db.File{}

	err := i.SQLXDB.DB.QueryRowx(query, args...).StructScan(&file)
	if err != nil && err != sql.ErrNoRows {
		traceError = err
	}

	i.SQLXDB.Trace(query, start, traceError)
	return file, err
}
