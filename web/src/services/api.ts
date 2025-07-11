import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

// 创建axios实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 可以在这里添加认证token
    // const token = localStorage.getItem('token');
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`;
    // }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

// 站点管理API
export const getSites = (params?: any) => api.get('/sites', { params });
export const getSite = (id: number) => api.get(`/sites/${id}`);
export const createSite = (data: any) => api.post('/sites', data);
export const updateSite = (id: number, data: any) => api.put(`/sites/${id}`, data);
export const deleteSite = (id: number) => api.delete(`/sites/${id}`);
export const testSite = (id: number) => api.post(`/sites/${id}/test`);
export const toggleSite = (id: number) => api.put(`/sites/${id}/toggle`);
export const runSiteTask = (id: number) => api.post(`/sites/${id}/run`);

// 任务管理API
export const getTasks = (params?: any) => api.get('/tasks', { params });
export const getTask = (id: number) => api.get(`/tasks/${id}`);
export const createTask = (data: any) => api.post('/tasks', data);
export const updateTask = (id: number, data: any) => api.put(`/tasks/${id}`, data);
export const deleteTask = (id: number) => api.delete(`/tasks/${id}`);
export const startTask = (id: number) => api.post(`/tasks/${id}/start`);
export const stopTask = (id: number) => api.post(`/tasks/${id}/stop`);
export const getTaskLogs = (id: number, params?: any) => api.get(`/tasks/${id}/logs`, { params });
export const getTaskStatus = (id: number) => api.get(`/tasks/${id}/status`);

// 数据管理API
export const getItems = (params?: any) => api.get('/data/items', { params });
export const getItem = (id: number) => api.get(`/data/items/${id}`);
export const deleteItem = (id: number) => api.delete(`/data/items/${id}`);
export const searchItems = (data: any) => api.post('/data/items/search', data);
export const exportItems = (params?: any) => api.get('/data/items/export', { params });
export const getStatistics = () => api.get('/data/statistics');

// 系统管理API
export const getSystemStatus = () => api.get('/system/status');
export const getSystemConfig = () => api.get('/system/config');
export const updateSystemConfig = (data: any) => api.put('/system/config', data);
export const getSystemLogs = (params?: any) => api.get('/system/logs', { params });
export const createBackup = () => api.post('/system/backup');
export const getBackups = () => api.get('/system/backups');
export const restoreBackup = (data: any) => api.post('/system/restore', data);

// 辅助函数
export const getRecentTasks = (params?: any) => getTasks({ ...params, page_size: 10 });

export default api; 