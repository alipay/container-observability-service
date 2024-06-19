
export interface SimpleOptions {
  filterConfig: FilterConfig[]
  options: FilterOption[]
}

export interface FilterConfig {
  filterKey: string;
  optionConnectMark: string;
  valueConnectMark: string;
  keyPrefix: string;
  keySuffix: string
  valuePrefix: string;
  valueSuffix: string
}

export interface FilterOption {
  label: string;
  value: string;
  belongTo: string[]
  isOpen: boolean
}
