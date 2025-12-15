import React, { ChangeEvent, PureComponent } from 'react';
import { Input, InlineField } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { HeaderWithValue, SLSDataSourceOptions, SLSSecureJsonData } from './types';
import CustomHeaders from './custom-header/custom-headers';
import { ConfigSection } from './components/configSection';
import { DataSourceDescription } from './components/description';
import { SecretInput } from './components/QueryEditor/SecretInput';
import { InlineSwitch } from './components/Switch/switch';

interface Props extends DataSourcePluginOptionsEditorProps<SLSDataSourceOptions> {}

interface State {}

export class SLSConfigEditor extends PureComponent<Props, State> {
  onEndpointChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    options.url = event.target.value;
    onOptionsChange({ ...options });
  };

  onProjectChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      project: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onLogStoreChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      logstore: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onRegionChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      region: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onRoleArnChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      roleArn: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  onAKIDChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    var accessKeySecret = '';
    if (options.secureJsonData !== undefined) {
      if (options.secureJsonData.hasOwnProperty('accessKeySecret')) {
        // @ts-ignore
        accessKeySecret = options.secureJsonData['accessKeySecret'];
      }
    }
    onOptionsChange({
      ...options,
      secureJsonData: {
        accessKeyId: event.target.value,
        accessKeySecret: accessKeySecret,
      },
    });
  };

  onAKSecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    var accessKeyId = '';
    if (options.secureJsonData !== undefined) {
      if (options.secureJsonData.hasOwnProperty('accessKeyId')) {
        // @ts-ignore
        accessKeyId = options.secureJsonData['accessKeyId'];
      }
    }
    onOptionsChange({
      ...options,
      secureJsonData: {
        accessKeyId: accessKeyId,
        accessKeySecret: event.target.value,
      },
    });
  };

  onResetAKID = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        accessKeyId: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        accessKeyId: '',
      },
    });
  };
  onResetAKSecret = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        accessKeySecret: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        accessKeySecret: '',
      },
    });
  };

  onHeadersChange = (headers: HeaderWithValue[]) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      headers,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onInternationalChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      international: event.target.checked,
    };
    onOptionsChange({ ...options, jsonData });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields, url } = options;
    const secureJsonData = (options.secureJsonData || {}) as SLSSecureJsonData;

    return (
      <>
        <DataSourceDescription
          docsLink="https://help.aliyun.com/zh/sls/developer-reference/connect-log-service-to-grafana"
          dataSourceName="SLS数据源"
        />
        <hr style={{ margin: '20px 0' }} />
        <ConfigSection title="Connection">
          {/* EndPonit */}
          <InlineField
            htmlFor="connection-url"
            label={'Endpoint *'}
            labelWidth={24}
            tooltip={
              <>
                Specify a complete Endpoint URL
                <br />
                (for example: cn-hangzhou.log.aliyuncs.com)
              </>
            }
            grow
            interactive
          >
            <Input
              id="connection-url"
              aria-label="Project 所在的地域"
              onChange={this.onEndpointChange}
              value={url || ''}
              placeholder={'请输入 Endpoint'}
            />
          </InlineField>

          {/*  Project */}
          <InlineField
            htmlFor="connection-project"
            label={'Project *'}
            tooltip={'Project'}
            labelWidth={24}
            grow
            interactive
          >
            <Input
              id="connection-project"
              aria-label="Project"
              onChange={this.onProjectChange}
              value={jsonData.project || ''}
              placeholder={'请输入 Project'}
            />
          </InlineField>

          {/* AccessKeyID */}
          <InlineField
            htmlFor="connection-AccessKeyID"
            label={'AccessKeyID *'}
            labelWidth={24}
            tooltip={'阿里云账号AccessKeyID'}
            grow
            interactive
          >
            <SecretInput
              isConfigured={(secureJsonFields && secureJsonFields.accessKeyId) as boolean}
              value={secureJsonData.accessKeyId || ''}
              label="AccessKeyID"
              placeholder={'阿里云账号AccessKeyID'}
              onReset={this.onResetAKID}
              onChange={this.onAKIDChange}
            />
          </InlineField>

          {/* AccessKeySecret */}
          <InlineField
            htmlFor="connection-AccessKeySecret"
            label={'AccessKeySecret *'}
            labelWidth={24}
            tooltip={'阿里云账号AccessKeySecret'}
            grow
            interactive
          >
            <SecretInput
              isConfigured={(secureJsonFields && secureJsonFields.accessKeySecret) as boolean}
              value={secureJsonData.accessKeySecret || ''}
              label="AccessKeySecret"
              placeholder={'阿里云账号AccessKeySecret'}
              onReset={this.onResetAKSecret}
              onChange={this.onAKSecretChange}
            />
          </InlineField>

          {/* Other  */}
          <InlineField
            style={{ marginTop: '8px' }}
            htmlFor="connection-Region"
            label={'Region *'}
            labelWidth={24}
            tooltip={
              <>
                Region
                <br />
                (for example: cn-hangzhou)
              </>
            }
            grow
            interactive
          >
            <Input
              id="connection-Region"
              aria-label="region"
              onChange={this.onRegionChange}
              value={jsonData.region || ''}
              placeholder={'请输入 Region(选填, optional)'}
            />
          </InlineField>
        </ConfigSection>

        {/* Other  */}
        <hr style={{ margin: '20px 0' }} />
        <ConfigSection title="Other">
          {/* international */}
          <InlineField
            htmlFor="connection-international"
            label={'International'}
            labelWidth={24}
            tooltip={
              <>
                Whether to use the international endpoint
              </>
            }
            grow
            interactive
          >
            <InlineSwitch
              id="connection-international"
              aria-label="是否为国际站"
              onChange={this.onInternationalChange}
              value={jsonData.international || false}
            />
          </InlineField>
          <InlineField
            style={{ marginTop: '8px' }}
            tooltip={'请输入默认的Logstore,如果不填写,请确保您的 ak 账号具备该 project 的ListLogStore读权限'}
            htmlFor="connection-logstore"
            label={'Default Logstore'}
            labelWidth={24}
            grow
            interactive
          >
            <Input
              id="other-logstore-url"
              aria-label="logstore"
              placeholder='请输入默认的Logstore'
              onChange={this.onLogStoreChange}
              value={jsonData.logstore || ''}
            />
          </InlineField>

          {/* roleArn */}
          <InlineField
            style={{ marginTop: '8px' }}
            htmlFor="connection-roleArn"
            label={'RoleArn'}
            labelWidth={24}
            grow
            interactive
          >
            <Input
              id="other-roleArn-url"
              aria-label="roleArn"
              onChange={this.onRoleArnChange}
              value={jsonData.roleArn || ''}
              placeholder={'配置STS跳转(选填, optional)'}
            />
          </InlineField>

          {/* headers */}
          <CustomHeaders headers={jsonData.headers || []} onChange={this.onHeadersChange} readOnly={false} />
        </ConfigSection>
      </>
    );
  }
}
