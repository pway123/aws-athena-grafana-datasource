import defaults from 'lodash/defaults';

import _ from 'lodash';
import { getBackendSrv, BackendSrv } from '@grafana/runtime';
import {
  DataQueryRequest,
  DataQueryResponse,
  DataSourceApi,
  DataSourceInstanceSettings,
  MutableDataFrame,
  MutableField,
  FieldType,
} from '@grafana/data';

import { AthenaDsQuery, AthenaDsOptions, defaultQuery, CustomMetadata, RowValueType, QueryType, FormatType } from './types';

const BACKEND_URL = '/api/tsdb/query';

export class AthenaDataSource extends DataSourceApi<AthenaDsQuery, AthenaDsOptions> {
  private backendSrv: BackendSrv;
  private headers: any;

  constructor(instanceSettings: DataSourceInstanceSettings<AthenaDsOptions>) {
    super(instanceSettings);
    this.backendSrv = getBackendSrv();
    this.headers = {
      'Content-Type': 'application/json',
    };
  }

  async query(options: DataQueryRequest<AthenaDsQuery>): Promise<DataQueryResponse> {
    const { range, targets } = options;
    const from = range!.from.valueOf().toString();
    const to = range!.to.valueOf().toString();

    const queries = targets.filter(this.isValidQuery);
    const request = {
      url: BACKEND_URL,
      method: 'POST',
      headers: this.headers,
      data: {
        from,
        to,
        queries: queries.map(target => {
          const t = defaults(target, defaultQuery);
          return {
            datasourceId: this.id,
            ...t,
          };
        }),
      },
    };
    return this.backendSrv
      .datasourceRequest(request)
      .then((res: any) => {
        const data: any[] = [];
        for (const query of queries) {
          const result = res.data.results[query.refId];
          if (!_.isEmpty(result.series)) {
            // time series
            _.forEach(result.series, serie => {
              data.push({ target: serie.name, datapoints: serie.points });
            });
          }
          if (!_.isEmpty(result.tables)) {
            // table
            const meta: CustomMetadata = result.meta;
            _.forEach(result.tables, table => {
              const tableData = new MutableDataFrame({
                refId: query.refId,
                fields: table.columns?.map(
                  (col: any, colIndex: number): MutableField => {
                    return {
                      name: col.text,
                      values: table.rows.map((row: any[]) => row[colIndex]),
                      type: determineFieldType(col.text, meta.colInfos.find(info => info.colName === col.text)?.colType, query.timeColumn),
                      config: {},
                    };
                  }
                ),
              });
              data.push(tableData);
            });
          }
        }
        res.data = data;
        return res;
      })
      .catch((err: any = { data: { error: '' } }) => {
        console.log('Err: ', err);
        if (err.data && err.data.message === 'Metric request error' && err.data.error) {
          err.data.message = err.data.error;
        }

        throw err;
      });
  }

  async testDatasource() {
    // call backend with queryType '' for healthcheck
    try {
      const res = await this.backendSrv.datasourceRequest({
        url: BACKEND_URL,
        method: 'POST',
        headers: this.headers,
        data: {
          queries: [
            {
              datasourceId: this.id,
              ...defaultQuery,
            },
          ],
        },
      });
      const isSuccess = res.status === 200;
      return {
        status: isSuccess ? 'success' : 'failed',
        message: isSuccess ? 'Success' : 'Failed',
      };
    } catch {
      return {
        status: 'error',
        message: 'Error',
      };
    }
  }

  async getNamedQueries() {
    return await this.doMetricQueryRequest(QueryType.GetNamedQueryMetrics);
  }

  private isValidQuery(q: AthenaDsQuery): boolean {
    // TODO more validation
    return q.hide !== true;
  }

  private transformSuggestDataFromTable(suggestData: any) {
    return _.map(suggestData.results['metricFindQuery'].tables[0].rows, v => {
      return {
        text: v[0],
        value: v[1],
        label: v[1],
      };
    });
  }

  private async doMetricQueryRequest(queryType: string, params?: Partial<AthenaDsQuery>) {
    const res = await this.backendSrv.datasourceRequest({
      url: BACKEND_URL,
      method: 'POST',
      headers: this.headers,
      data: {
        queries: [
          _.extend(
            {
              refId: 'metricFindQuery',
              datasourceId: this.id,
              queryType,
              format: FormatType.Table,
            },
            params
          ),
        ],
      },
    });
    return this.transformSuggestDataFromTable(res.data);
  }
}

function determineFieldType(fieldName: string, rowValueType?: RowValueType, timeColumn?: string): FieldType {
  if (fieldName === timeColumn) {
    return FieldType.time;
  }
  switch (rowValueType) {
    case RowValueType.BOOL:
      return FieldType.boolean;
    case RowValueType.INT:
    case RowValueType.DOUBLE:
      return FieldType.number;
    default:
      return FieldType.string;
  }
}
