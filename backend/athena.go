package main

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-model/go/datasource"
	hclog "github.com/hashicorp/go-hclog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/athena"
)

//IAwsAthenaQueryHandler ...
type IAwsAthenaQueryHandler interface {
	HandleQuery(ctx context.Context, opt *AthenaDatasourceQueryOption) (*AthenaQueryResult, error)
}

//QueryCacheInfo ...
type QueryCacheInfo struct {
	QueryName      string
	ExecResultID   string
	ExpirationTime time.Time
}

//AwsAthenaQueryHandler ...
type AwsAthenaQueryHandler struct {
	logger hclog.Logger
	// cache of NamedQueryID to cache Info
	cache map[string]*QueryCacheInfo
}

//HandleQuery handle athena query from grafana
func (handler *AwsAthenaQueryHandler) HandleQuery(ctx context.Context, opt *AthenaDatasourceQueryOption) (*AthenaQueryResult, error) {
	handler.logger.Debug("HandleQuery Query opt : ", opt)
	defer handler.cleanCache()
	sessionToken := ""
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}
	// TODO different creds provider?
	if opt.AccessKey != "" && opt.SecretKey != "" {
		cfg.Credentials = aws.NewStaticCredentialsProvider(opt.AccessKey, opt.SecretKey, sessionToken)
	}
	cfg.Region = opt.Region
	client := athena.New(cfg)

	switch opt.QueryType {
	case NamedQuery:
		return handler.handleNamedQuery(ctx, opt, client)
	case ExecutionQuery:
		return handler.handleExecutionQuery(ctx, opt, client)
	case GetNamedQueryMetrics:
		return handler.handleGetNamedQueryMetricsQuery(ctx, opt, client)
	default:
		return handler.handleTestQuery(ctx, opt, client)
	}
}

func (handler *AwsAthenaQueryHandler) isValidExecutionQuery(opt *AthenaDatasourceQueryOption) bool {
	return opt.ExecutionID != ""
}

func (handler *AwsAthenaQueryHandler) isValidNamedQuery(opt *AthenaDatasourceQueryOption) bool {
	return opt.NamedQuery != "" && opt.WorkGroup != ""
}

func (handler *AwsAthenaQueryHandler) handleGetNamedQueryMetricsQuery(ctx context.Context, opt *AthenaDatasourceQueryOption, athenaSvc *athena.Client) (*AthenaQueryResult, error) {
	handler.logger.Debug("handleGetNamedQueryMetricsQuery opt : ", opt)

	namedQueries, err := handler.getNamedQueries(ctx, opt, athenaSvc)
	if err != nil {
		return nil, err
	}
	// return as text, value
	result := &AthenaQueryResult{}
	result.Opt = opt
	result.ColumnInfoMap = make(map[int]*ColumnInfo)
	result.ColumnInfoMap[0] = &ColumnInfo{
		Type:       datasource.RowValue_TYPE_STRING,
		ColumnName: "text",
	}
	result.ColumnInfoMap[1] = &ColumnInfo{
		Type:       datasource.RowValue_TYPE_STRING,
		ColumnName: "value",
	}
	result.Rows = make([][]string, 0)
	for i := range namedQueries {
		result.Rows = append(result.Rows, make([]string, 2))
		result.Rows[i][0] = *namedQueries[i].Name
		result.Rows[i][1] = *namedQueries[i].Name
	}
	return result, nil
}

func (handler *AwsAthenaQueryHandler) handleTestQuery(ctx context.Context, opt *AthenaDatasourceQueryOption, athenaSvc *athena.Client) (*AthenaQueryResult, error) {
	handler.logger.Debug("handleTestQuery opt : ", opt)
	// try to call some athena request
	listNamedQueryReq := athenaSvc.ListNamedQueriesRequest(&athena.ListNamedQueriesInput{
		WorkGroup: &opt.WorkGroup,
	})
	_, err := listNamedQueryReq.Send(ctx)
	if err != nil {
		return nil, err
	}
	result := &AthenaQueryResult{}
	result.Opt = opt
	result.ColumnInfoMap = make(map[int]*ColumnInfo)
	result.Rows = make([][]string, 0)
	return result, nil
}

