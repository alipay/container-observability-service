import React from 'react';
import { PanelProps } from '@grafana/data';
import { SimpleOptions } from 'types';

import { PanelDataErrorView } from '@grafana/runtime';
import FilterPanel from './FilterComponent';

interface Props extends PanelProps<SimpleOptions> {}


export const SimplePanel: React.FC<Props> = ({ options, data, width, height, fieldConfig, id, replaceVariables }) => {


  if (data.series.length === 0) {
    return <PanelDataErrorView fieldConfig={fieldConfig} panelId={id} data={data} needsStringField />;
  }

  return (
    <div>
      <FilterPanel options={options} data={data} width={width} height={height} replaceVariables={replaceVariables}></FilterPanel>
    </div>
  );
};
