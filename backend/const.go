package main

import "time"

// strings
const (
	PluginName = "aws-athena-datasource-plugin"
)

// TimestampLayout of athena response
const TimestampLayout = "2006-01-02 15:04:05"

// Wait settings
const (
	RequestTimeout  = time.Duration(60) * time.Second
	RequestInterval = time.Duration(500) * time.Millisecond
)

//Cache settings
const (
	CacheExpiryTime = time.Duration(12) * time.Hour
)

// Format type
const (
	TimeSeries FormatType = "timeseries"
	Table      FormatType = "table"
)

// Query Type
const (
	NamedQuery           QueryType = "NamedQuery"
	ExecutionQuery       QueryType = "ExecutionQuery"
	GetNamedQueryMetrics QueryType = "GetNamedQueryMetrics"
)

// Auth Type
const (
	Static  AuthType = "Static"
	RoleArn AuthType = "RoleArn"
)
