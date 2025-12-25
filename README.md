# 图床系统

一个前后端分离的图床系统，支持图片上传、管理和访问。

## 技术栈

- 后端: Go 1.22+ / Gin
- 前端: Vue3 / Vite
- 存储: 本地文件系统 (可扩展至 S3/OSS)

## 项目结构

```
image-hosting/
├── backend/                 # 后端服务
│   ├── main.go             # 入口文件
│   ├── config.yaml         # 配置文件
│   └── internal/
│       ├── config/         # 配置管理
│       ├── handler/        # HTTP 处理器
│       ├── middleware/     # 中间件
│       ├── model/          # 数据模型
│       ├── service/        # 业务逻辑
│       └── storage/        # 存储抽象
└── frontend/               # 前端应用
    ├── src/
    │   ├── api/            # API 封装
    │   ├── router/         # 路由配置
    │   ├── views/          # 页面组件
    │   └── styles/         # 样式文件
    └── vite.config.js
```

## 快速开始

### 后端

```bash
cd backend

# 安装依赖
go mod tidy

# 运行
go run main.go
```

服务默认运行在 `http://localhost:8080`

### 前端

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev
```

前端默认运行在 `http://localhost:3000`

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/upload | 上传图片 |
| GET | /api/v1/images | 获取图片列表 |
| GET | /api/v1/image/:id | 获取图片详情 |
| DELETE | /api/v1/image/:id | 删除图片 |

### 响应格式

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

## 配置说明

编辑 `backend/config.yaml`:

```yaml
server:
  port: "8080"

storage:
  type: "local"
  base_path: "./storage/images"
  base_url: "/images"

auth:
  enabled: false
  tokens:
    - "your-token"

image:
  quality: 75
  max_size: 10485760
```

## 鉴权

启用鉴权后，请求需携带 Token:

```
Authorization: Bearer your-token
```

## License

MIT
