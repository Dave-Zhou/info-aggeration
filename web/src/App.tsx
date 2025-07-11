import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { Layout } from 'antd';
import MainLayout from './components/Layout/MainLayout';
import Dashboard from './pages/Dashboard';
import SiteManagement from './pages/SiteManagement';
import TaskManagement from './pages/TaskManagement';
import DataManagement from './pages/DataManagement';
import SystemSettings from './pages/SystemSettings';
import './App.css';

const { Content } = Layout;

function App() {
  return (
    <div className="App">
      <MainLayout>
        <Content style={{ padding: '24px', minHeight: '100vh' }}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/sites" element={<SiteManagement />} />
            <Route path="/tasks" element={<TaskManagement />} />
            <Route path="/data" element={<DataManagement />} />
            <Route path="/settings" element={<SystemSettings />} />
          </Routes>
        </Content>
      </MainLayout>
    </div>
  );
}

export default App; 