import { PanelPlugin } from '@grafana/data';
import { SimpleOptions } from './types';
import { SimplePanel } from './components/SimplePanel';
import { SimpleEditor } from './components/SimpleEditor';

export const plugin = new PanelPlugin<SimpleOptions>(SimplePanel).setPanelOptions((builder) => {

  return builder
    .addStringArray({
      path: 'params',
      name: 'Variables',
      defaultValue: ['podinfo', 'podinfovalue'],
    })
    .addBooleanSwitch({
      path: 'showTime',
      name: 'Time',
      defaultValue: true,
    })
    .addBooleanSwitch({
      path: 'showLabels',
      name: 'Unique labels',
      defaultValue: false,
    })
    .addBooleanSwitch({
      path: 'showCommonLabels',
      name: 'Common labels',
      defaultValue: false,
    })
    .addBooleanSwitch({
      path: 'wrapLogMessage',
      name: 'Wrap lines',
      defaultValue: true,
    })
    .addBooleanSwitch({
      path: 'prettifyLogMessage',
      name: 'Prettify JSON',
      defaultValue: true,
    })
    .addBooleanSwitch({
      path: 'enableLogDetails',
      name: 'Enable log details',
      defaultValue: true,
    })
    .addRadio({
      path: 'sortOrder',
      defaultValue: 'Ascending',
      name: 'Order',
      settings: {
        options: [
          {
            value: 'Ascending',
            label: 'Ascending',
          },
          {
            value: 'Descending',
            label: 'Descending',
          },
        ],
      },
    })
    .addRadio({
      path: 'dedupStrategy',
      defaultValue: 'signature',
      name: 'Deduplication',
      settings: {
        options: [
          {
            value: 'none',
            label: 'None',
          },
          {
            value: 'exact',
            label: 'Exact',
          },
          {
            value: 'numbers',
            label: 'Numbers',
          },
          {
            value: 'signature',
            label: 'Signature',
          }
        ],
      },
    })
    .addCustomEditor({
      id: 'label',
      path: 'label',
      name: 'Label',
      defaultValue:[],
      editor: SimpleEditor,
    });
});

