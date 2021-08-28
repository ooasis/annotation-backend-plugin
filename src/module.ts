import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { AnnoQuery, AnnoDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, AnnoQuery, AnnoDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
