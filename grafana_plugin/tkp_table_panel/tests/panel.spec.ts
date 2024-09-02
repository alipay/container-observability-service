import { test, expect } from '@grafana/plugin-e2e';

test('should display "No data" in case panel data is empty', async ({
  gotoPanelEditPage,
  readProvisionedDashboard,
}) => {
  const dashboard = await readProvisionedDashboard({ fileName: 'dashboard.json' });
  const panelEditPage = await gotoPanelEditPage({ dashboard, id: '2' });
  await expect(panelEditPage.panel.locator).toContainText('No data');
});

test('should display circle when data is passed to the panel', async ({
  panelEditPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Tkp-Table');
  await expect(page.getByTestId('simple-panel-circle')).toBeVisible();
});

test('should display series counter when "Show series counter" option is enabled', async ({
  panelEditPage,
  readProvisionedDataSource,
  page,
  selectors,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Tkp-Table');
  await panelEditPage.collapseSection('Tkp-Table');
  await expect(page.getByTestId('simple-panel-circle')).toBeVisible();
  const showSeriesSwitch = panelEditPage
    .getByGrafanaSelector(selectors.components.PanelEditor.OptionsPane.fieldLabel('Tkp-Table Show series counter'))
    .getByLabel('Toggle switch');
  await showSeriesSwitch.click();
  await expect(page.getByTestId('simple-panel-series-counter')).toBeVisible();
});
