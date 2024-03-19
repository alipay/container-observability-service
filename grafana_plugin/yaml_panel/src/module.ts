import { PanelPlugin } from '@grafana/data';
import { SimpleOptions } from './types';
import { SimplePanel } from './components/YamlPanel';
// import { SimplePanel } from './components/SimplePanel';

export const plugin = new PanelPlugin<SimpleOptions>(SimplePanel).setPanelOptions((builder) => {
  return builder
    .addRadio({
      path: 'displayModel',
      defaultValue: 'json',
      name: 'Display Model',
      settings: {
        options: [
          {
            value: 'json',
            label: 'JSON',
          },
          {
            value: 'yaml',
            label: 'YAML',
          }
        ],
      }
    })
    .addRadio({
      path: 'theme',
      defaultValue: 'idea',
      name: 'Default Theme',
      settings: {
        options: [
          {
            value: 'idea',
            label: 'Light',
          },
          {
            value: 'base16-dark',
            label: 'Dark',
          }
        ],
      }
    });
});
