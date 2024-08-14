import React, { useEffect, useState } from 'react';
import { PanelProps } from '@grafana/data';
import { SimpleOptions } from 'types';
import {
  Table,
  Form,
  Input,
  Select,
  Button,
  Card,
  TableProps,
  Space,
  message,
  PopconfirmProps,
  Tag,
  notification,
  Modal,
  Row,
  Col,
  Tooltip,
} from 'antd';
import dayjs from 'dayjs';
import axios from 'axios';
import { locationService } from '@grafana/runtime';

interface Props extends PanelProps<SimpleOptions> { }

type NotificationType = 'success' | 'info' | 'warning' | 'error';


interface DataType {
  action: string;
  cluster: string;
  namespace: string;
  state: string;
  createTime: any,
  nodeip: string,
  podip: string,
  podname: string,
  podphase: string,
  poduid: string,
  workloadInfo: {
    Name?: string
  },
  ableToTkp?: boolean,
  tooltip?: string
}

function getTagColor(value: string): string {
  switch (value) {
    case '在线': case 'Running': case 'running': case 'Succeeded':
      return 'success';
    case '已删除': case 'Terminating': case 'Pending': case 'Failed': case 'terminated':
      return 'error';
    case '发起删除': 
      return 'warning';
    default:
      return 'default'
  }
}


