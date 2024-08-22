import {
  TraceLog,
  TraceKeyValuePair,
  TraceSpanRow,
  DataFrame,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  FieldType,
  MutableDataFrame,
  Vector,
} from '@grafana/data';
import { SLSDataSourceOptions, SLSQuery } from './types';
import { DataSourceWithBackend, getBackendSrv, getTemplateSrv } from '@grafana/runtime';
import _ from 'lodash';
import { map } from 'rxjs/operators';

export class SLSDataSource extends DataSourceWithBackend<SLSQuery, SLSDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SLSDataSourceOptions>) {
    super(instanceSettings);
  }

  query(options: DataQueryRequest<SLSQuery>) {
    // console.log("into query.")
    options.targets.forEach((q: SLSQuery) => {
      q.query = replaceQueryParameters(q, options);
    });
    if (options.targets[0].xcol === 'trace') {
      return super.query(options).pipe(map(responseToDataQueryResponse));
    }
    return super.query(options);
  }

  metricFindQuery(query: SLSQuery|string, options?: any) {
    // console.log("into metricFindQuery")
    const data = {
      from: options.range.from.valueOf().toString(),
      to: options.range.to.valueOf().toString(),
      queries: [
        {
          // datasource: this.name,
          datasource: {type: this.type, uid: this.uid},
          datasourceId: this.id,
          query: replaceQueryParameters(query, options),
        },
      ],
    };
    return getBackendSrv()
      .post('/api/ds/query', data)
      .then((response) => {
        // console.log("metricFindQuery /api/ds/query response", response)
        return response.results.A.frames[0].data.values;
      })
      .then(mapToTextValue);
  }
}

function valueToTag(key: any, value: any): TraceKeyValuePair {
  return {
    key,
    value,
  };
}

function transServiceTags(result: Map<string, Vector>, i: number): TraceKeyValuePair[] {
  const resource = JSON.parse(result.get('resource')?.get(i));
  const resourceArray = Array.from(Object.entries(resource), ([name, value]) => valueToTag(name, value));
  resourceArray.push(valueToTag('host', result.get('host')?.get(i)));
  return resourceArray;
}

function transTags(result: Map<string, Vector>, i: number): TraceKeyValuePair[] {
  const attribute = JSON.parse(result.get('attribute')?.get(i));
  const resourceArray = Array.from(Object.entries(attribute), ([name, value]) => valueToTag(name, value));
  resourceArray.push(valueToTag('statusCode', result.get('statusCode')?.get(i)));
  resourceArray.push(valueToTag('statusMessage', result.get('statusMessage')?.get(i)));
  return resourceArray;
}

function transLogs(result: Map<string, Vector>, i: number): TraceLog[] {
  let traceLogs: TraceLog[] = [];
  const slsLogs = JSON.parse(result.get('logs')?.get(i)) as Object[];
  for (const slsLog of slsLogs) {
    const logMap = new Map(Object.entries(slsLog));
    const attributeArray = Array.from(Object.entries(slsLog), ([name, value]) => valueToTag(name, value));
    let time = logMap.get('time');
    if (time !== undefined) {
      traceLogs.push({
        timestamp: Number(time) / 1000000,
        fields: attributeArray,
      });
    }
  }
  return traceLogs;
}

function transformSpan(df: DataFrame): TraceSpanRow[] {
  let traceSpanRows: TraceSpanRow[] = [];
  const fields = df.fields;
  const result = new Map(fields.map((key) => [key.name, key.values]));
  for (let i = 0; i < df.length; i++) {
    const tsd = {
      traceID: result.get('traceID')?.get(i),
      spanID: result.get('spanID')?.get(i),
      parentSpanID: result.get('parentSpanID')?.get(i),
      operationName: result.get('operationName')?.get(i),
      serviceName: result.get('serviceName')?.get(i),
      serviceTags: transServiceTags(result, i),
      startTime: result.get('startTime')?.get(i),
      duration: result.get('duration')?.get(i),
      tags: transTags(result, i),
      errorIconColor: result.get('statusCode')?.get(i) === 'ERROR' ? '#f00' : '',
      logs: transLogs(result, i),
    };
    traceSpanRows.push(tsd);
  }
  return traceSpanRows;
}

