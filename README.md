# 智能网络爬虫管理系统

一个基于Go语言和React技术栈的企业级Web爬虫管理平台，支持可视化配置、实时监控、数据管理等功能。

## 📋 目录

- [系统概述](#系统概述)
- [功能特性](#功能特性)
- [技术架构](#技术架构)
- [快速开始](#快速开始)
- [详细部署](#详细部署)
- [API文档](#api文档)
- [使用指南](#使用指南)
- [配置说明](#配置说明)
- [故障排除](#故障排除)

## 🚀 系统概述

本系统是一个功能强大的Web爬虫管理平台，提供了完整的爬虫生命周期管理功能：

- **可视化配置**: 通过Web界面轻松配置爬虫站点和规则
- **实时监控**: 实时查看爬虫运行状态和进度
- **数据管理**: 统一管理抓取到的数据，支持多种导出格式
- **任务调度**: 灵活的任务调度和管理
- **系统监控**: 全面的系统性能监控和日志管理

## ✨ 功能特性

### 核心功能

- 🕷️ **智能爬虫引擎**: 基于Colly的高性能爬虫框架
- 🌐 **Web管理界面**: 现代化的React前端界面
- 🗄️ **多数据源支持**: 支持MySQL、SQLite等数据库存储
- 📊 **实时监控面板**: ECharts图表展示数据统计
- 🔧 **灵活配置**: 支持YAML配置文件和Web界面配置

### 高级特性

- 🔄 **任务调度**: 支持定时任务和手动触发
- 📈 **数据可视化**: 丰富的图表和统计分析
- 🛡️ **反爬虫机制**: 支持代理、用户代理轮换等
- 💾 **数据导出**: 支持JSON、CSV、Excel等格式导出
- 🔍 **全文搜索**: 支持对抓取数据的全文搜索

## 🏗️ 技术架构

### 后端技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **爬虫引擎**: Colly v2
- **数据库**: MySQL 8.0+, SQLite
- **日志**: 结构化日志支持
- **配置**: YAML配置文件

### 前端技术栈

- **框架**: React 18+ (TypeScript)
- **UI库**: Ant Design
- **图表**: ECharts
- **路由**: React Router
- **HTTP客户端**: Axios

### 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web 前端      │    │   API 服务器    │    │   爬虫引擎      │
│   (React)       │────│   (Gin)         │────│   (Colly)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                       ┌─────────────────┐
                       │   数据存储      │
                       │   (MySQL)       │
                       └─────────────────┘
```

## 🚀 快速开始

### 环境要求

- Go 1.21 或更高版本
- Node.js 16+ 和 npm (用于前端开发)
- MySQL 8.0+ 或 SQLite
- Git

### 一键启动

```bash
# 1. 克隆项目
git clone <repository-url>
cd info-aggeration

# 2. 配置数据库连接
# 编辑 config/config.yaml 中的数据库配置

# 3. 一键启动
chmod +x scripts/start-web.sh
./scripts/start-web.sh
```

访问 http://localhost:8080 即可使用Web管理界面。

## 📚 详细部署

### 1. 环境准备

#### 安装Go语言环境

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# CentOS/RHEL
sudo yum install golang

# macOS (使用Homebrew)
brew install go
```

#### 安装Node.js环境

```bash
# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# CentOS/RHEL
curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
sudo yum install -y nodejs

# macOS (使用Homebrew)
brew install node
```

#### 安装MySQL数据库

```bash
# Ubuntu/Debian
sudo apt install mysql-server

# CentOS/RHEL
sudo yum install mysql-server

# macOS (使用Homebrew)
brew install mysql
```

### 2. 项目构建

#### 后端构建

```bash
# 下载依赖
go mod download

# 构建Web服务器
go build -o bin/webserver ./cmd/webserver

# 构建爬虫程序
go build -o bin/crawler ./cmd/crawler
```

#### 前端构建

```bash
cd web

# 安装依赖
npm install

# 构建生产版本
npm run build

cd ..
```

### 3. 数据库配置

#### MySQL配置

1. 创建数据库：

```sql
CREATE DATABASE crawler_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'crawler_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON crawler_db.* TO 'crawler_user'@'localhost';
FLUSH PRIVILEGES;
```

2. 更新配置文件 `config/config.yaml`:

```yaml
storage:
  database:
    driver: "mysql"
    host: "localhost"
    port: 3306
    username: "crawler_user"
    password: "your_password"
    database: "crawler_db"
```

### 4. 启动服务

#### 启动Web服务器

```bash
./bin/webserver --config config/config.yaml
```

#### 使用Docker部署

```bash
# 构建镜像
docker build -t crawler-web .

# 运行容器
docker run -d -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/data:/app/data \
  crawler-web
```

## 📖 API文档

### 站点管理API

#### 获取站点列表

```http
GET /api/v1/sites?page=1&page_size=10
```

#### 创建站点

```http
POST /api/v1/sites
Content-Type: application/json

{
  "name": "示例站点",
  "base_url": "https://example.com",
  "description": "站点描述",
  "start_urls": ["https://example.com/page1"],
  "selectors": {
    "title": "h1",
    "content": ".content"
  },
  "enabled": true
}
```

#### 更新站点

```http
PUT /api/v1/sites/{id}
Content-Type: application/json

{
  "name": "更新的站点名称",
  "enabled": false
}
```

### 任务管理API

#### 获取任务列表

```http
GET /api/v1/tasks?page=1&page_size=10&status=running
```

#### 创建任务

```http
POST /api/v1/tasks
Content-Type: application/json

{
  "name": "测试任务",
  "site_id": 1,
  "config": {
    "max_pages": 100,
    "concurrent": 3
  }
}
```

#### 启动任务

```http
POST /api/v1/tasks/{id}/start
```

#### 停止任务

```http
POST /api/v1/tasks/{id}/stop
```

### 数据管理API

#### 获取数据列表

```http
GET /api/v1/data/items?page=1&page_size=10&site_id=1
```

#### 搜索数据

```http
POST /api/v1/data/items/search
Content-Type: application/json

{
  "keyword": "搜索关键词",
  "site_id": 1,
  "start_date": "2023-01-01",
  "end_date": "2023-12-31"
}
```

#### 导出数据

```http
GET /api/v1/data/items/export?format=json&site_id=1
```

## 🎯 使用指南

### 1. 创建爬虫站点

1. 访问 Web 管理界面：http://localhost:8080
2. 点击左侧菜单"站点管理"
3. 点击"新增站点"按钮
4. 填写站点基本信息：
   - **站点名称**: 为站点起一个便于识别的名称
   - **基础URL**: 站点的根URL
   - **描述**: 站点的详细描述
5. 配置起始URL：
   - 每行填写一个URL
   - 这些URL将作为爬虫的入口点
6. 配置选择器：
   - 使用JSON格式配置CSS选择器
   - 常用选择器示例：
   ```json
   {
     "title": "h1, .title",
     "content": ".content, .article-body",
     "author": ".author, .byline",
     "date": ".date, .publish-time",
     "category": ".category, .tag"
   }
   ```

### 2. 创建爬虫任务

1. 进入"任务管理"页面
2. 点击"新增任务"
3. 选择要爬取的站点
4. 配置任务参数：
   - **并发数**: 同时运行的爬虫数量
   - **最大页面数**: 限制爬取的页面总数
   - **深度限制**: 限制爬取的链接深度
5. 启动任务并监控进度

### 3. 监控和管理

#### 实时监控

- **仪表盘**: 查看系统整体状态和统计信息
- **任务状态**: 实时查看任务执行进度
- **系统资源**: 监控CPU、内存使用情况

#### 日志查看

- **任务日志**: 查看特定任务的详细日志
- **系统日志**: 查看系统级别的日志信息
- **错误日志**: 快速定位和解决问题

### 4. 数据管理

#### 查看和搜索数据

- **数据列表**: 分页浏览抓取到的数据
- **全文搜索**: 在标题和内容中搜索关键词
- **过滤功能**: 按站点、时间范围等条件过滤

#### 数据导出

- 支持JSON、CSV、Excel格式
- 可按站点或时间范围导出
- 大量数据自动分批处理

## ⚙️ 配置说明

### 爬虫配置

```yaml
spider:
  concurrent: 5          # 并发数
  delay: 1000           # 请求间隔（毫秒）
  timeout: 30           # 超时时间（秒）
  retries: 3            # 重试次数
  user_agent: "..."     # 用户代理
  max_depth: 10         # 最大深度
  max_pages: 1000       # 最大页面数
```

### 存储配置

```yaml
storage:
  type: "database"      # 存储类型
  database:
    driver: "mysql"     # 数据库驱动
    host: "localhost"   # 主机地址
    port: 3306          # 端口
    username: "root"    # 用户名
    password: "password" # 密码
    database: "crawler_db" # 数据库名
```

### Web服务器配置

```yaml
web:
  port: 8080           # 服务器端口
  static_path: "./web/build" # 静态文件路径
  api_prefix: "/api/v1" # API前缀
  cors:
    origins:           # 允许的来源
      - "http://localhost:3000"
    methods:           # 允许的方法
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
```

## 🔧 故障排除

### 常见问题

#### 1. 数据库连接失败

**问题**: 启动时提示数据库连接失败

**解决方案**:
- 检查MySQL服务是否启动
- 验证配置文件中的数据库连接信息
- 确保数据库用户具有足够权限

```bash
# 检查MySQL状态
sudo systemctl status mysql

# 重启MySQL
sudo systemctl restart mysql
```

#### 2. 前端页面无法访问

**问题**: 访问Web界面显示404错误

**解决方案**:
- 确保前端已正确构建
- 检查静态文件路径配置
- 验证Web服务器是否正常启动

```bash
# 重新构建前端
cd web
npm run build

# 检查构建文件
ls -la build/
```

#### 3. 爬虫无法抓取数据

**问题**: 爬虫任务运行但没有抓取到数据

**解决方案**:
- 检查目标网站是否有反爬虫机制
- 验证CSS选择器是否正确
- 调整并发数和请求间隔

#### 4. 内存使用过高

**问题**: 爬虫运行时内存占用过大

**解决方案**:
- 减少并发数
- 增加请求间隔
- 限制最大页面数
- 定期重启爬虫进程

### 性能优化

#### 1. 数据库优化

```sql
-- 添加索引
CREATE INDEX idx_crawl_time ON crawl_data(crawl_time);
CREATE INDEX idx_site_id ON crawl_data(site_id);
CREATE INDEX idx_url ON crawl_data(url(255));

-- 优化表结构
ALTER TABLE crawl_data ADD INDEX idx_status (status);
```

#### 2. 爬虫优化

- 合理设置并发数（建议不超过10）
- 适当增加请求间隔（避免被封IP）
- 使用代理池分散请求
- 定期清理过期数据

#### 3. 系统优化

- 定期备份数据库
- 监控磁盘空间使用
- 设置日志轮转
- 配置系统监控告警

### 日志分析

#### 查看系统日志

```bash
# 查看Web服务器日志
tail -f data/logs/webserver.log

# 查看爬虫日志
tail -f data/logs/crawler.log

# 查看错误日志
grep ERROR data/logs/*.log
```

#### 性能监控

```bash
# 监控进程资源使用
ps aux | grep crawler

# 监控网络连接
netstat -an | grep :8080

# 监控磁盘使用
df -h
du -sh data/
```

## 📞 技术支持

如果您在使用过程中遇到问题，可以通过以下方式获取帮助：

- 📧 邮箱支持: support@example.com
- 📚 在线文档: https://docs.example.com
- 🐛 问题反馈: https://github.com/example/issues

## 📄 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

---

**最后更新**: 2024年1月 