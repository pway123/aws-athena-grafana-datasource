import defaults from 'lodash/defaults';

import React, { PureComponent, ChangeEvent } from 'react';
import { FormField, Button, Select, FormLabel, Input } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { AthenaDataSource } from './DataSource';
import { AthenaDsQuery, AthenaDsOptions, defaultQuery, QueryType, FormatType } from './types';
import { getDataSourceSrv } from '@grafana/runtime';

type Props = QueryEditorProps<AthenaDataSource, AthenaDsQuery, AthenaDsOptions>;

const queryTypes = [
  { label: 'Select', value: QueryType.TestQuery },
  { label: 'Exec Named Query', value: QueryType.NamedQuery },
  { label: 'Fetch Exec Results', value: QueryType.ExecutionQuery },
];

const formatTypes = [
  { label: 'Time Series', value: FormatType.TimeSeries },
  { label: 'Table', value: FormatType.Table },
];

interface QueryEditorState {
  namedQueries: SelectableValue[];
  selectedQueryType: SelectableValue;
  selectedNameQuery: SelectableValue;
  selectedFormatType: SelectableValue;
}

const FIELD_WIDTH = 15;

export class QueryEditor extends PureComponent<Props, QueryEditorState> {
  constructor(props: Props) {
    super(props);
    this.state = {
      selectedQueryType: queryTypes.find(q => q.value === props.query.queryType) || queryTypes[0],
      namedQueries: [],
      selectedNameQuery: { label: '', value: '' } as SelectableValue,
      selectedFormatType: formatTypes.find(f => f.value === props.query.format) || formatTypes[0],
    };
  }

  componentDidMount() {
    this.loadNamedQueries();
  }

  async loadNamedQueries() {
    const ds = await getDataSourceSrv().get('aws-athena-datasource-plugin');
    const res = await (ds as AthenaDataSource).getNamedQueries();
    this.setState({
      namedQueries: res as SelectableValue[],
    });
  }

  onChangeHof = (fieldName: string, isNumeric = false, shouldRunQuery = false) => {
    return (event: ChangeEvent<HTMLInputElement>) => {
      const { onChange, query } = this.props;
      const value = isNumeric ? parseFloat(event.target.value) : event.target.value;
      onChange({ ...query, [fieldName]: value });
      if (shouldRunQuery) {
        this.runQuery();
      }
    };
  };

  onChangeHofCheckbox = (fieldName: string, shouldRunQuery = false) => {
    return (event: ChangeEvent<HTMLInputElement>) => {
      const { onChange, query } = this.props;
      const value = event.target.checked;
      onChange({ ...query, [fieldName]: value });
      if (shouldRunQuery) {
        this.runQuery();
      }
    };
  };

  onClickRunQuery = () => {
    this.runQuery();
  };

  isValidQuery = (): boolean => {
    // TODO validate query
    return true;
  };

  runQuery = () => {
    const { onRunQuery } = this.props;
    if (!this.isValidQuery()) {
      return;
    }
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { useCache, timeColumn, valueColumns, metricColumn, executionId } = query;

    return (
      <div className="gf-form-group">
        <div className="gf-form-inline">
          <FormLabel width={FIELD_WIDTH}>Query Type</FormLabel>
          <Select
            width={FIELD_WIDTH}
            options={queryTypes}
            value={this.state.selectedQueryType}
            onChange={v => {
              this.props.onChange({ ...query, queryType: v.value });
              this.setState({ selectedQueryType: v });
            }}
          />
        </div>
        {this.state.selectedQueryType.value === QueryType.NamedQuery && (
          <div className="gf-form-inline">
            <FormLabel width={FIELD_WIDTH}>Named Queries</FormLabel>
            <Select
              width={FIELD_WIDTH}
              options={this.state.namedQueries}
              value={this.state.selectedNameQuery}
              onChange={v => {
                this.props.onChange({ ...query, namedQuery: v.value });
                this.setState({ selectedNameQuery: v });
              }}
            />
          </div>
        )}
        {this.state.selectedQueryType.value === QueryType.ExecutionQuery && (
          <div className="gf-form">
            <FormField
              labelWidth={FIELD_WIDTH}
              value={executionId || ''}
              onChange={this.onChangeHof('executionId')}
              label="Execution ID"
              tooltip="Aws athena named query to execute"
            ></FormField>
          </div>
        )}
        {this.state.selectedFormatType.value === FormatType.TimeSeries && (
          <div className="gf-form">
            <FormField
              labelWidth={FIELD_WIDTH}
              value={metricColumn || ''}
              onChange={this.onChangeHof('metricColumn')}
              label="Metric Column"
              tooltip="Aws athena named query to execute"
            ></FormField>
          </div>
        )}
        {this.state.selectedFormatType.value === FormatType.TimeSeries && (
          <div className="gf-form">
            <FormField
              labelWidth={FIELD_WIDTH}
              value={timeColumn || ''}
              onChange={this.onChangeHof('timeColumn')}
              label="Time Column"
              tooltip="Aws athena named query to execute"
            ></FormField>
          </div>
        )}
        {this.state.selectedFormatType.value === FormatType.TimeSeries && (
          <div className="gf-form">
            <FormField
              labelWidth={FIELD_WIDTH}
              value={valueColumns || ''}
              onChange={this.onChangeHof('valueColumns')}
              label="Value Columns"
              tooltip="Comma separated column names for values. Default all numerical columns will be treated as value columns"
            ></FormField>
          </div>
        )}

        <div className="gf-form-inline">
          <FormLabel width={FIELD_WIDTH}>Format</FormLabel>
          <Select
            width={FIELD_WIDTH}
            options={formatTypes}
            value={this.state.selectedFormatType}
            onChange={v => {
              this.props.onChange({ ...query, format: v.value });
              this.setState({ selectedFormatType: v });
            }}
          />
        </div>
        <div className="gf-form-inline">
          <FormLabel width={FIELD_WIDTH}>Use Cache</FormLabel>
          <Input type="checkbox" checked={useCache} onChange={this.onChangeHofCheckbox('useCache')} />
        </div>
        <div className="gf-form">
          <Button variant="primary" onClick={this.onClickRunQuery}>
            Run Query
          </Button>
        </div>
      </div>
    );
  }
}
