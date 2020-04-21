import { DataQuery, DataSourceJsonData } from '@grafana/data';
export enum QueryType {
  NamedQuery = 'NamedQuery',
  ExecutionQuery = 'ExecutionQuery',
  GetNamedQueryMetrics = 'GetNamedQueryMetrics',
  TestQuery = '',
}

export enum FormatType {
  TimeSeries = 'timeseries',
  Table = 'table',
}

export interface AthenaDsQuery extends DataQuery {
  namedQuery?: string;
  queryType?: QueryType;
  timeColumn?: string;
  metricColumn?: string;
  valueColumns?: string;
  executionId?: string;
  format?: FormatType;
  useCache?: boolean;
}

export const defaultQuery: Partial<AthenaDsQuery> = {
  namedQuery: '',
  queryType: QueryType.TestQuery,
  timeColumn: 'time',
  metricColumn: 'metric',
  valueColumns: '',
  executionId: '',
  format: FormatType.TimeSeries,
  useCache: true,
};

/**
 * These are options configured for each DataSource instance
 */
export interface AthenaDsOptions extends DataSourceJsonData {
  accessKey: string;
  region: string;
  workGroup: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface AthenaDsSecureJsonData {
  secretAccessKey: string;
}

export interface ColumnInfo {
  colName: string;
  colType: RowValueType;
}
export interface CustomMetadata {
  colInfos: ColumnInfo[];
}

export enum RowValueType {
  NULL = 0,
  DOUBLE = 1,
  INT = 2,
  BOOL = 3,
  STRING = 4,
  BYTES = 5,
}
