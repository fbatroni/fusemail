package db

import (
	"database/sql"
	"time"
)

// -----COLLECTION JOB-----

/*
CollectionJob is the complete Job to collect usage data from one vendor source
and insert it to the billing database. The Collection Job is composed of Steps (individual tasks).
The Steps are: Download data, Transform data (might be multiple transformations),
import data to the database and archive the files involved in these steps.

The CollectionJob struct represents the Collection Job information at billing database.

Database: billing
Table: collection_job
Status ENUM Values: 'IN PROGRESS', 'FINISHED', 'ERROR'
*/
type CollectionJob struct {
	ID                int64     `db:"collection_job_id" json:"collection_job_id"`
	SourceID          int64     `db:"source_id" json:"source_id"`
	Status            string    `db:"status" json:"status"`
	CurrentStepTypeID int64     `db:"current_step_type_id" json:"current_step_type_id"`
	StarteDate        time.Time `db:"start_date" json:"start_date"`
	EndDate           time.Time `db:"end_date" json:"end_date"`
}

//CollectionJobStatus represents status values for a Collection Job
type CollectionJobStatus string

//ToString returns the string value of the CollectionJobStatus type
func (i CollectionJobStatus) ToString() string {
	return string(i)
}

const (

	// CollectionJobStatusInProgress represents the In Progress status (initial status) of a CollectionJob
	CollectionJobStatusInProgress CollectionJobStatus = "IN PROGRESS"
	// CollectionJobStatusFinished represents the Finished status (ending status) of a CollectionJob
	CollectionJobStatusFinished CollectionJobStatus = "FINISHED"
	// CollectionJobStatusError represents the Error status (error ending status) of a CollectionJob
	CollectionJobStatusError CollectionJobStatus = "ERROR"
)

// -----STEP-----

/*
Step is a independent process taken in the Usage Data Collection Job. There is 4 types of steps:
- Download: to acquire the data
- Transform (it might have sub types such as Summarization, Translation):  to process the data
- Importer: to import the data into the database
- Archive: to archive files involved in the Usage Data Collection Job

The Step struct represents the step information at billing database.

Database: billing
Table: step
Status ENUM Values: 'IN PROGRESS', 'FINISHED', 'ERROR'
*/
type Step struct {
	ID              int64         `db:"step_id" json:"step_id"`
	StepTypeID      int64         `db:"step_type_id" json:"step_type_id"`
	CollectionJobID int64         `db:"collection_job_id" json:"collection_job_id"`
	StarteDate      time.Time     `db:"start_date" json:"start_date"`
	EndDate         time.Time     `db:"end_date" json:"end_date"`
	FileID          sql.NullInt64 `db:"file_id" json:"file_id"`
	Status          string        `db:"status" json:"status"`
	Error           string        `db:"error" json:"error"`
}

//StepStatus represents status values for a Step
type StepStatus string

//ToString returns the string value of the StepStatus Type
func (i StepStatus) ToString() string {
	return string(i)
}

const (
	// StepStatusInProgress represents the In Progress status (initial status) of a Step
	StepStatusInProgress StepStatus = "IN PROGRESS"
	// StepStatusFinished represents the Finished status (ending status) of a Step
	StepStatusFinished StepStatus = "FINISHED"
	// StepStatusError represents the Error status (error ending status) of a Step
	StepStatusError StepStatus = "ERROR"
)

// -----STEP TYPE-----

/*
StepType represents the type of an Collection job Step. It defines the Name of the Step, the source
that onws this step type and the order of this step type in the Collection Job complete process.

The StepType struct represents the StepType information at billing database.

Database: billing
Table: step
Name ENUM Values: 'DOWNLOAD', 'SUMMARIZE', 'TRANSLATE', 'IMPORT', 'ARCHIVE'
*/
type StepType struct {
	ID        int64  `db:"step_type_id" json:"step_type_id"`
	SourceID  int64  `db:"source_id" json:"source_id"`
	Name      string `db:"name" json:"name"`
	StepOrder int    `db:"step_order" json:"step_order"`
}

//StepTypeName represents name values for a StepType
type StepTypeName string

//ToString returns the string value of the StepTypeName Type
func (i StepTypeName) ToString() string {
	return string(i)
}

const (
	// StepTypeDownload represents the In Progress status (initial status) of a Step
	StepTypeDownload StepTypeName = "DOWNLOAD"
	// StepTypeSummarize represents the Finished status (ending status) of a Step
	StepTypeSummarize StepTypeName = "SUMMARIZE"
	// StepTypeTranslate represents the Error status (error ending status) of a Step
	StepTypeTranslate StepTypeName = "TRANSLATE"
	// StepTypeImport represents the Error status (error ending status) of a Step
	StepTypeImport StepTypeName = "IMPORT"
	// StepTypeArchive represents the Error status (error ending status) of a Step
	StepTypeArchive StepTypeName = "ARCHIVE"
)

// -----FILE-----

/*
File represents a file that is either an output of a Step process or an input for a Step process
For example, the Download Step produces a file while an Import Step consumes a file.

The File struct represents the file information at the billing database.

Database: billing
Table: file
*/
type File struct {
	ID       int64  `db:"file_id" json:"file_id"`
	Checksum string `db:"checksum" json:"checksum"`
	Name     string `db:"name" json:"name"`
	FilePath string `db:"file_path" json:"file_path"`
}