func (handler *AwsAthenaQueryHandler) handleExecutionQuery(ctx context.Context, opt *AthenaDatasourceQueryOption, athenaSvc *athena.Client) (*AthenaQueryResult, error) {
	handler.logger.Debug("handleExecutionQuery opt : ", opt)
	if !handler.isValidExecutionQuery(opt) {
		return nil, fmt.Errorf("Error. Invalid Execution Query")
	}
	return handler.retrieveExecResult(ctx, opt, &opt.ExecutionID, athenaSvc)
}

func (handler *AwsAthenaQueryHandler) handleNamedQuery(ctx context.Context, opt *AthenaDatasourceQueryOption, athenaSvc *athena.Client) (*AthenaQueryResult, error) {
	handler.logger.Debug("handleNamedQuery opt : ", opt)

	if !handler.isValidNamedQuery(opt) {
		return nil, fmt.Errorf("Error. Invalid Named Query")
	}
	namedQueries, err := handler.getNamedQueries(ctx, opt, athenaSvc)
	if err != nil {
		return nil, err
	}
	targetNamedQuery := find(&namedQueries, func(q athena.NamedQuery) bool {
		return *q.Name == opt.NamedQuery && *q.WorkGroup == opt.WorkGroup
	})
	if targetNamedQuery == nil {
		return nil, fmt.Errorf("Error. Named Query not found")
	}

	// use cache results if exist and not expired and useCache
	if cacheInfo, ok := handler.cache[*targetNamedQuery.NamedQueryId]; ok {
		handler.logger.Debug("Cache found...")
		if opt.UseCache && !cacheInfo.IsExpired() {
			handler.logger.Debug("Not expired, using cache...")
			return handler.retrieveExecResult(ctx, opt, &cacheInfo.ExecResultID, athenaSvc)
		}
		handler.logger.Debug("Cache Expired or explicitly skip cache, firing new request..")
	}

	// get work group info
	getWorkGrpReq := athenaSvc.GetWorkGroupRequest(&athena.GetWorkGroupInput{
		WorkGroup: targetNamedQuery.WorkGroup,
	})
	getWorkGrpRes, err := getWorkGrpReq.Send(ctx)
	if err != nil {
		return nil, err
	}
	handler.logger.Debug("res ", getWorkGrpRes)

	if getWorkGrpRes.WorkGroup.Configuration.ResultConfiguration.OutputLocation == nil {
		return nil, fmt.Errorf("Error. Please configure output location for workgroup %s", opt.WorkGroup)
	}

	return handler.execQuery(ctx, targetNamedQuery, getWorkGrpRes.WorkGroup, opt, athenaSvc)
}

func (handler *AwsAthenaQueryHandler) execQuery(ctx context.Context, targetNamedQuery *athena.NamedQuery, workGrp *athena.WorkGroup, opt *AthenaDatasourceQueryOption, athenaSvc *athena.Client) (*AthenaQueryResult, error) {
	handler.logger.Debug("Start execQuery..")
	// exec named query
	execNamedQueryReq := athenaSvc.StartQueryExecutionRequest(&athena.StartQueryExecutionInput{
		QueryString:         targetNamedQuery.QueryString,
		WorkGroup:           targetNamedQuery.WorkGroup,
		ResultConfiguration: workGrp.Configuration.ResultConfiguration,
	})
	execNamedQueryRes, err := execNamedQueryReq.Send(ctx)
	if err != nil {
		return nil, err
	}
	handler.logger.Debug("res ", execNamedQueryRes)
	// wait for result to be ready
	ch := make(chan athena.QueryExecutionState)
	go func(ch chan athena.QueryExecutionState) {
		state := athena.QueryExecutionStateFailed
		for i := 0; i < int(RequestTimeout/RequestInterval); i++ {
			handler.logger.Debug("Waiting...")
			time.Sleep(RequestInterval)
			getExecResultReq := athenaSvc.GetQueryExecutionRequest(&athena.GetQueryExecutionInput{
				QueryExecutionId: execNamedQueryRes.QueryExecutionId,
			})
			getExecResultRes, err := getExecResultReq.Send(ctx)
			if err != nil {
				break
			}
			if getExecResultRes.QueryExecution.Status.State == athena.QueryExecutionStateSucceeded {
				state = athena.QueryExecutionStateSucceeded
				break
			}
		}
		ch <- state
	}(ch)
	execState := <-ch
	handler.logger.Debug("execState ", execState)
	if execState != athena.QueryExecutionStateSucceeded {
		return nil, fmt.Errorf("Error executing request.. ExecState is %v", execState)
	}
	// cache execution ID
	handler.cache[*targetNamedQuery.NamedQueryId] = &QueryCacheInfo{
		QueryName:      *targetNamedQuery.Name,
		ExecResultID:   *execNamedQueryRes.QueryExecutionId,
		ExpirationTime: time.Now().Add(CacheExpiryTime),
	}

	return handler.retrieveExecResult(ctx, opt, execNamedQueryRes.QueryExecutionId, athenaSvc)
}

