import React, { useState } from 'react';
import { Layout, Menu, Button } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  DashboardOutlined,
  GlobalOutlined,
  PlayCircleOutlined,
  DatabaseOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons';

const { Header, Sider } = Layout;

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: '/sites',
      icon: <GlobalOutlined />,
      label: '站点管理',
    },
    {
      key: '/tasks',
      icon: <PlayCircleOutlined />,
      label: '任务管理',
    },
    {
      key: '/data',
      icon: <DatabaseOutlined />,
      label: '数据管理',
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
  ];

  const handleMenuClick = (e: any) => {
    navigate(e.key);
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed}>
        <div className="logo" style={{ 
          height: '64px', 
          padding: '16px', 
          color: 'white',
          textAlign: 'center',
          fontSize: '18px',
          fontWeight: 'bold'
        }}>
          {collapsed ? '爬虫' : '爬虫管理系统'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header style={{ 
          padding: '0 16px', 
          background: 'white',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between'
        }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            style={{ fontSize: '16px' }}
          />
          <div style={{ fontSize: '16px', fontWeight: 'bold' }}>
            智能网络爬虫管理平台
          </div>
        </Header>
        {children}
      </Layout>
    </Layout>
  );
};

export default MainLayout; 