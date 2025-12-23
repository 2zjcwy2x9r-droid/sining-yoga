# 微信小程序连接后端配置指南

## 一、开发环境配置

### 1. 修改小程序API地址

编辑 `miniprogram/app.js`，修改 `apiBaseUrl`：

```javascript
globalData: {
  userInfo: null,
  // 开发环境：使用本地IP地址（不能使用localhost）
  // 获取本机IP：ifconfig (Linux/Mac) 或 ipconfig (Windows)
  apiBaseUrl: 'http://192.168.x.x:8080/api/v1'  // 替换为你的实际IP
}
```

**重要提示**：
- 微信小程序不能使用 `localhost` 或 `127.0.0.1`
- 必须使用本机的局域网IP地址（如 `192.168.1.100`）
- 确保手机和电脑在同一局域网

### 2. 获取本机IP地址

**Linux/Mac:**
```bash
ifconfig | grep "inet " | grep -v 127.0.0.1
# 或
ip addr show | grep "inet " | grep -v 127.0.0.1
```

**Windows:**
```bash
ipconfig
# 查找 IPv4 地址，通常是 192.168.x.x
```

### 3. 配置微信开发者工具

1. 打开微信开发者工具
2. 点击右上角"详情"
3. 在"本地设置"中：
   - ✅ 勾选"不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书"
   - ✅ 勾选"不校验安全域名、TLS版本以及HTTPS证书"

### 4. 确保后端服务可访问

```bash
# 检查后端服务是否运行
curl http://localhost:8080/health

# 检查防火墙（确保8080端口开放）
# Linux:
sudo ufw allow 8080
# 或
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

## 二、生产环境配置

### 1. 配置HTTPS域名

小程序生产环境必须使用HTTPS，需要：

1. **购买域名**（如：`api.yourdomain.com`）
2. **配置SSL证书**（可以使用Let's Encrypt免费证书）
3. **配置反向代理**（Nginx推荐）

### 2. Nginx配置示例

```nginx
server {
    listen 443 ssl;
    server_name api.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 3. 修改小程序API地址

```javascript
globalData: {
  userInfo: null,
  // 生产环境：使用HTTPS域名
  apiBaseUrl: 'https://api.yourdomain.com/api/v1'
}
```

### 4. 配置微信小程序后台

1. 登录[微信公众平台](https://mp.weixin.qq.com/)
2. 进入"开发" -> "开发管理" -> "开发设置"
3. 在"服务器域名"中配置：
   - **request合法域名**：添加 `https://api.yourdomain.com`
   - **uploadFile合法域名**：添加 `https://api.yourdomain.com`
   - **downloadFile合法域名**：添加 `https://api.yourdomain.com`

## 三、快速测试

### 1. 开发环境测试步骤

```bash
# 1. 获取本机IP
hostname -I | awk '{print $1}'
# 或
ip route get 1 | awk '{print $7;exit}'

# 2. 修改 miniprogram/app.js 中的 apiBaseUrl

# 3. 在微信开发者工具中测试
# - 打开小程序
# - 进入"AI问答"页面
# - 发送测试消息
```

### 2. 检查网络连接

在微信开发者工具的"调试器" -> "Network"中查看请求：
- 请求URL是否正确
- 请求是否成功（状态码200）
- 响应数据是否正确

## 四、常见问题

### Q1: 提示"不在以下 request 合法域名列表中"

**解决方案**：
- 开发环境：在微信开发者工具中勾选"不校验合法域名"
- 生产环境：在微信公众平台配置服务器域名

### Q2: 提示"网络请求失败"

**检查清单**：
1. ✅ 后端服务是否运行：`curl http://localhost:8080/health`
2. ✅ IP地址是否正确（不能使用localhost）
3. ✅ 防火墙是否开放8080端口
4. ✅ 手机和电脑是否在同一WiFi网络
5. ✅ 微信开发者工具网络设置是否正确

### Q3: 跨域问题

后端已配置CORS，如果仍有问题，检查：
- 后端是否添加了CORS中间件
- 请求头是否正确

## 五、API接口说明

小程序主要使用的API：

1. **AI聊天**
   - `POST /api/v1/ai/chat`
   - 请求体：`{ message, history, base_id }`

2. **知识库列表**
   - `GET /api/v1/knowledge-bases`

3. **知识项列表**
   - `GET /api/v1/knowledge-bases/:base_id/items`

详细API文档见 `README.md`

## 六、配置示例

### 开发环境配置（app.js）

```javascript
App({
  onLaunch() {
    // ...
  },
  globalData: {
    userInfo: null,
    // 开发环境：使用本机IP
    apiBaseUrl: 'http://192.168.1.100:8080/api/v1'
  }
})
```

### 生产环境配置（app.js）

```javascript
App({
  onLaunch() {
    // ...
  },
  globalData: {
    userInfo: null,
    // 生产环境：使用HTTPS域名
    apiBaseUrl: 'https://api.yourdomain.com/api/v1'
  }
})
```

### 环境变量配置（推荐）

可以创建配置文件来区分环境：

```javascript
// config.js
const config = {
  development: {
    apiBaseUrl: 'http://192.168.1.100:8080/api/v1'
  },
  production: {
    apiBaseUrl: 'https://api.yourdomain.com/api/v1'
  }
}

const env = 'development' // 或 'production'
module.exports = config[env]
```

然后在 `app.js` 中引入：
```javascript
const config = require('./config')
globalData: {
  apiBaseUrl: config.apiBaseUrl
}
```

