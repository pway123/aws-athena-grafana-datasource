import React, { PureComponent, ChangeEvent } from 'react';
import { SecretFormField, FormField, FormLabel, Select } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, SelectableValue } from '@grafana/data';
import { AthenaDsOptions, AthenaDsSecureJsonData, AuthType } from './types';

interface Props extends DataSourcePluginOptionsEditorProps<AthenaDsOptions> {}

interface State {
  selectedAuthType: SelectableValue;
}

const authTypes = [
  { label: 'Static', value: AuthType.Static },
  { label: 'Role ARN', value: AuthType.RoleArn },
];

export class ConfigEditor extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      selectedAuthType: authTypes.find(t => t.value === props.options.jsonData.authType) || authTypes[0],
    };
  }

  onAccessKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      accessKey: event.target.value,
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
        <div className="gf-form-inline">
          <FormLabel width={6}>Format</FormLabel>
          <Select
            width={20}
            options={authTypes}
            value={this.state.selectedAuthType}
            onChange={v => {
              const jsonData = {
                ...options.jsonData,
                authType: v.value,
              };
              this.props.onOptionsChange({ ...options, jsonData });
              this.setState({ selectedAuthType: v });
            }}
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
        {this.state.selectedAuthType.value === AuthType.Static && (
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
        )}
        {this.state.selectedAuthType.value === AuthType.Static && (
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
        )}
        {this.state.selectedAuthType.value === AuthType.RoleArn && (
          <div className="gf-form">
            <FormField
              label="Role ARN"
              labelWidth={6}
              inputWidth={20}
              onChange={this.onRoleArnChange}
              value={jsonData.roleArn || ''}
              placeholder="role arn"
            />
          </div>
        )}
      </div>
    );
  }
}
