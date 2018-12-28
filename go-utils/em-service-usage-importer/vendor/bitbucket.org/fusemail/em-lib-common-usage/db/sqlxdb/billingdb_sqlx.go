//Package sqlxdb implements the db (package) interfaces
package sqlxdb

import (
	"database/sql"
	"errors"
	"time"

	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/fm-lib-commons-golang/deps"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//ErrNoActiveTransaction occurs when a transaction action is called but there is no active transaction.
var ErrNoActiveTransaction = errors.New("There is no active transaction")

//Options defines the configuration variables to create a SQLXDB entity
type Options struct {
	DBAddr string
	DBUser string
	DBPass string
	DBName string
}

//BillingSQLXDB is an implementation of BillingTransactionDB
type BillingSQLXDB struct {
	SQLXDB *deps.SQLXDB
}

//BillingSQLXTx is an implementation of BillingTx
type BillingSQLXTx struct {
	tx *sqlx.Tx
}

//New creates a new db.BillingDB based on SQLXDB implementation
func New(options Options) *db.BillingDB {

	conf := &mysql.Config{
		Net:                  "tcp",
		Addr:                 options.DBAddr,
		User:                 options.DBUser,
		Passwd:               options.DBPass,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	if options.DBName != "" {
		conf.DBName = options.DBName
	}

	sqlxDB := deps.NewSQLXDB(conf)

	db := &db.BillingDB{
		BillingTransactionDB:    &BillingSQLXDB{SQLXDB: sqlxDB},
		CollectionJobRepository: &CollectionJobSqlxRepository{SQLXDB: sqlxDB},
		StepRepository:          &StepSqlxRepository{SQLXDB: sqlxDB},
		StepTypeRepository:      &StepTypeSqlxRepository{SQLXDB: sqlxDB},
		FileRepository:          &FileSqlxRepository{SQLXDB: sqlxDB},
	}

	return db

}

// Check implements the Health interface from fusemail go common lib
func (i *BillingSQLXDB) Check() (map[string]interface{}, error) {
	return i.SQLXDB.Check()
}

//NewTransaction is the implementation of NewTransaction from db.BillingTransationalDB
func (i *BillingSQLXDB) NewTransaction() (db.BillingTx, error) {

	tx, err := i.SQLXDB.DB.Beginx()
	btx := &BillingSQLXTx{tx: tx}
	return btx, err
}

/*
	BulkInsert insert all given entities in one single insert statment.
*/
func (i *BillingSQLXDB) BulkInsert(sqlBase string, sqlValues string, entities [][]interface{}) error {

	//Values to be inserted
	valsInsert := []interface{}{}

	/*
		Creates the string for all group of values (?,?...?),(?,?...?)...(?,?...?) and
		the slice containing all values to replace the ? in the final sql statment
	*/
	var bulkStmt string
	for index := 0; index < len(entities); index++ {

		bulkStmt = bulkStmt + sqlValues
		if index < len(entities)-1 {
			bulkStmt = bulkStmt + ", "
		}

		valsInsert = append(valsInsert, entities[index]...)
	}

	start := time.Now()
	var traceError error

	// Creates the INSERT statment: INSERT INTO table (Column1, Column2...ColumnN) VALUES (?,?...?),(?,?...?)...(?,?...?)
	sqlStmt := sqlBase + " " + bulkStmt

	stmt, traceError := i.SQLXDB.DB.Prepare(sqlStmt)
	if traceError == nil {
		_, execError := stmt.Exec(valsInsert...)
		if execError != nil {
			traceError = execError
		}
		closeError := stmt.Close()
		if closeError != nil {
			traceError = closeError
		}
	}

	i.SQLXDB.Trace(sqlStmt, start, traceError)

	return traceError

}

//Connect to the database
func (i *BillingSQLXDB) Connect() error {
	return i.SQLXDB.Connect()
}

//RunSelect is the implementation of BillingTx interface
func (i *BillingSQLXTx) RunSelect(dest interface{}, query string, args ...interface{}) error {
	if i.tx == nil {
		return ErrNoActiveTransaction
	}
	return i.tx.Select(dest, query, args)
}

//RunQueryRow is the implementation of BillingTx interface
func (i *BillingSQLXTx) RunQueryRow(dest interface{}, query string, args ...interface{}) error {
	if i.tx == nil {
		return ErrNoActiveTransaction
	}
	return i.tx.QueryRowx(query, args...).StructScan(dest)
}

//RunExec is the implementation of BillingTx interface
func (i *BillingSQLXTx) RunExec(query string, args ...interface{}) (sql.Result, error) {
	if i.tx == nil {
		return nil, ErrNoActiveTransaction
	}
	return i.tx.Exec(query, args...)
}

//RunMustExec is the implementation of BillingTx interface
func (i *BillingSQLXTx) RunMustExec(query string, args ...interface{}) sql.Result {
	return i.tx.MustExec(query, args...)
}

//CommitTransaction is the implementation of BillingTx interface
func (i *BillingSQLXTx) CommitTransaction() error {
	if i.tx == nil {
		return ErrNoActiveTransaction
	}
	return i.tx.Commit()
}

//RollbackTransaction is the implementation of BillingTx interface
func (i *BillingSQLXTx) RollbackTransaction() error {
	if i.tx == nil {
		return ErrNoActiveTransaction
	}
	return i.tx.Rollback()
}
