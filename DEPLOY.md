# 宝塔面板部署指南

## 前置准备

### 1. 安装 Go 环境

在宝塔面板的「软件商店」中搜索并安装 Go 语言环境，或通过 SSH 手动安装：

```bash
# 下载 Go (以 1.22 为例)
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz

# 解压到 /usr/local
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# 配置环境变量 (添加到 /etc/profile)
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
echo 'export GOPATH=/root/go' >> /etc/profile
source /etc/profile

# 验证安装
go version
```

### 2. 安装 Node.js

在宝塔「软件商店」→「运行环境」中安装 Node.js (建议 18+)。

---

## 部署步骤

### 步骤一：上传项目

将项目上传到服务器，建议放在 `/www/wwwroot/image-hosting/`

```
/www/wwwroot/image-hosting/
├── backend/
└── frontend/
```

### 步骤二：编译后端

```bash
cd /www/wwwroot/image-hosting/backend

# 下载依赖
go mod tidy

# 编译为可执行文件
go build -o image-hosting main.go

# 创建存储目录
mkdir -p storage/images

# 修改配置文件
vim config.yaml
```

修改 `config.yaml`：

```yaml
server:
  host: "127.0.0.1"
  port: "8080"

storage:
  type: "local"
  base_path: "/www/wwwroot/image-hosting/backend/storage/images"
  base_url: "/images"

auth:
  enabled: true                    # 生产环境建议开启
  tokens:
    - "your-secure-token-here"     # 改成你自己的 Token

image:
  quality: 75
  max_size: 10485760
  allowed_types:
    - "image/jpeg"
    - "image/png"
    - "image/webp"
```

### 步骤三：配置 Supervisor 守护进程

在宝塔面板中：「软件商店」→ 搜索「Supervisor」→ 安装

安装后点击「设置」→「添加守护进程」：

| 配置项 | 值 |
|--------|-----|
| 名称 | image-hosting |
| 启动命令 | /www/wwwroot/image-hosting/backend/image-hosting |
| 运行目录 | /www/wwwroot/image-hosting/backend |
| 启动用户 | root |

保存后启动进程。

### 步骤四：构建前端

```bash
cd /www/wwwroot/image-hosting/frontend

# 安装依赖
npm install

# 构建生产版本
npm run build
```

构建完成后会生成 `dist` 目录。

### 步骤五：配置 Nginx

在宝塔面板中创建网站：

1. 「网站」→「添加站点」
2. 域名填写你的域名（如 `img.example.com`）
3. 根目录设置为 `/www/wwwroot/image-hosting/frontend/dist`
4. PHP 版本选择「纯静态」

然后修改 Nginx 配置（点击网站「设置」→「配置文件」）：

```nginx
server {
    listen 80;
    server_name img.example.com;  # 改成你的域名
    
    root /www/wwwroot/image-hosting/frontend/dist;
    index index.html;

    # 前端路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 代理到后端
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # 上传文件大小限制
        client_max_body_size 20m;
    }

    # 图片静态文件
    location /images/ {
        alias /www/wwwroot/image-hosting/backend/storage/images/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # 日志
    access_log /www/wwwlogs/image-hosting.log;
    error_log /www/wwwlogs/image-hosting.error.log;
}
```

保存后重载 Nginx。

### 步骤六：配置 SSL（可选但推荐）

在宝塔网站设置中点击「SSL」→「Let's Encrypt」申请免费证书。

---

## 验证部署

1. 访问 `https://img.example.com` 应该看到上传页面
2. 上传一张图片测试
3. 检查图片是否保存到 `/www/wwwroot/image-hosting/backend/storage/images/` 目录

---

## 常见问题

### 上传失败 413 错误

Nginx 默认限制上传大小，确保配置了 `client_max_body_size`：

```nginx
client_max_body_size 20m;
```

### 图片无法访问

检查存储目录权限：

```bash
chmod -R 755 /www/wwwroot/image-hosting/backend/storage
chown -R www:www /www/wwwroot/image-hosting/backend/storage
```

### 后端服务未启动

检查 Supervisor 状态：

```bash
supervisorctl status image-hosting
supervisorctl restart image-hosting
```

查看日志：

```bash
tail -f /www/server/panel/plugin/supervisor/log/image-hosting.out.log
```

### API 返回 401 错误

如果开启了鉴权，前端需要配置 Token。编辑 `frontend/src/api/index.js`，在构建前设置默认 Token，或在浏览器 localStorage 中设置：

```javascript
localStorage.setItem('api_token', 'your-secure-token-here')
```

---

## 目录结构（部署后）

```
/www/wwwroot/image-hosting/
├── backend/
│   ├── image-hosting          # 编译后的可执行文件
│   ├── config.yaml            # 配置文件
│   └── storage/
│       └── images/            # 图片存储目录
│           ├── metadata.json  # 元数据
│           └── 2024/12/       # 按年月存储的图片
└── frontend/
    └── dist/                  # 前端构建产物
```

---

## 备份建议

定期备份以下内容：

1. `backend/storage/images/` - 所有上传的图片
2. `backend/storage/images/metadata.json` - 图片元数据
3. `backend/config.yaml` - 配置文件
