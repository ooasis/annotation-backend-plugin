import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { AnnoDataSourceOptions, AnnoQuery } from './types';

export class DataSource extends DataSourceWithBackend<AnnoQuery, AnnoDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<AnnoDataSourceOptions>) {
    super(instanceSettings);
  }
}
