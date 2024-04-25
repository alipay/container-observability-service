import { PanelPlugin } from '@grafana/data';
import { SimpleOptions } from './types';
import { SimplePanel } from './components/YamlPanel';
// import { SimplePanel } from './components/SimplePanel';

export const plugin = new PanelPlugin<SimpleOptions>(SimplePanel).setPanelOptions((builder) => {
  return builder
<<<<<<< HEAD
    .addTextInput({
      path: 'text',
      name: 'Simple text option',
      description: 'Description of panel option',
      defaultValue: 'Default value of text input option',
    })
    .addBooleanSwitch({
      path: 'showSeriesCount',
      name: 'Show series counter',
      defaultValue: false,
    })
    .addRadio({
      path: 'seriesCountSize',
      defaultValue: 'sm',
      name: 'Series counter size',
      settings: {
        options: [
          {
            value: 'sm',
            label: 'Small',
          },
          {
            value: 'md',
            label: 'Medium',
          },
          {
            value: 'lg',
            label: 'Large',
          },
        ],
      },
      showIf: (config) => config.showSeriesCount,
=======
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
>>>>>>> main
    });
});
