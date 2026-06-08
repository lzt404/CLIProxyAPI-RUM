# Docker 镜像发布与一键使用

这份文档面向 RUM fork。目标是把镜像发布到 GitHub Container Registry
（GHCR），让用户不用克隆源码，只下载 compose 文件就能运行。

## 镜像地址

用户默认使用 `latest`：

```text
ghcr.io/lzt404/cliproxyapi-rum:latest
```

每次推送 RUM release tag 时，Actions 会同时发布一个不可变版本 tag：

```text
ghcr.io/lzt404/cliproxyapi-rum:v7.1.31-rum.1
```

## 发布者：发布镜像

本仓库的 `.github/workflows/docker-image.yml` 会发布多架构镜像：

- `linux/amd64`
- `linux/arm64`

## 发布者：自动发布 latest

RUM fork 的 tag 必须保留 `-rum.N` 后缀，例如：

```bash
git tag v7.1.31-rum.1
git push origin v7.1.31-rum.1
```

推送 tag 后，GitHub Actions 会自动构建并发布：

```text
ghcr.io/lzt404/cliproxyapi-rum:latest
ghcr.io/lzt404/cliproxyapi-rum:v7.1.31-rum.1
```

Actions Summary 里也会自动显示可复制的拉取命令：

```bash
docker pull ghcr.io/lzt404/cliproxyapi-rum:latest
docker pull ghcr.io/lzt404/cliproxyapi-rum:v7.1.31-rum.1
```

如果只想手动重发当前分支的 `latest`，进入 GitHub 仓库页面：

```text
Actions -> docker-image -> Run workflow
```

手动运行不需要填写镜像 tag，会直接发布 `ghcr.io/lzt404/cliproxyapi-rum:latest`。

## 首次发布后的可见性

GHCR 包可能需要手动设为公开。第一次发布成功后，进入仓库右侧或个人主页的
`Packages`，找到 `cliproxyapi-rum`，在 package settings 里把 visibility 改成
public。

## 用户：一键运行

用户只需要 Docker，不需要 Go，也不需要克隆源码。

### Windows PowerShell

```powershell
mkdir cliproxy-rum
cd cliproxy-rum
curl.exe -fsSLO https://raw.githubusercontent.com/lzt404/CLIProxyAPI-RUM/main/docker-compose.yml

$env:CLI_PROXY_API_KEY = "sk-local-change-me"
$env:MANAGEMENT_PASSWORD = "change-me"
docker compose up -d
```

### Linux/macOS

```bash
mkdir -p cliproxy-rum
cd cliproxy-rum
curl -fsSLO https://raw.githubusercontent.com/lzt404/CLIProxyAPI-RUM/main/docker-compose.yml

CLI_PROXY_API_KEY="sk-local-change-me" MANAGEMENT_PASSWORD="change-me" \
  docker compose up -d
```

启动后访问：

```text
http://127.0.0.1:8317
```

OpenAI 兼容地址：

```text
http://127.0.0.1:8317/v1
```

客户端 API Key 填 `CLI_PROXY_API_KEY` 的值，例如：

```text
sk-local-change-me
```

## 数据保存位置

首次启动会生成：

```text
cliproxy-data/config.yaml
cliproxy-data/auths/
cliproxy-data/logs/
```

容器删掉后，这些数据仍然保留在当前目录。

## 重要说明

Docker 镜像只负责运行 CLIProxyAPI。用户仍然需要配置至少一种上游凭据，例如：

- OAuth 登录后的 Codex/Claude/Gemini 账号
- OpenAI 兼容中转站的 API Key
- Gemini API Key、Claude API Key、Codex API Key 等

如果没有任何上游账号或 API Key，服务可以启动，但没有模型可调用。

首次启动后，如果要修改配置，编辑：

```text
cliproxy-data/config.yaml
```

然后重启：

```bash
docker compose restart
```
