package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-model/go/datasource"
)

// Query interface
func (ds *AwsAthenaDatasource) Query(ctx context.Context, req *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	ds.logger.Debug("Query Req : %v", req)
	ds.logger.Debug("Context ctx : %v", ctx)

	opts, err := ds.parseDatasourceRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	ds.logger.Debug("opts  : ", opts)

	results, err := ds.handleAthenaQuery(ctx, opts)
	if err != nil {
		return nil, err
	}
	ds.logger.Debug("results  : ", results)

	parsedResults := make([]*datasource.QueryResult, 0)
	for _, result := range results {
		parsed, err := ds.parseResult(result)
		if err != nil {
			return nil, err
		}
		parsedResults = append(parsedResults, parsed)
	}
	res := &datasource.DatasourceResponse{
		Results: parsedResults,
	}
	ds.logger.Debug("parsed results  : ", res)
	return res, nil
}

func (ds *AwsAthenaDatasource) parseDatasourceRequest(ctx context.Context, req *datasource.DatasourceRequest) ([]*AthenaDatasourceQueryOption, error) {
	ds.logger.Debug("parseDataSourceRequest!")

	fromEpochMs := req.TimeRange.GetFromEpochMs()
	toEpochMs := req.TimeRange.GetToEpochMs()
	from := time.Unix(fromEpochMs/1000, fromEpochMs%1000*1000*1000)
	to := time.Unix(toEpochMs/1000, toEpochMs%1000*1000*1000)

	opts := make([]*AthenaDatasourceQueryOption, 0)
	for _, query := range req.Queries {
		opt := &AthenaDatasourceQueryOption{}
		opt.SecretKey = req.Datasource.GetDecryptedSecureJsonData()["secretAccessKey"]
		if err := json.Unmarshal([]byte(req.Datasource.GetJsonData()), &opt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(query.ModelJson), &opt); err != nil {
			return nil, err
		}
		opt.From = from
		opt.To = to
		opts = append(opts, opt)
	}
	return opts, nil
}

func (ds *AwsAthenaDatasource) handleAthenaQuery(ctx context.Context, queryOpts []*AthenaDatasourceQueryOption) ([]*AthenaQueryResult, error) {
	ds.logger.Debug("handleAthenaQuery!")
	results := make([]*AthenaQueryResult, 0)
	for _, opt := range queryOpts {
		res, err := ds.athena.HandleQuery(ctx, opt)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

func (ds *AwsAthenaDatasource) parseResult(result *AthenaQueryResult) (*datasource.QueryResult, error) {
	// gen metadata
	colInfos := make([]ColumnInfo, 0)
	for i := 0; i < len(result.ColumnInfoMap); i++ {
		colInfos = append(colInfos, *result.ColumnInfoMap[i])
	}
	metadata, err := json.Marshal(&QueryResultMetadata{
		ColumnInfos: colInfos,
	})
	if err != nil {
		return nil, err
	}

	parsedRes := &datasource.QueryResult{
		RefId:    result.Opt.RefID,
		MetaJson: string(metadata),
	}

	switch result.Opt.Format {
	case TimeSeries:
		series, err := ds.parseTimeSeries(result)
		if err != nil {
			return nil, err
		}
		parsedRes.Series = series
		return parsedRes, nil
	case Table:
		tables, err := ds.parseTable(result)
		if err != nil {
			return nil, err
		}
		parsedRes.Tables = tables
		return parsedRes, nil
	default:
		return nil, fmt.Errorf("Unexpected format type")
	}
}

func (ds *AwsAthenaDatasource) parseTimeSeries(result *AthenaQueryResult) ([]*datasource.TimeSeries, error) {
	opt := result.Opt
	seriesMap := make(map[string]*datasource.TimeSeries)
	valueColumns := make(map[string]bool)
	for _, valCol := range strings.Split(opt.ValueColumns, ",") {
		valCol = strings.Trim(valCol, " ")
		if valCol == "" {
			continue
		}
		valueColumns[valCol] = true
	}

	for _, row := range result.Rows {
		var t time.Time
		var err error
		var timestamp int64 = 0
		var metricVal string = ""
		tags := make(map[string]string)
		values := make(map[string]float64)

		for colIndex, rowValue := range row {
			colInfo := *result.ColumnInfoMap[colIndex]
			colName := colInfo.ColumnName

			switch colName {
			case opt.TimeColumn:
				t, err = time.Parse(TimestampLayout, rowValue)
				if err != nil {
					return nil, err
				}
				timestamp = t.Unix() * 1000
			case opt.MetricColumn:
				metricVal = rowValue
			default:
				if !(colInfo.Type == datasource.RowValue_TYPE_DOUBLE || colInfo.Type == datasource.RowValue_TYPE_INT64) {
					tags[colName] = rowValue
					continue
				}
				if _, ok := valueColumns[colName]; ok || len(valueColumns) == 0 {
					value, _ := strconv.ParseFloat(rowValue, 64)
					values[colName] = value
					continue
				}
			}
		}

		if !t.IsZero() && (t.Before(opt.From) || t.After(opt.To)) {
			continue
		}
		for colName, val := range values {
			seriesName := formatSeriesName(metricVal, colName)
			if seriesMap[seriesName] == nil {
				seriesMap[seriesName] = &datasource.TimeSeries{
					Name: seriesName,
					Tags: tags,
				}
			}
			seriesMap[seriesName].Points = append(seriesMap[seriesName].Points, &datasource.Point{
				Timestamp: timestamp,
				Value:     val,
			})
		}
	}

	series := make([]*datasource.TimeSeries, 0)
	for _, serie := range seriesMap {
		sort.Slice(serie.Points, func(i int, j int) bool {
			return serie.Points[i].Timestamp < serie.Points[j].Timestamp
		})
		series = append(series, serie)
	}
	return series, nil
}

func formatSeriesName(metricVal string, valueColName string) string {
	if metricVal == "" {
		return valueColName
	}
	return metricVal + " " + valueColName
}

func (ds *AwsAthenaDatasource) parseTable(result *AthenaQueryResult) ([]*datasource.Table, error) {
	table := datasource.Table{
		Columns: func(res *AthenaQueryResult) []*datasource.TableColumn {
			cols := make([]*datasource.TableColumn, 0)
			for i := 0; i < len(res.ColumnInfoMap); i++ {
				cols = append(cols, &datasource.TableColumn{
					Name: res.ColumnInfoMap[i].ColumnName,
				})
			}
			return cols
		}(result),
		Rows: make([]*datasource.TableRow, 0),
	}

	for _, row := range result.Rows {
		values := make([]*datasource.RowValue, 0)
		for colIndex, value := range row {
			info := *result.ColumnInfoMap[colIndex]
			var intVal int64 = 0
			var doubleVal float64 = 0.0

			//parse types
			switch info.Type {
			case datasource.RowValue_TYPE_INT64:
				// try parse as time
				t, err := time.Parse(TimestampLayout, value)
				if err == nil {
					intVal = t.Unix() * 1000 // to epoch millis
				} else {
					i, err := strconv.Atoi(value)
					if err == nil {
						intVal = int64(i)
					}
				}
			case datasource.RowValue_TYPE_DOUBLE:
				d, err := strconv.ParseFloat(value, 64)
				if err == nil {
					doubleVal = d
				}
			}
			values = append(values, &datasource.RowValue{
				Kind:        info.Type,
				Int64Value:  intVal,
				DoubleValue: doubleVal,
				StringValue: value,
			})
		}
		table.Rows = append(table.Rows, &datasource.TableRow{Values: values})
	}

	return []*datasource.Table{&table}, nil
}