export const SimplePanel: React.FC<Props> = ({ options, data, width, height, fieldConfig, id, replaceVariables }) => {
  const [form] = Form.useForm();
  const [tableData, setTableData] = useState([])
  const [isLoading, setIsloading] = useState(false)
  const [api, contextHolder] = notification.useNotification();
  const [isOpen, setIsOpen] = useState(false);
  const [isResetOpen, setIsResetOpen] = useState(false);
  const [podinfo, setPodinfo] = useState<any>();

  const openNotificationWithIcon = (type: NotificationType, message: string, description: string) => {
    api[type]({
      message: message,
      description: description,
    });
  };


  const onSearch = () => {
    setIsloading(true)
    form.validateFields().then(async (values: any) => {
      locationService.partial({ 'var-podinfo': values.podinfo, 'var-podinfovalue': values.podinfovalue }, true);
      axios.get(options.searchUrl.replace('$podinfo', values['podinfo']).replace('$podinfovalue', values['podinfovalue']))
        .then(response => {
          if (response.status === 200) {
            setTableData(response.data.map((item: any) => {
              item.ableToTkp = true;
              return item
            }))
          }
        })
        .catch(error => {
          throw new Error(error)
        })
        .finally(() => {
          setIsloading(false)
        })
    });
  }

  useEffect(() => {
    const key = replaceVariables('$podinfo');
    const value = replaceVariables('$podinfovalue');
    if (key && value && tableData.length === 0) {
      form.setFieldsValue({ podinfo: key, podinfovalue: value })
      onSearch()
    }
  },[])

  const cancel: PopconfirmProps['onCancel'] = () => {
    message.error('取消操作');
  };

  const handleTkpHosting = () => {
    if (!podinfo.workloadInfo.Name) {
      openNotificationWithIcon('warning', '托管中断', '当前pod没有所属的workload, 无法托管')
      const newList = [...tableData]
      newList.map((item: any) => {
        if (item.poduid === podinfo.poduid) {
          item.ableToTkp = false
          item.tooltip = '当前pod没有所属的workload, 无法托管'
        }
        return item
      })
      setTableData(newList)
      return
    }
    const url = options.tkpHosting;
    axios.post(url, {
      siteName: podinfo.site,
      namespace: podinfo.namespace,
      podName: podinfo.podname,
      cluster: `sigma-${podinfo.cluster}`
    })
      .then(response => {
        if (response.data.code === 200 && response.data.data.pod_name && response.data.data.pod_namespace) {
          if (response.data.data.vtkp) {
            localStorage.setItem('tkpName', response.data.data.vtkp)
            locationService.partial({
              "var-workloadName": podinfo.workloadInfo.Name,
              'var-tkpName': response.data.data.vtkp,
              'var-cluster': podinfo.cluster,
              'var-namespace': podinfo.namespace,
              'var-site': podinfo.site,
              'refresh': '5s', 'from': 'now-5m', 'to': 'now'
            }, true);
            openNotificationWithIcon('success', '开始托管', 'TKP已经开始托管您的Pod')
            return
          } else {
            handleTkpHosting()
          }
        } else if (response.data.code === 400) {
          localStorage.removeItem('tkpName')
          locationService.partial({}, true);
          openNotificationWithIcon('warning', '托管中断', '当前pod所对应的workload还未被TKP托管')
          const newList = [...tableData]
          newList.map((item: any) => {
            if (item.poduid === podinfo.poduid) {
              item.ableToTkp = false
              item.tooltip = '当前pod所对应的workload还未被TKP托管'
            }
            return item
          })
          setTableData(newList)
          return
        } else if (response.data.code === 404) {
          localStorage.removeItem('tkpName')
          locationService.partial({}, true);
          openNotificationWithIcon('warning', '托管中断', '当前pod已删除')
          const newList = [...tableData]
          newList.map((item: any) => {
            if (item.poduid === podinfo.poduid) {
              item.ableToTkp = false
              item.tooltip = '当前pod已删除'
            }
            return item
          })
          setTableData(newList)
          return
        }
      })
      .catch(error => {
        throw new Error(error)
      })
      .finally()
  }

  const handleReset = () => {
    const currentUrl = window.location.href
    const newUrl = currentUrl.split("?")[0];
    form.resetFields(); setTableData([])
    window.open(newUrl, '_self')
  }

  const columns: TableProps<DataType>['columns'] = [
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle" key={record.poduid + 'space'}>
          {record.ableToTkp ? <a key={record.poduid + 'tkp'} style={{ color: '#0f71f8' }} onClick={() => { setIsOpen(true); setPodinfo(record) }}>开始托管</a>
            : <Tooltip placement="top" title={record.tooltip}>
              <Tag color='default'>无法托管</Tag>
            </Tooltip>}
          <a key={record.poduid + 'check'} style={{ color: '#0f71f8' }} onClick={() => { window.open(`${options.debugPodUrl}&var-podinfo=uid&var-podinfovalue=${record.poduid}`, '_black') }}>诊断</a>
        </Space>
      ),
      width: 150
    },
    {
      title: 'PodName',
      dataIndex: 'podname',
      key: 'podname',
      width: 300,
    },
    {
      title: '集群',
      dataIndex: 'cluster',
      key: 'cluster',
      sorter: true,
      width: 150
    },
    {
      title: '命名空间',
      dataIndex: 'namespace',
      key: 'namespace',
      sorter: true,
      width: 150
    },
    {
      title: 'PodPhase',
      dataIndex: 'podphase',
      key: 'podphase',
      width: 150,
      render: (_, record) => {
        return <Tag color={getTagColor(record.podphase)}>{record.podphase}</Tag>
      }
    },
    {
      title: '状态',
      dataIndex: 'state',
      key: 'state',
      sorter: true,
      width: 150,
      render: (_, record) => {
        return <Tag color={getTagColor(record.state)}>{record.state}</Tag>
      }
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime',
      width: 150,
      sorter: true,
      render: (_, record) => {
        return <p key={record.createTime}>{dayjs(record.createTime).format('YYYY-MM-DD HH:mm:ss')}</p>
      }
    },
    {
      title: 'NodeIp',
      dataIndex: 'nodeip',
      key: 'nodeip',
      width: 150,
    },
    {
      title: 'PodIp',
      dataIndex: 'podip',
      key: 'podip',
      width: 150,
    },
    {
      title: 'PodUid',
      dataIndex: 'poduid',
      key: 'poduid',
      width: 300,
    },
    {
      title: 'Workload',
      dataIndex: 'workloadInfo',
      key: 'workloadInfo',
      width: 300,
      render: (_, record) => {
        return (record.workloadInfo && record.workloadInfo.Name && <p key={record.workloadInfo.Name}>{record.workloadInfo.Name}</p>)
      }
    },
  ];

  return (
    <div style={{ width: width, height: height }}>
      {contextHolder}
      <Card
        style={{ width: width, height: height }}
        extra={<>
          <Form
            layout='inline'
            form={form}
            style={{ width: width }}
            onFinish={onSearch}
          >
            <Form.Item label="查询条件" name='podinfo' rules={[{ required: true, message: '请选择一个查询条件' }]} style={{ minWidth: width * 0.25 }} initialValue='name'>
              <Select >
                <Select.Option value="name">podName</Select.Option>
                <Select.Option value="uid">podUid</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item label="查询内容" name='podinfovalue' rules={[{ required: true, message: '请输入要查询的内容' }]} style={{ minWidth: width * 0.46 }}>
              <Input placeholder="请输入查询内容" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" style={{ marginLeft: 8 }} loading={isLoading}>查询</Button>
              <Button type="default" onClick={() => setIsResetOpen(true)} style={{ marginLeft: 8 }}>重置</Button>
            </Form.Item>
          </Form>
        </>}
      >
        <Table columns={columns} dataSource={tableData} pagination={false} size='middle' showHeader={tableData.length !== 0} scroll={tableData.length === 0 ? undefined : { x: 'calc(700px + 50%)', y: height * 0.7 }} />
      </Card>
      <Modal title="托管确认" open={isOpen}
        footer={(_, { }) => (
          <>
            <Button type='primary' onClick={() => { handleTkpHosting(); setIsOpen(false) }}>开始托管</Button>
            <Button type='default' onClick={() => { cancel(); setIsOpen(false) }}>取消操作</Button>
          </>
        )}>
        <Row>
          <Col span={4}>Pod名称: </Col>
          <Col span={20}>{podinfo?.podname}</Col>
        </Row>
        <Row>
          <Col span={4}>命名空间: </Col>
          <Col span={20}>{podinfo?.namespace}</Col>
        </Row>
        <Row>
          <Col span={4}>集群:</Col>
          <Col span={20}>{podinfo?.cluster}</Col>
        </Row>
        <Row>
          <Col span={4}>workload:</Col>
          <Col span={20}>{podinfo?.workloadInfo.Name}</Col>
        </Row>
      </Modal>
      <Modal title="确定要重置吗" open={isResetOpen}
        footer={(_, { }) => (
          <>
            <Button type='default' onClick={handleReset}>确定重置</Button>
            <Button type='primary' onClick={() => { setIsResetOpen(false) }}>取消操作</Button>
          </>
        )}>
        重置后所有查询内容（包含托管的实时动态等）都会重置，页面将回到初始状态，确定重置吗？
      </Modal>
    </div>
  );
};
