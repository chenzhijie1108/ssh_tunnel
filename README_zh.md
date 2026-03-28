# SSH Tunnel Manager

[English](./README.md) | 中文

SSH 隧道管理工具，提供 Web 界面管理 SSH 隧道。

## 功能特性

- **Web 界面** - 简洁的网页管理界面
- **本地与远程隧道** - 支持 `-L`（本地）和 `-R`（远程）两种 SSH 隧道类型
- **RESTful API** - 完整的隧道增删改查 API
- **SSH 认证** - 支持 SSH 密钥和密码两种认证方式
- **自动重试** - 隧道启动失败时自动重试机制
- **进程持久化** - 重启后自动恢复隧道状态
- **跨平台** - 支持 Windows、macOS、Linux

## 快速开始

### 编译

```bash
go build -o ssh_tunnel_manager .
```

### 运行

```bash
./ssh_tunnel_manager
```

启动后程序会：
1. 在端口 `11108` 启动 HTTP 服务
2. 自动打开浏览器访问 `http://localhost:11108`
3. 从 `config.json` 加载已有隧道配置

## 配置说明

隧道配置存储在 `config.json` 中：

```json
{
  "tunnels": [
    {
      "name": "dev_db",
      "type": "local",
      "local_port": "33306",
      "remote_host": "localhost",
      "remote_port": "13306",
      "ssh_host": "155.248.183.71",
      "ssh_port": "22",
      "ssh_user": "ubuntu",
      "ssh_key": "/path/to/key.pem",
      "ssh_pass": ""
    }
  ]
}
```

### 字段说明

| 字段 | 说明 | 必填 |
|------|------|------|
| `name` | 隧道唯一标识 | 是 |
| `type` | 隧道类型：`local` 或 `remote` | 是 |
| `local_port` | 本地端口 | 是 |
| `remote_host` | 远程目标主机 | 是 |
| `remote_port` | 远程目标端口 | 是 |
| `ssh_host` | SSH 服务器地址 | 是 |
| `ssh_port` | SSH 服务器端口（默认：22） | 否 |
| `ssh_user` | SSH 用户名 | 是 |
| `ssh_key` | SSH 私钥路径 | 是* |
| `ssh_pass` | SSH 密码 | 是* |

*`ssh_key` 和 `ssh_pass` 二选一提供。

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/tunnels` | 获取所有隧道列表 |
| POST | `/api/tunnels` | 创建新隧道 |
| GET | `/api/tunnels/{id}` | 获取隧道详情 |
| PUT | `/api/tunnels/{id}` | 更新隧道配置 |
| DELETE | `/api/tunnels/{id}` | 删除隧道 |
| POST | `/api/tunnels/{id}/start` | 启动隧道 |
| POST | `/api/tunnels/{id}/stop` | 停止隧道 |
| GET | `/api/tunnels/{id}/status` | 获取隧道状态 |
| GET | `/api/tunnels/{id}/stats` | 获取隧道统计信息 |
| GET | `/api/ping` | 健康检查 |

## 隧道类型

### 本地隧道 (`-L`)

将本地端口的流量通过 SSH 服务器转发到远程主机：

```
localhost:local_port -> SSH服务器 -> remote_host:remote_port
```

### 远程隧道 (`-R`)

将远程端口的流量转发到本地机器：

```
SSH服务器:remote_port -> localhost:local_port
```

## 依赖

- [gorilla/mux](https://github.com/gorilla/mux) - HTTP 路由库

## 开源协议

MIT
