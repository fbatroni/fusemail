package importer

import (
	"bitbucket.org/fusemail/em-lib-common-usage/common"
	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/fm-lib-commons-golang/trace"
)

//ColumnMapper subtype for YAML file
type ColumnMapper struct {
	TableIndex int `yaml:"tableIndex"`
	CsvIndex   int `yaml:"csvIndex"`
}

//VendorMapper type for YAML file
type VendorMapper struct {
	VendorID      int                     `yaml:"vendorID"`
	VendorName    string                  `yaml:"vendorName"`
	SqlBase       string                  `yaml:"sqlBase"`
	SqlValues     string                  `yaml:"sqlValues"`
	ColumnMappers map[string]ColumnMapper `yaml:"columnMappers,flow"`
}

// Importer is the interface for Importer Step
type ImporterStep interface {
	ExecuteImportStep(vendorMapper VendorMapper) error
}

// ImporterService implements the Importer Interface. Execute the Transform step based on the Vendor
type ImporterService struct {
	common.Options
	billingDB *db.BillingDB
	trace.Tracers
}
