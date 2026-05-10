# Inner Auth

统一的内部服务鉴权网关，提供认证页面和反向代理功能。

## 功能

- 用户名 + 密码 + 可选 TOTP 认证
- JWT Cookie 会话管理
- 反向代理到上游服务
- IP 级别和全局级别速率限制
- 自动切换浅色/深色主题
- 中英文自动识别
- Bitwarden 等密码管理器友好

## 安装

```bash
go build -o inner-auth
```

## 使用

```bash
# 使用默认配置文件 ./config.json
./inner-auth

# 指定配置文件
./inner-auth -c /path/to/config.json

# 查看版本
./inner-auth -v

# 显示 TOTP 导入链接
./inner-auth -show-totp
```

## 配置

复制 `config.json.template` 为 `config.json` 并修改：

```json
{
  "listen_port": 6904,
  "upstream": "http://127.0.0.1:4096",
  "title": "My Service Login",
  "jwt_secret": "change-this-to-a-random-secret-key",
  "session_ttl_hours": 168,
  "rate_limit": {
    "max_attempts_per_ip": 5,
    "ip_window_seconds": 60,
    "ip_lockout_seconds": 300,
    "global_max_attempts": 200,
    "global_window_seconds": 3600
  },
  "auth": {
    "user": "your_username",
    "password": "your_password",
    "totp_token": ""
  }
}
```

### 配置项说明

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `listen_port` | int | 6904 | 监听端口 |
| `upstream` | string | - | 上游服务地址（必填） |
| `title` | string | Login | 登录页面标题 |
| `jwt_secret` | string | - | JWT 签名密钥（必填） |
| `session_ttl_hours` | int | 168 | 会话时长（小时），默认 7 天 |
| `rate_limit` | object | - | 速率限制配置 |
| `auth` | object | - | 认证信息（必填） |

#### rate_limit

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `max_attempts_per_ip` | int | 5 | 单 IP 最大尝试次数 |
| `ip_window_seconds` | int | 60 | IP 限制窗口期（秒） |
| `ip_lockout_seconds` | int | 300 | IP 锁定时长（秒） |
| `global_max_attempts` | int | 200 | 全局最大尝试次数 |
| `global_window_seconds` | int | 3600 | 全局限制窗口期（秒） |

#### auth

| 字段 | 类型 | 说明 |
|------|------|------|
| `user` | string | 用户名（必填） |
| `password` | string | 密码（必填） |
| `totp_token` | string | TOTP 密钥（base32 编码），为空时跳过 TOTP 验证 |

## TOTP 设置

1. 生成 TOTP 密钥并配置到 `config.json`
2. 获取导入链接：
   ```bash
   ./inner-auth -show-totp
   ```
3. 将链接导入到 Authenticator App 或 Bitwarden

## 路由

| 路径 | 说明 |
|------|------|
| `/inner-login` | 登录页面 |
| `/inner-logout` | 退出登录 |
| 其他路径 | 需要认证，认证通过后转发到上游服务 |

## 安全特性

- **Cookie 安全**：HttpOnly，防止 XSS 攻击
- **速率限制**：IP 级别（5次/分钟）+ 全局级别（200次/小时）
- **JWT 过期**：默认 7 天自动过期
- **TOTP**：SHA1 算法，6 位验证码，30 秒周期，±30 秒容错

## 许可证

MIT
