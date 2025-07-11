import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Button, 
  Space, 
  Modal, 
  Form, 
  Input, 
  Switch, 
  Tag,
  message,
  Popconfirm,
  Drawer
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  PlayCircleOutlined,
  StopOutlined,
  EyeOutlined
} from '@ant-design/icons';
import { getSites, createSite, updateSite, deleteSite, toggleSite } from '../services/api';

const { TextArea } = Input;

const SiteManagement: React.FC = () => {
  const [sites, setSites] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [editingSite, setEditingSite] = useState<any>(null);
  const [viewingSite, setViewingSite] = useState<any>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchSites();
  }, []);

  const fetchSites = async () => {
    try {
      setLoading(true);
      const response = await getSites();
      setSites(response.data || []);
    } catch (error) {
      message.error('获取站点列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingSite(null);
    setModalVisible(true);
    form.resetFields();
  };

  const handleEdit = (site: any) => {
    setEditingSite(site);
    setModalVisible(true);
    form.setFieldsValue({
      ...site,
      start_urls: site.start_urls?.join('\n') || '',
      selectors: JSON.stringify(site.selectors || {}, null, 2)
    });
  };

  const handleView = (site: any) => {
    setViewingSite(site);
    setDrawerVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteSite(id);
      message.success('删除成功');
      fetchSites();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleToggle = async (id: number) => {
    try {
      await toggleSite(id);
      message.success('状态更新成功');
      fetchSites();
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      const submitData = {
        ...values,
        start_urls: values.start_urls.split('\n').filter((url: string) => url.trim()),
        selectors: JSON.parse(values.selectors || '{}')
      };

      if (editingSite) {
        await updateSite(editingSite.id, submitData);
        message.success('更新成功');
      } else {
        await createSite(submitData);
        message.success('创建成功');
      }
      
      setModalVisible(false);
      fetchSites();
    } catch (error) {
      message.error('操作失败');
    }
  };

  const columns = [
    {
      title: '站点名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: '基础URL',
      dataIndex: 'base_url',
      key: 'base_url',
      width: 200,
      render: (url: string) => (
        <a href={url} target="_blank" rel="noopener noreferrer">
          {url}
        </a>
      )
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      width: 200,
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 80,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? '启用' : '禁用'}
        </Tag>
      )
    },
    {
      title: '运行状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusColors = {
          'ready': 'default',
          'running': 'processing',
          'stopped': 'warning',
          'error': 'error'
        };
        return <Tag color={statusColors[status as keyof typeof statusColors]}>{status}</Tag>;
      }
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (time: string) => new Date(time).toLocaleString()
    },
    {
      title: '操作',
      key: 'action',
      width: 250,
      render: (record: any) => (
        <Space size="small">
          <Button 
            type="link" 
            icon={<EyeOutlined />}
            onClick={() => handleView(record)}
          >
            查看
          </Button>
          <Button 
            type="link" 
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button 
            type="link" 
            icon={record.enabled ? <StopOutlined /> : <PlayCircleOutlined />}
            onClick={() => handleToggle(record.id)}
          >
            {record.enabled ? '禁用' : '启用'}
          </Button>
          <Popconfirm
            title="确定要删除这个站点吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button 
              type="link" 
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  return (
    <div>
      <Card>
        <div style={{ marginBottom: 16 }}>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            新增站点
          </Button>
        </div>
        
        <Table
          columns={columns}
          dataSource={sites}
          loading={loading}
          rowKey="id"
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      <Modal
        title={editingSite ? '编辑站点' : '新增站点'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={800}
        okText="保存"
        cancelText="取消"
      >
        <Form form={form} layout="vertical">
          <Form.Item 
            label="站点名称" 
            name="name"
            rules={[{ required: true, message: '请输入站点名称' }]}
          >
            <Input placeholder="请输入站点名称" />
          </Form.Item>
          
          <Form.Item 
            label="基础URL" 
            name="base_url"
            rules={[{ required: true, message: '请输入基础URL' }]}
          >
            <Input placeholder="https://example.com" />
          </Form.Item>
          
          <Form.Item label="描述" name="description">
            <TextArea placeholder="站点描述" rows={2} />
          </Form.Item>
          
          <Form.Item 
            label="起始URL" 
            name="start_urls"
            rules={[{ required: true, message: '请输入起始URL' }]}
          >
            <TextArea 
              placeholder="每行一个URL，例如：&#10;https://example.com/page1&#10;https://example.com/page2"
              rows={4}
            />
          </Form.Item>
          
          <Form.Item 
            label="选择器配置" 
            name="selectors"
            rules={[{ required: true, message: '请输入选择器配置' }]}
          >
            <TextArea 
              placeholder='JSON格式，例如：&#10;{&#10;  "title": "h1",&#10;  "content": ".content",&#10;  "author": ".author"&#10;}'
              rows={8}
            />
          </Form.Item>
          
          <Form.Item label="启用状态" name="enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title="站点详情"
        placement="right"
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
        width={600}
      >
        {viewingSite && (
          <div>
            <h3>基本信息</h3>
            <p><strong>名称:</strong> {viewingSite.name}</p>
            <p><strong>基础URL:</strong> {viewingSite.base_url}</p>
            <p><strong>描述:</strong> {viewingSite.description}</p>
            <p><strong>状态:</strong> {viewingSite.enabled ? '启用' : '禁用'}</p>
            
            <h3>起始URL</h3>
            <ul>
              {viewingSite.start_urls?.map((url: string, index: number) => (
                <li key={index}>{url}</li>
              ))}
            </ul>
            
            <h3>选择器配置</h3>
            <pre style={{ background: '#f5f5f5', padding: '10px', borderRadius: '4px' }}>
              {JSON.stringify(viewingSite.selectors || {}, null, 2)}
            </pre>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default SiteManagement; 