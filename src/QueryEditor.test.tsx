import { SLSQueryEditor } from './QueryEditor';
import { getDataSourceSrv, getTemplateSrv } from '@grafana/runtime';

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getDataSourceSrv: jest.fn(),
  getTemplateSrv: jest.fn(),
}));

jest.mock('./SLS-monaco-editor/MonacoQueryField', () => () => null);
jest.mock('./SLS-monaco-editor/MonacoQueryFieldOld', () => () => null);

const flushPromises = () => new Promise((resolve) => setTimeout(resolve, 0));

describe('data source resources', () => {
  let postResource: jest.Mock;
  let editor: SLSQueryEditor;

  beforeEach(() => {
    postResource = jest.fn();
    (getDataSourceSrv as jest.Mock).mockReturnValue({
      getInstanceSettings: () => ({ jsonData: { project: 'example-project' } }),
    });
    (getTemplateSrv as jest.Mock).mockReturnValue({ replace: (value: string) => value });
    editor = new SLSQueryEditor({
      datasource: { uid: 'example-uid', postResource },
      query: { query: '* | select count(*)', logstore: 'example-logstore' },
    } as any);
    editor.setState = jest.fn((state) => {
      editor.state = { ...editor.state, ...state };
    }) as any;
  });

  it.each([
    ['all', ''],
    ['logstore', 'None'],
    ['metricsql', 'Metrics'],
    ['metricstore', 'Metrics'],
  ])('loads %s stores through the runtime resource API', async (type, telemetryType) => {
    postResource.mockResolvedValue({ data: ['example-logstore'] });
    editor.getList(type);
    await flushPromises();
    expect(postResource).toHaveBeenCalledWith('api/getLogstoreList', {
      Project: 'example-project',
      TelemetryType: telemetryType,
    });
    expect(editor.state.logstoreList).toEqual(['example-logstore']);
  });

  it('opens the URL returned by the runtime resource API', async () => {
    const popup = { opener: {} };
    const open = jest.spyOn(window, 'open').mockReturnValue(popup as any);
    postResource.mockResolvedValue({ err: '', url: 'https://example.com/logsearch', message: '' });
    editor.gotoSLS();
    await flushPromises();
    expect(postResource).toHaveBeenCalledWith('api/gotoSLS', {
      Encoding: expect.stringContaining('encode=base64&queryString='),
      logstore: 'example-logstore',
      type: 'all',
    });
    expect(open).toHaveBeenCalledWith('https://example.com/logsearch', '_blank');
    expect(popup.opener).toBeNull();
    expect(editor.state.loading).toBe(false);
    open.mockRestore();
  });

  it('preserves the warning returned by the resource handler', async () => {
    postResource.mockResolvedValue({ err: 'roleCheckError', message: 'Invalid role', url: 'https://example.com/' });
    editor.gotoSLS();
    await flushPromises();
    expect(editor.state).toMatchObject({
      loading: false,
      showAlert: true,
      message: 'Invalid role',
      url: 'https://example.com/',
    });
  });
});
