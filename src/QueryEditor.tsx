import { defaults } from 'lodash';

import React, { ChangeEvent, PureComponent, FormEvent } from 'react';
import { InlineFormLabel, Icon, Input, Tooltip, InlineField, Button, ConfirmModal, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
// import { EditorField, EditorRow, QueryOptionGroup } from '@grafana/experimental';
import { InlineSwitch } from './components/Switch/switch';
import { SLSDataSource, replaceFormat } from './datasource';
import { defaultEidtorQuery, defaultQuery, SLSDataSourceOptions, SLSQuery } from './types';
import { dataSourceType, version, xSelectOptions } from './const';
import MonacoQueryField from './SLS-monaco-editor/MonacoQueryField';
import MonacoQueryFieldOld from 'SLS-monaco-editor/MonacoQueryFieldOld';
import { getDataSourceSrv, getTemplateSrv } from '@grafana/runtime';
import { Base64 } from 'js-base64';
import { SelectTips } from 'SelectTips';
import EditorRow from './components/QueryEditor/EditorRow';
import QueryOptionGroup from './components/QueryEditor/QueryOptionGroup';
import EditorField from './components/QueryEditor/EditorField';
import RadioButtonGroup from './components/RadioButtonGroup/RadioButtonGroup';

// const { FormField } = LegacyForms;

type Props = QueryEditorProps<SLSDataSource, SLSQuery, SLSDataSourceOptions> & { isVariable?: boolean };

const onSelectChange = (realXCol: string, onFind: () => void, onNotFind: () => void) => {
  if (xSelectOptions.find((e) => e.value === realXCol)) {
    onFind();
    return realXCol;
  }
  onNotFind();
  return 'custom';
};

export class SLSQueryEditor extends PureComponent<Props> {
  state = {
    loading: false,
    showAlert: false,
    xColDisabled:
      onSelectChange(
        this.props.query.xcol ?? 'custom',
        () => {},
        () => {}
      ) !== 'custom',
    url: '',
    message: '',
    logstoreList: [],
    logstore: '',
    powerSql: false,
    datasourceType: 'all'
  };

  componentDidMount() {
    const { query } = this.props;
    this.getList(query?.type || 'logstore');
  }

  autoSelect = (realXCol: string) => {
    return onSelectChange(
      realXCol,
      () => {
        this.setState({
          xColDiabled: true,
        });
      },
      () => {
        this.setState({
          xColDiabled: false,
        });
      }
    );
  };

  onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, query: event.target.value });
  };

  onQueryTextChangeString = (value: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, query: value });
  };

  onQueryTextChangeWithRunQuery = (value: string) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, query: value });
    // executes the query
    onRunQuery();
  };

  onXChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, xcol: event.target.value });
    this.autoSelect(event.target.value);
    // executes the query
    onRunQuery();
  };

  onYChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, ycol: event.target.value });
    // executes the query
    onRunQuery();
  };

  onDatasourceTypeChange = (v: SelectableValue) => {
    const { onChange, query } = this.props;
    onChange({ ...query, type: v.value, logstore: '' });
    this.setState({
      datasourceType: v.value,
    });

    this.getList(v.value);
  };

  onLogstoreChange = (v: SelectableValue) => {
    const { onChange, query } = this.props;
    let value = v.value;
    this.setState({
      logstore: value,
    });
    onChange({ ...query, logstore: value });
  };

  onPowerSqlChange = (v: SelectableValue) => {
    const { onChange, query } = this.props;
    const value = v.target.checked;
    this.setState({
      powerSql: value,
    });
    onChange({ ...query, powerSql: value });
  };

  onQueryTypeChange = (v: any) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, queryType: v });
    // executes the query
    onRunQuery?.();
  };

  onFormatChange = (event: FormEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;

    onChange({ ...query, legendFormat: event.currentTarget.value });
    // executes the query
    onRunQuery();
  };

  onStepChange = (event: FormEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, step: event.currentTarget.value });
    // executes the query
    onRunQuery();
  };

  onTotalLogsChange = (event: FormEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, totalLogs: Number(event.currentTarget.value) });
    // executes the query
    onRunQuery();
  };

  gotoSLS = () => {
    const { datasource, query, range } = this.props;
    const startTime = range?.from.unix() ?? Math.round(Date.now() / 1000 - 900);
    const endTime = range?.to.unix() ?? Math.round(Date.now() / 1000);
    // let logstore = this.state.logstore;
    let logstore = this.state.logstore !== '' ? this.state.logstore : query.logstore;
    if (/__custom__@/.test(logstore ?? '')) {
      const currentLogstore: any = getTemplateSrv()
        .getVariables()
        .find((item) => item.name === logstore?.split('@')[1]);
      const value = currentLogstore?.current?.value;
      logstore = value;
    }
    const queryStr = getTemplateSrv().replace(query.query ?? '', {}, replaceFormat);

    try {
      this.setState({
        loading: true,
      });

      datasource
        .postResource('api/gotoSLS', {
          Encoding: `encode=base64&queryString=${encodeURIComponent(
            Base64.encode(queryStr ?? '')
          )}&queryTimeType=99&startTime=${startTime}&endTime=${endTime}`,
          logstore: logstore,
          type: this.state.datasourceType
        })
        .then((response) => {
          this.setState({
            loading: false,
          });
          const err = response.err;
          const url = response.url;
          if (err !== '') {
            this.setState({
              showAlert: true,
              message: response.message,
              url: url,
            });
          } else {
            const newWindow = window.open(url, '_blank');
            if (newWindow) {
              newWindow.opener = null;
            }
          }
        })
        .catch((error) => {
          console.error(error);
          this.setState({
            loading: false,
          });
        });
    } catch (err) {
    } finally {
    }
  };

  getList = (type: string) => {
    const { datasource } = this.props;

    const settings = getDataSourceSrv().getInstanceSettings(datasource.uid)?.jsonData || {};

    const { project } = settings as SLSDataSourceOptions;

    try {
      datasource
        .postResource('api/getLogstoreList', {
          Project: project,
          TelemetryType: type === 'all' || !type ? '' : type === 'logstore' ? 'None' : 'Metrics',
        })
        .then((response) => {
          if (response?.data) {
            this.setState({
              logstoreList: response.data || [],
            });
          }
        })
        .catch((error) => {
          console.log('error', error);
        });
    } catch (err) {
      console.log('err', err);
    }
  };

  getCollapsedInfo(query: SLSQuery): string[] {
    const items: string[] = [];

    items.push(`Legend: ${query.legendFormat || 'auto'}`);
    items.push(`Step: ${query.step || 'auto'}`);
    items.push(`Type: ${query.queryType || 'range'}`);
    return items;
  }

  render() {
    const { isVariable, datasource } = this.props;
    const settings = getDataSourceSrv().getInstanceSettings(datasource.uid)?.jsonData || {};
    const dq = defaults(this.props.query, isVariable ? defaultEidtorQuery : defaultQuery);
    const { query, xcol, ycol, type, logstore, queryType, legendFormat, step, totalLogs, powerSql } = dq;
    const { logstore: defaultLogstore } = settings as SLSDataSourceOptions;
    const { logstoreList } = this.state;
    const variables = getTemplateSrv().getVariables();
    const customLogstore = variables.filter((item) => {
      return item.type === 'custom' && (/logstore/i.test(item?.label ?? '') || /logstore/i.test(item.name));
    });
    let uniqLogstoreList = [...new Set([defaultLogstore, ...logstoreList])].map((i) =>
      i === defaultLogstore
        ? { label: i, value: i, description: 'Default Logstore', id: i }
        : { label: i, value: i, id: i }
    );
    if (customLogstore.length > 0) {
      for (const ls of customLogstore) {
        uniqLogstoreList.unshift({
          label: ls?.label ?? ls.name,
          value: `__custom__@${ls.name}`,
          id: `__custom__@${ls.name}`,
          description: 'Custom Logstore',
        });
      }
    }

    return (
      <>
        <div className="gf-form-inline" style={{ lineHeight: '32px', verticalAlign: 'center' }}>
          <InlineField label={'数据源类型'} labelWidth={24}>
            <Select width={24} options={dataSourceType} onChange={this.onDatasourceTypeChange} value={type} />
          </InlineField>

          <InlineField label={'日志库列表'} labelWidth={24}>
            <Select
              width={24}
              options={uniqLogstoreList}
              onChange={this.onLogstoreChange}
              value={logstore || defaultLogstore}
              allowCustomValue={true}
            />
          </InlineField>

          {!isVariable && (
            <Button
              style={{ float: 'right', marginLeft: '10px' }}
              disabled={this.state.loading}
              fill="outline"
              onClick={() => {
                this.gotoSLS();
              }}
            >
              <Icon name="external-link-alt" />
              {this.state.loading ? ' loading...' : ' goto SLS'}
            </Button>
          )}

          {!isVariable && (this.state.datasourceType === 'logstore' || this.state.datasourceType === 'all') && (
            <InlineField label={'增强Sql'} style={{ marginLeft: '10px' }}>
              <InlineSwitch onChange={this.onPowerSqlChange} value={powerSql} />
            </InlineField>

          )}
        </div>

        <div className="gf-form gf-form--grow flex-shrink-1 min-width-15">
          <ConfirmModal
            isOpen={this.state.showAlert}
            title="配置STS跳转出错"
            body={`原因: ${this.state.message} 点击Confirm按照非STS逻辑跳转`}
            confirmText="Confirm"
            icon="exclamation-triangle"
            onConfirm={() => {
              const newWindow = window.open(this.state.url, '_blank');
              if (newWindow) {
                newWindow.opener = null;
              }
              this.setState({
                showAlert: false,
                message: '',
                url: '',
              });
            }}
            onDismiss={() => {
              this.setState({
                showAlert: false,
              });
            }}
          />
          <InlineFormLabel width={6} className="query-keyword">
            Query
          </InlineFormLabel>
          {version === '' ||
          version.startsWith('8.0') ||
          version.startsWith('8.1') ||
          version.startsWith('8.2') ||
          version.startsWith('7') ? (
            // <input
            //   className="gf-form-input"
            //   value={query}
            //   onChange={this.onQueryTextChange}
            //   onBlur={this.onQueryTextChange}
            //   style={{ fontFamily: 'Menlo, monospace' }}
            // ></input>
            <div style={{ width: '100%' }}>
              <MonacoQueryFieldOld
                onChange={this.onQueryTextChangeString}
                onRunQuery={this.onQueryTextChangeWithRunQuery}
                onBlur={this.onQueryTextChangeString}
                initialValue={query ?? ''}
                languageProvider={{} as any}
                history={[]}
                placeholder={''}
              />
            </div>
          ) : (
            <MonacoQueryField
              onChange={this.onQueryTextChangeString}
              onRunQuery={this.onQueryTextChangeWithRunQuery}
              onBlur={this.onQueryTextChangeString}
              initialValue={query ?? ''}
              languageProvider={{} as any}
              history={[]}
              placeholder={''}
            />
          )}
        </div>

        {type !== 'metricstore' && !isVariable && (
          <div className="gf-form-inline" style={{ lineHeight: '32px', verticalAlign: 'center' }}>
            <InlineField label={'ycol'} labelWidth={12}>
              <Input
                width={60}
                prefix={<Icon name="text-fields" />}
                value={ycol}
                onChange={this.onYChange}
                label="ycol"
                suffix={
                  <Tooltip content={<SelectTips type="ycol" />} interactive theme="info-alt">
                    <Icon name="question-circle" />
                  </Tooltip>
                }
              />
            </InlineField>

            <InlineField label={'xcol'} labelWidth={12}>
              <div style={{ display: 'flex' }}>
                <Select
                  width={20}
                  menuShouldPortal
                  options={xSelectOptions}
                  value={this.autoSelect(xcol ?? 'time')}
                  onChange={(v) => {
                    const { onChange, query, onRunQuery } = this.props;
                    if (v.value !== 'custom') {
                      this.setState({
                        xColDisabled: true,
                      });
                      onChange({ ...query, xcol: v.value });
                      onRunQuery();
                    } else {
                      this.setState({
                        xColDisabled: false,
                      });
                      onChange({ ...query, xcol: 'time' });
                    }
                  }}
                  prefix={<Icon name="palette" />}
                />
                <Input
                  width={this.state.xColDisabled ? 20 : 40}
                  disabled={this.state.xColDisabled}
                  prefix={<Icon name="x" />}
                  value={xcol}
                  onChange={this.onXChange}
                  label="xcol"
                  suffix={
                    <Tooltip content={<SelectTips type="xcol" />} interactive theme="info-alt">
                      <Icon name="question-circle" />
                    </Tooltip>
                  }
                />
              </div>
            </InlineField>

            {this.props.query.xcol === '' && (
              <InlineField label={'totalLogs'} labelWidth={12}>
                <Input
                  type="number"
                  width={20}
                  prefix={<Icon name="text-fields" />}
                  value={totalLogs}
                  onChange={this.onTotalLogsChange}
                  label="totalLogs"
                  suffix={
                    <Tooltip content={<SelectTips type="totalLogs" />} interactive theme="info-alt">
                      <Icon name="question-circle" />
                    </Tooltip>
                  }
                />
              </InlineField>
            )}
          </div>
        )}

        {type === 'metricstore' && (
          <EditorRow>
            <div>
              <QueryOptionGroup title="Options" collapsedInfo={this.getCollapsedInfo(dq)}>
                {!isVariable && (
                  <EditorField label="Format" style={{ margin: '0 0 4px 4px' }}>
                    <Input
                      type="text"
                      placeholder={'auto'}
                      onChange={this.onFormatChange}
                      defaultValue={legendFormat}
                    />
                  </EditorField>
                )}
                <EditorField label="Step" style={{ margin: '0 0 4px 4px' }}>
                  <Input type="text" placeholder={'auto'} onChange={this.onStepChange} defaultValue={step} />
                </EditorField>
                <EditorField label="Type">
                  <RadioButtonGroup
                    options={[
                      { label: 'Range', value: 'range' as any },
                      { label: 'Instant', value: 'instant' },
                    ]}
                    value={(queryType || 'range') as any}
                    onChange={this.onQueryTypeChange}
                  />
                </EditorField>
              </QueryOptionGroup>
            </div>
          </EditorRow>
        )}
      </>
    );
  }
}