func (handler *AwsAthenaQueryHandler) retrieveExecResult(ctx context.Context, opt *AthenaDatasourceQueryOption, queryExecutionID *string, athenaSvc *athena.Client) (*AthenaQueryResult, error) {
	handler.logger.Debug("Start retrieveExecResult..")
	getQueryResultReq := athenaSvc.GetQueryResultsRequest(&athena.GetQueryResultsInput{
		QueryExecutionId: queryExecutionID,
	})

	getQueryResultRes, err := getQueryResultReq.Send(ctx)
	if err != nil {
		return nil, err
	}
	handler.logger.Debug("res ", getQueryResultRes)

	return handler.parseResultSet(opt, getQueryResultRes.ResultSet), nil
}

func (handler *AwsAthenaQueryHandler) parseResultSet(opt *AthenaDatasourceQueryOption, resultSet *athena.ResultSet) *AthenaQueryResult {
	result := &AthenaQueryResult{}
	result.Opt = opt
	result.ColumnInfoMap = make(map[int]*ColumnInfo)
	result.Rows = make([][]string, 0)

	// parse response
	for i, info := range resultSet.ResultSetMetadata.ColumnInfo {
		result.ColumnInfoMap[i] = &ColumnInfo{
			ColumnName: *info.Name,
			Type:       athenaToGrafanaType(*info.Type),
		}
	}
	if len(resultSet.Rows) > 1 {
		// first result row is header
		for i, row := range resultSet.Rows[1:] {
			result.Rows = append(result.Rows, make([]string, 0))
			for _, data := range row.Data {
				result.Rows[i] = append(result.Rows[i], *data.VarCharValue)
			}
		}
	}

	return result
}

// IsExpired ..
func (info *QueryCacheInfo) IsExpired() bool {
	return time.Now().After(info.ExpirationTime)
}

func (handler *AwsAthenaQueryHandler) cleanCache() {
	for k, v := range handler.cache {
		if v.IsExpired() {
			delete(handler.cache, k)
		}
	}
}

func (handler *AwsAthenaQueryHandler) getNamedQueries(ctx context.Context, opt *AthenaDatasourceQueryOption, athenaSvc *athena.Client) ([]athena.NamedQuery, error) {
	// get named Ids
	listNamedQueryReq := athenaSvc.ListNamedQueriesRequest(&athena.ListNamedQueriesInput{
		WorkGroup: &opt.WorkGroup,
	})
	listNamedQueryRes, err := listNamedQueryReq.Send(ctx)
	if err != nil {
		return nil, err
	}
	handler.logger.Debug("res ", listNamedQueryRes)

	getNamedQueryReq := athenaSvc.BatchGetNamedQueryRequest(&athena.BatchGetNamedQueryInput{
		NamedQueryIds: listNamedQueryRes.NamedQueryIds,
	})
	getNamedQueryRes, err := getNamedQueryReq.Send(ctx)
	if err != nil {
		return nil, err
	}
	return getNamedQueryRes.NamedQueries, nil
}

func find(namedQueries *[]athena.NamedQuery, fn func(athena.NamedQuery) bool) *athena.NamedQuery {
	for _, q := range *namedQueries {
		if fn(q) {
			return &q
		}
	}
	return nil
}

func athenaToGrafanaType(athenaType string) datasource.RowValue_Kind {
	switch athenaType {
	case "varchar":
		return datasource.RowValue_TYPE_STRING
	case "timestamp":
		return datasource.RowValue_TYPE_INT64
	case "bigint":
		return datasource.RowValue_TYPE_INT64
	default:
		return datasource.RowValue_TYPE_STRING
	}
}
