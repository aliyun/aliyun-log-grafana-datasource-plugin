import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface SLSQuery extends DataQuery {
  query?: string;
  xcol?: string;
  ycol?: string;
  logsPerPage?: number;
  currentPage?: number;
  type?: string;
  logstore?: string
  legendFormat?: string // 图例 format
  step?: string // promql step
  queryType?: 'range' | 'instant'; // 查询类型
  totalLogs?: number
  powerSql?: boolean
}

export const defaultQuery: Partial<SLSQuery> = {
  type: 'all',
  query: '* | select count(*) as c, __time__-__time__%60 as t group by t',
  xcol: 't',
  ycol: '',
  logsPerPage: 100,
  currentPage: 1,
  totalLogs: 100,
  powerSql: false
};
export const defaultEidtorQuery: Partial<SLSQuery> = {
  type: 'all',
  query: '',
  xcol: '',
  ycol: '',
  logsPerPage: 100,
  currentPage: 1,
  totalLogs: 100,
  powerSql: false
};

/**
 * These are options configured for each DataSource instance
 */
export interface SLSDataSourceOptions extends DataSourceJsonData {
  endpoint?: string;
  project?: string;
  logstore?: string;
  roleArn?: string;
  region?: string;
  international?: boolean;
  headers?: HeaderWithValue[]
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

export type Header = {
  name: string;
  configured: boolean;
};

export type HeaderWithValue = Header & { value: string };

export type LocalHeader = HeaderWithValue & { id: string };
