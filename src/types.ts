import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface AnnoQuery extends DataQuery {
  tags?: string;
}

export const defaultQuery: Partial<AnnoQuery> = {
  tags: '',
};

/**
 * These are options configured for each DataSource instance.
 */
export interface AnnoDataSourceOptions extends DataSourceJsonData {
  serverUrl?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiKey?: string;
}
