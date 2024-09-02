import { PanelPlugin } from '@grafana/data';
import { SimpleOptions } from './types';
import { SimplePanel } from './components/SimplePanel';

export const plugin = new PanelPlugin<SimpleOptions>(SimplePanel).setPanelOptions((builder) => {
  return builder
    .addTextInput({
      path: 'title',
      name: 'Panel title',
      description: 'Edit title of panel',
      defaultValue: '',
    })
    .addTextInput({
      path: 'searchUrl',
      name: 'SearchUrl',
      description: 'Edit SearchUrl',
      defaultValue: '',
    })
    .addTextInput({
      path: 'tkpHosting',
      name: 'TKPHosting',
      description: 'Edit TKPHosting',
      defaultValue: '',
    })
    .addTextInput({
      path: 'debugPodUrl',
      name: 'DebugPodUrl',
      description: 'Edit DebugPodUrl',
      defaultValue: '',
    })
});
