import { defaults } from 'lodash';

import React, { ChangeEvent, PureComponent, SyntheticEvent } from 'react';
import { Button, LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, AnnoDataSourceOptions, AnnoQuery } from './types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, AnnoQuery, AnnoDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onTagsChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, tags: event.target.value });
  };

  onSubmit = () => {
    const { onRunQuery } = this.props;
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { tags } = query;

    return (
      <div className="gf-form">
        <FormField labelWidth={8} value={tags || ''} onChange={this.onTagsChange} label="Annotation Tags" />
        <Button onClick={this.onSubmit} />
      </div>
    );
  }
}