export function transformResponse(df: DataFrame): DataFrame {
  const spanRows = transformSpan(df);
  const frame = new MutableDataFrame({
    fields: [
      { name: 'traceID', type: FieldType.string },
      { name: 'spanID', type: FieldType.string },
      { name: 'parentSpanID', type: FieldType.string },
      { name: 'operationName', type: FieldType.string },
      { name: 'serviceName', type: FieldType.string },
      { name: 'serviceTags', type: FieldType.other },
      { name: 'startTime', type: FieldType.number },
      { name: 'duration', type: FieldType.number },
      { name: 'logs', type: FieldType.other },
      { name: 'tags', type: FieldType.other },
      { name: 'errorIconColor', type: FieldType.string },
    ],
    meta: {
      preferredVisualisationType: 'trace',
    },
  });

  for (const span of spanRows) {
    frame.add(span);
  }

  return frame;
}

function responseToDataQueryResponse(response: DataQueryResponse): DataQueryResponse {
  console.log(response);
  return {
    data: [transformResponse(response.data[0])],
  };
}

export function replaceFormat(
    value: { forEach: (arg0: (v: string) => void) => void; join: (arg0: string) => void },
    variable: { multi: any; includeAll: any; name: string; label: any; description: string }
) {
  if (typeof value === 'object' && (variable.multi || variable.includeAll)) {
    const a: string[] = [];
    value.forEach(function (v: string) {
      if (variable.name === variable.label || (variable.description && variable.description.indexOf('field_search') >= 0)) {
        a.push('"' + variable.name + '":"' + v + '"');
      } else {
        a.push(v);
      }
    });
    return a.join(' OR ');
  }
  if (_.isArray(value)) {
    return value.join(' OR ');
  }
  return value;
}

export function mapToTextValue(result: any) {
  if (Array.isArray(result) && result.length === 2) {
    return _.map(result[0], (d, i) => {
      return { text: d, value: result[1][i] };
    });
  }
  return _.map(result[0], (d, i) => {
    if (d && d.text && d.value) {
      return { text: d.text, value: d.value };
    } else if (_.isObject(d)) {
      return { text: d, value: i };
    }
    return { text: d, value: d };
  });
}

export function replaceQueryParameters(q: SLSQuery|string, options: DataQueryRequest<SLSQuery>) {
  if (typeof q !== "string" && q.hide) {
    return;
  }
  let varQuery;
  if (typeof q === "string") {
    varQuery = q
  }else {
    varQuery = q.query
  }
  let query = getTemplateSrv().replace(
      varQuery,
    options.scopedVars,
      replaceFormat
  );

  const re = /\$([0-9]+)([dmhs])/g;
  const reArray = query.match(re);
  _(reArray).forEach(function (col) {
    const old = col;
    col = col.replace('$', '');
    let sec = 1;
    if (col.indexOf('s') !== -1) {
      sec = 1;
    } else if (col.indexOf('m') !== -1) {
      sec = 60;
    } else if (col.indexOf('h') !== -1) {
      sec = 3600;
    } else if (col.indexOf('d') !== -1) {
      sec = 3600 * 24;
    }
    col = col.replace(/[smhd]/g, '');
    let v = parseInt(col, 10);
    v = v * sec;
    console.log(old, v, col, sec, query);
    query = query.replace(old, String(v));
  });
  if (query.indexOf('#time_end') !== -1) {
    query = query.replace('#time_end', String(options.range.to.unix() / 1000));
  }
  if (query.indexOf('#time_begin') !== -1) {
    query = query.replace('#time_begin', String(options.range.from.unix() / 1000));
  }
  return query;
}
