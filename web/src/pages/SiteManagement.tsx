import React, { useState, useEffect } from 'react';
import { Table, Button, message, Tag, Space, Modal, Form, Input, Switch, Select, Popconfirm, Drawer } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined, SyncOutlined, StopOutlined, EyeOutlined } from '@ant-design/icons';
import * as api from '../services/api';
import { Site, SiteRequest, SiteRules } from '../types';

const { TextArea } = Input;
const { Option } = Select;

const SiteManagement: React.FC = () => {
  const [sites, setSites] = useState<Site[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [editingSite, setEditingSite] = useState<Site | null>(null);
  const [viewingSite, setViewingSite] = useState<Site | null>(null);
  const [form] = Form.useForm();
  const [runningTasks, setRunningTasks] = useState<number[]>([]);

  useEffect(() => {
    fetchSites();
  }, []);

  const fetchSites = async () => {
    try {
      setLoading(true);
      const res = await api.getSites();
      setSites(res.data.data || []);
    } catch (error) {
      message.error('获取站点列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRunTask = async (id: number) => {
    setRunningTasks((prev: number[]) => [...prev, id]);
    try {
      await api.runSiteTask(id);
      message.success('任务已成功在后台启动');
      // 可以在此处设置一个定时器或使用WebSocket来更新任务状态
      // 为简单起见，我们仅在启动时更新，并假设它会运行
      setSites((prevSites: Site[]) => prevSites.map((s: Site) => s.id === id ? { ...s, status: 'running' } : s));
    } catch (error) {
      message.error('启动任务失败');
    } finally {
      // 即使任务在后台运行，我们也可以从UI上移除loading状态，让用户可以继续操作
      setTimeout(() => {
        setRunningTasks((prev: number[]) => prev.filter((taskId: number) => taskId !== id));
      }, 5000); // 5秒后恢复按钮，防止任务失败时按钮一直锁定
    }
  };

  const handleEdit = (site: Site) => {
    setEditingSite(site);
    setModalVisible(true);
    form.setFieldsValue({
      ...site,
      start_urls: site.start_urls.join('\n'),
    });
  };

  const handleView = (site: Site) => {
    setViewingSite(site);
    setDrawerVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await api.deleteSite(id);
      message.success('站点删除成功');
      fetchSites();
    } catch (error) {
      message.error('删除失败');
    }
  };
  
  const handleToggle = async (id: number) => {
    try {
      await api.toggleSite(id);
      message.success('状态切换成功');
      fetchSites();
    } catch (error) {
      message.error('状态切换失败');
    }
  };


  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const requestData: SiteRequest = {
        ...values,
        start_urls: values.start_urls.split('\n').filter((url: string) => url.trim() !== ''),
      };
      
      if (editingSite) {
        await api.updateSite(editingSite.id, requestData);
        message.success('站点更新成功');
      } else {
        await api.createSite(requestData);
        message.success('站点创建成功');
      }
      setModalVisible(false);
      setEditingSite(null);
      fetchSites();
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };


  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string, record: Site) => {
        let color = 'default';
        let text = status.toUpperCase();
        if (status === 'running') {
          color = 'processing';
          text = '运行中';
        } else if (status === 'ready') {
          color = 'success';
          text = '就绪';
        } else if (status === 'failed') {
          color = 'error';
          text = '失败';
        }

        return <Tag color={record.enabled ? color : 'default'}>{record.enabled ? text : '已禁用'}</Tag>;
      }
    },
    {
      title: '上次运行',
      dataIndex: 'last_run_at',
      key: 'last_run_at',
      render: (text: string) => text ? new Date(text).toLocaleString() : '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => new Date(text).toLocaleString(),
    },
    {
        title: '启用',
        dataIndex: 'enabled',
        key: 'enabled',
        render: (enabled: boolean, record: Site) => (
            <Switch checked={enabled} onChange={() => handleToggle(record.id)} />
        )
    },
    {
      title: '操作',
      key: 'action',
      width: 350,
      render: (record: Site) => {
        const isRunning = runningTasks.includes(record.id) || record.status === 'running';
        return (
          <Space size="small">
            <Button 
              type="link" 
              icon={isRunning ? <SyncOutlined spin /> : <PlayCircleOutlined />}
              onClick={() => handleRunTask(record.id)}
              disabled={isRunning}
            >
              {isRunning ? '运行中' : '运行任务'}
            </Button>
            <Button type="link" icon={<EyeOutlined />} onClick={() => handleView(record)}>
              查看
            </Button>
            <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
              编辑
            </Button>
            <Popconfirm title="确定删除这个站点吗?" onConfirm={() => handleDelete(record.id)} okText="确定" cancelText="取消">
              <Button type="link" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          </Space>
        )
      }
    }
  ];
  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
          新建站点
        </Button>
      </div>
      <Table columns={columns} dataSource={sites} rowKey="id" loading={loading} />
      
      <Modal
        title={editingSite ? '编辑站点' : '新建站点'}
        visible={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          setEditingSite(null);
        }}
        width={800}
      >
        <Form form={form} layout="vertical" name="site_form">
          <Form.Item name="name" label="站点名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="base_url" label="基础URL" rules={[{ required: true, type: 'url' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <TextArea rows={2} />
          </Form.Item>
          <Form.Item name="start_urls" label="起始URL (每行一个)" rules={[{ required: true }]}>
            <TextArea rows={4} />
          </Form.Item>
          <Form.Item name="enabled" label="是否启用" valuePropName="checked">
            <Switch />
          </Form.Item>
          <p>CSS选择器</p>
          <Form.Item name={['selectors', 'item']} label="列表项 (item)">
            <Input placeholder="e.g., .article-item" />
          </Form.Item>
          <Form.Item name={['selectors', 'title']} label="标题 (title)">
            <Input placeholder="e.g., h2.title a" />
          </Form.Item>
          <Form.Item name={['selectors', 'content']} label="内容 (content)">
            <Input placeholder="e.g., .article-content" />
          </Form.Item>
           <Form.Item name={['selectors', 'author']} label="作者 (author)">
            <Input placeholder="e.g., .author-name" />
          </Form.Item>
          <p>爬取规则</p>
           <Form.Item name={['rules', 'concurrent']} label="并发数">
            <Input type="number" />
          </Form.Item>
          <Form.Item name={['rules', 'delay']} label="延迟 (毫秒)">
            <Input type="number" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        width={640}
        placement="right"
        closable={false}
        onClose={() => setDrawerVisible(false)}
        visible={drawerVisible}
      >
        {viewingSite && (
          <div>
            <h2>{viewingSite.name}</h2>
            <p><strong>ID:</strong> {viewingSite.id}</p>
            <p><strong>基础URL:</strong> <a href={viewingSite.base_url} target="_blank" rel="noopener noreferrer">{viewingSite.base_url}</a></p>
            <p><strong>描述:</strong> {viewingSite.description}</p>
            <p><strong>状态:</strong> <Tag>{viewingSite.status}</Tag></p>
            <p><strong>已启用:</strong> {viewingSite.enabled ? '是' : '否'}</p>
            <p><strong>创建时间:</strong> {new Date(viewingSite.created_at).toLocaleString()}</p>
            <p><strong>上次运行:</strong> {viewingSite.last_run_at ? new Date(viewingSite.last_run_at).toLocaleString() : '-'}</p>
            <h3>起始URL:</h3>
            <ul>
              {viewingSite.start_urls.map(url => <li key={url}><a href={url} target="_blank" rel="noopener noreferrer">{url}</a></li>)}
            </ul>
            <h3>选择器:</h3>
            <ul>
              {Object.entries(viewingSite.selectors).map(([key, value]) => <li key={key}><strong>{key}:</strong> <code>{value}</code></li>)}
            </ul>
             <h3>规则:</h3>
            <ul>
              <li>并发数: {viewingSite.rules.concurrent || '默认'}</li>
              <li>延迟: {viewingSite.rules.delay || '默认'} ms</li>
            </ul>
          </div>
        )}
      </Drawer>

    </div>
  );
};

export default SiteManagement; 