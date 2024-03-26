
import * as common from '@grafana/schema';
export interface FieldConfig {
  name: string;
  type: string;
  overrideName: string;
  url: string

}
export interface SimpleOptions {
  dedupStrategy: common.LogsDedupStrategy;
  enableLogDetails: boolean;
  prettifyLogMessage: boolean;
  showCommonLabels: boolean;
  showLabels: boolean;
  showTime: boolean;
  sortOrder: common.LogsSortOrder;
  wrapLogMessage: boolean;
  label: FieldConfig[];
  params: string[]
}
