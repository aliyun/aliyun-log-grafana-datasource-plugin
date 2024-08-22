import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface SLSQuery extends DataQuery {
  query?: string;
  xcol?: string;
  ycol?: string;
  logsPerPage?: number;
  currentPage?: number;
  totalResults?: number;
}

export const defaultQuery: Partial<SLSQuery> = {
  query: '* | select count(*) as c, __time__-__time__%60 as t group by t',
  xcol: 't',
  ycol: '',
  logsPerPage: 100,
  currentPage: 1,
  totalResults: 100,
};

/**
 * These are options configured for each DataSource instance
 */
export interface SLSDataSourceOptions extends DataSourceJsonData {
  endpoint?: string;
  project?: string;
  logstore?: string;
  roleArn?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SLSSecureJsonData {
  accessKeyId?: string;
  accessKeySecret?: string;
}

// export declare type SlsLog = {
//   time: number;
//   attribute: Map<string, string>;
//   name: string;
// };
