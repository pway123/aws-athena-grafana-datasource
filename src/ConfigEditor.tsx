import React, { PureComponent, ChangeEvent } from 'react';
import { SecretFormField, FormField } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { AthenaDsOptions, AthenaDsSecureJsonData } from './types';

interface Props extends DataSourcePluginOptionsEditorProps<AthenaDsOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onAccessKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      accessKey: event.target.value,
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

  onWorkgroupChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      workGroup: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  onSecretAccessKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        secretAccessKey: event.target.value,
      },
    });
  };

  onResetSecretAccessKey = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        secretAccessKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        secretAccessKey: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as AthenaDsSecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="Access Key"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onAccessKeyChange}
            value={jsonData.accessKey || ''}
            placeholder="aws access key"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="Region"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onRegionChange}
            value={jsonData.region || ''}
            placeholder="ap-southeast-1"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="Workgroup"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onWorkgroupChange}
            value={jsonData.workGroup || ''}
            placeholder="primary"
          />
        </div>

        <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.secretAccessKey) as boolean}
              value={secureJsonData.secretAccessKey || ''}
              label="Secret Key"
              placeholder="aws secret access key"
              labelWidth={6}
              inputWidth={20}
              onReset={this.onResetSecretAccessKey}
              onChange={this.onSecretAccessKeyChange}
            />
          </div>
        </div>
      </div>
    );
  }
}
