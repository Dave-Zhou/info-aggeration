import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Table, Tag, Progress } from 'antd';
import { 
  DatabaseOutlined, 
  GlobalOutlined, 
  PlayCircleOutlined, 
  CheckCircleOutlined 
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import { getStatistics, getRecentTasks } from '../services/api';

const Dashboard: React.FC = () => {
  const [statistics, setStatistics] = useState<any>({});
  const [recentTasks, setRecentTasks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [statsRes, tasksRes] = await Promise.all([
        getStatistics(),
        getRecentTasks()
      ]);
      setStatistics(statsRes.data || {});
      setRecentTasks(tasksRes.data || []);
    } catch (error) {
      console.error('获取数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const taskColumns = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '站点',
      dataIndex: 'site_name',
      key: 'site_name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusColors = {
          'running': 'processing',
          'completed': 'success',
          'failed': 'error',
          'pending': 'default',
          'stopped': 'warning'
        };
        return <Tag color={statusColors[status as keyof typeof statusColors]}>{status}</Tag>;
      },
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      render: (progress: number) => (
        <Progress percent={Math.round(progress)} size="small" />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => new Date(time).toLocaleString(),
    },
  ];

  const chartOption = {
    title: {
      text: '最近7天抓取数据统计',
      left: 'center'
    },
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: statistics.recent_stats?.map((item: any) => item.date) || []
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '抓取数量',
        type: 'line',
        data: statistics.recent_stats?.map((item: any) => item.count) || [],
        smooth: true,
        areaStyle: {}
      }
    ]
  };

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总数据量"
              value={statistics.total_items || 0}
              prefix={<DatabaseOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃站点"
              value={statistics.total_sites || 0}
              prefix={<GlobalOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="运行任务"
              value={statistics.total_tasks || 0}
              prefix={<PlayCircleOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="今日抓取"
              value={statistics.today_items || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#eb2f96' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="数据趋势" loading={loading}>
            <ReactECharts option={chartOption} style={{ height: 300 }} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="站点统计" loading={loading}>
            <div style={{ height: 300, overflow: 'auto' }}>
              {statistics.site_stats?.map((item: any, index: number) => (
                <div key={index} style={{ marginBottom: 12 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <span>{item.name}</span>
                    <span style={{ fontWeight: 'bold' }}>{item.count}</span>
                  </div>
                  <Progress 
                    percent={item.count / Math.max(...statistics.site_stats?.map((s: any) => s.count) || [1]) * 100} 
                    showInfo={false}
                    size="small"
                  />
                </div>
              ))}
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card title="最近任务" loading={loading}>
            <Table
              dataSource={recentTasks}
              columns={taskColumns}
              pagination={{ pageSize: 10 }}
              size="small"
              rowKey="id"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard; 