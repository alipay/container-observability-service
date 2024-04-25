<<<<<<< HEAD
type SeriesSize = 'sm' | 'md' | 'lg';

export interface SimpleOptions {
  text: string;
  showSeriesCount: boolean;
  seriesCountSize: SeriesSize;
=======
export type DisplayModel = 'yaml' | 'json';
export type Theme = 'idea' | 'base16-dark';

export interface SimpleOptions {
  displayModel: DisplayModel;
  theme: Theme;
>>>>>>> main
}
