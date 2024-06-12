import { PanelPlugin } from '@grafana/data';
import { SimpleOptions } from './types';
import { SimplePanel } from './components/SimplePanel';
import { FilterConfigEditort } from './components/FilterConfigEditort'
import { FilterOptionEditort } from './components/FilterOptionEditort'

export const plugin = new PanelPlugin<SimpleOptions>(SimplePanel).setPanelOptions((builder) => {
  return builder
  .addCustomEditor({
    id: 'filterConfig',
    path: 'filterConfig',
    name: 'Filter Config',
    defaultValue: [],
    editor: FilterConfigEditort,
  })
  .addCustomEditor({
    id: 'options',
    path: 'options',
    name: 'Options',
    defaultValue: [],
    editor: FilterOptionEditort,
  })
});
