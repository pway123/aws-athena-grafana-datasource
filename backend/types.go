package main

import (
	"time"

	"github.com/grafana/grafana-plugin-model/go/datasource"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
)

// AwsAthenaDatasource plugin datasource
type AwsAthenaDatasource struct {
	plugin.NetRPCUnsupportedPlugin
	logger hclog.Logger
	athena IAwsAthenaQueryHandler
}

// AthenaDatasourceQueryOption mostly parsed from query request
type AthenaDatasourceQueryOption struct {
	RefID        string     `json:"refId"`
	QueryType    QueryType  `json:"queryType"`
	WorkGroup    string     `json:"workGroup"`
	TimeColumn   string     `json:"timeColumn"`
	NamedQuery   string     `json:"namedQuery"`
	ExecutionID  string     `json:"executionId"`
	MetricColumn string     `json:"metricColumn"`
	ValueColumns string     `json:"valueColumns"`
	UseCache     bool       `json:"useCache"`
	Format       FormatType `json:"format"`
	AuthType     AuthType   `json:"authType"`
	RoleARN      AuthType   `json:"roleArn"`
	Region       string     `json:"region"`
	AccessKey    string     `json:"accessKey"`
	SecretKey    string
	From         time.Time
	To           time.Time
}

//ColumnInfo ...
type ColumnInfo struct {
	Type       datasource.RowValue_Kind `json:"colType"`
	ColumnName string                   `json:"colName"`
}

//QueryResultMetadata ...
type QueryResultMetadata struct {
	ColumnInfos []ColumnInfo `json:"colInfos"`
}

// QueryType ...
type QueryType string

// AuthType ...
type AuthType string

// FormatType ...
type FormatType string

// AthenaQueryResult ...
type AthenaQueryResult struct {
	ColumnInfoMap map[int]*ColumnInfo
	Rows          [][]string
	Opt           *AthenaDatasourceQueryOption
}
