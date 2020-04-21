import { DataSourcePlugin } from '@grafana/data';
import { AthenaDataSource } from './DataSource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { AthenaDsQuery, AthenaDsOptions } from './types';

export const plugin = new DataSourcePlugin<AthenaDataSource, AthenaDsQuery, AthenaDsOptions>(AthenaDataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
