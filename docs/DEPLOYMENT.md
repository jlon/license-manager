# 生产部署实录（www.xclaw.live）

> **本文是唯一反映当前生产环境的部署文档。**
> `docs/github-images-deploy.md`、`docs/CI-images-deploy.md` 与
> `.github/workflows/deploy_{frontend,backend}.git.yml` 均已与实际环境不符，
> 见文末「已废弃的路径」。
>
> 最后核对时间：2026-10-27（以线上实际状态为准，非文档推测）。

## 一句话流程

本地 `docker build` → 推送 Docker Hub → 服务器 `podman-compose` 拉取并重建前端容器。

## 关键事实（与旧文档的差异）

| 项目 | 旧文档写的 | 实际生产环境 |
|---|---|---|
| 镜像仓库 | GHCR `ghcr.io/cedar-v/...` | **Docker Hub** `docker.io/johndockerhub/license-manager` |
| 容器运行时 | `docker` / `docker-compose` | **`podman` / `podman-compose`**（服务器上无 `docker` 命令） |
| 前端交付 | compose 拉预构建镜像 | 本地构建 + 手动 `docker push`，无 CI 参与 |
| 编排文件 | 项目根 `docker-compose.prod.yml` | 服务器 **`/opt/stellar-license/docker-compose.yml`** |
| 部署来源 | — | 源码仓库 `/opt/stellar-license/source`（仅作参考，镜像不经它构建） |

## 拓扑

```
外网 → nginx (:80/:443) → /license/      → 127.0.0.1:18080  frontend 容器（nginx:alpine）
                        → /license/api/  → 127.0.0.1:18888  api 容器
                        → /license/health→ 127.0.0.1:18888/health
```

- 前端构建参数固定为 `VITE_ADDRESS_BASE_URL=/license/`、`VITE_API_BASE_URL=/license/`
  （Dockerfile.frontend 的默认值），与 nginx 的 `/license/` 子路径一致。
- 前端镜像内置 `nginx.license.conf`（`NGINX_CONF` 默认值），不要再改成 `nginx.conf`。
- 容器名：`stellar-license_{frontend,api,mysql}_1`。

## 前端部署步骤

### 1. 本地构建

在仓库根目录（`Dockerfile.frontend` 需要 `frontend/` 作为上下文的一部分）：

```bash
TS=$(date +%Y%m%d%H%M%S)
TAG="frontend-$TS"
docker build -f Dockerfile.frontend -t "johndockerhub/license-manager:$TAG" .
```

镜像标签沿用 `frontend-<YYYYMMDDHHMMSS>` 约定，便于回滚与追溯。

### 2. 推送前自检（推荐）

确认镜像里真的是新代码，避免推送/部署的是旧产物：

```bash
docker run --rm "johndockerhub/license-manager:$TAG" \
  sh -c 'grep -rlo "设备管理" /usr/share/nginx/html/static/js/ | head -1'
```

### 3. 推送

```bash
docker push "johndockerhub/license-manager:$TAG"
```

### 4. 服务器更新镜像标签

```bash
ssh root@www.xclaw.live
cd /opt/stellar-license
cp docker-compose.yml docker-compose.yml.bak-$(date +%Y%m%d-%H%M%S)   # 必须先备份
sed -i "s|license-manager:frontend-[0-9]*|license-manager:$TAG|" docker-compose.yml
grep -n "license-manager:" docker-compose.yml                          # 确认只改了前端那行
```

只替换 `frontend` 的标签，`api` 的标签保持不动，避免误升级后端。

### 5. 拉取并重建（只动前端）

```bash
podman-compose pull frontend
podman-compose up -d --no-deps --force-recreate frontend
```

`--no-deps` 防止连带重启 api/mysql；这两个服务不应因前端发布而中断。

### 6. 验证

```bash
podman ps --format "{{.Names}}|{{.Image}}|{{.Status}}" | grep license
curl -sS -o /dev/null -w '%{http_code}\n' https://www.xclaw.live/license/
curl -sS -o /dev/null -w '%{http_code}\n' https://www.xclaw.live/license/health
```

判定标准：容器 `healthy`、两处均返回 `200`、`podman ps` 里的镜像标签等于本次 `$TAG`。

进一步确认打包内容（可验证具体文案/资源已上线）：

```bash
JS=$(curl -sS https://www.xclaw.live/license/ | grep -o '/license/static/js/index\.[^"]*\.js' | head -1)
curl -sS "https://www.xclaw.live$JS" | grep -c '设备管理'
```

## 后端部署

后端**不随前端发布**。如需升级后端：

```bash
docker build -f Dockerfile.backend -t "johndockerhub/license-manager:stellar-<时间戳>" .
docker push "johndockerhub/license-manager:stellar-<时间戳>"
```

然后在 `/opt/stellar-license/docker-compose.yml` 只改 `api` 的标签，`podman-compose up -d --no-deps --force-recreate api`。

注意 `api` 挂载了三个只读文件，缺失会导致启动失败：

- `./runtime/config.yaml` → `/app/backend/cmd/config.yaml`
- `./runtime/rsa_private_key.pem`
- `./runtime/rsa_public_key.pem`

## 回滚

镜像标签是按时间戳递增的，回滚就是把标签改回上一个：

```bash
cd /opt/stellar-license
grep -n "license-manager:" docker-compose.yml            # 看当前标签
cp docker-compose.yml docker-compose.yml.bak-$(date +%Y%m%d-%H%M%S)
sed -i "s|license-manager:frontend-[0-9]*|license-manager:frontend-<上一个时间戳>|" docker-compose.yml
podman-compose pull frontend
podman-compose up -d --no-deps --force-recreate frontend
```

目录下保留的 `docker-compose.yml.bak-*` 可直接拷回用于回滚。

## 常见坑

- **服务器没有 `docker` 命令。** 所有容器操作使用 `podman` / `podman-compose`。
- **不要用项目根的 `docker-compose.prod.yml` 部署生产。** 它是通用模板，镜像变量名（`FRONTEND_IMAGE`）与生产编排不一致。
- **前端子路径必须保持 `/license/`。** 若用 `VITE_ADDRESS_BASE_URL=/` 构建，静态资源会 404。
- **只重建前端时务必带 `--no-deps`。** 否则会连带重启 api 与 mysql。
- **`/license/api/` 有 10 r/m 限流**（nginx `limit_req_zone license_per_ip`），批量操作会触发 429；这是有意配置，不要为测试而调高。
- 修改 nginx 配置的参考片段在 `/opt/stellar-license/nginx.conf.snippet`，它需要合并进 `/etc/nginx/conf.d/xclaw.conf` 的 server 块，而不是单独作为一个 conf 文件。

## 已废弃的路径（不要使用）

以下内容仍留在仓库里，但**不反映生产环境**，看到时请直接忽略或按本文执行：

| 文件 | 为什么废弃 |
|---|---|
| `.github/workflows/deploy_frontend.git.yml` | 构建后 rsync 到 `/home/pi/lm-www/frontend`。该目录在服务器上**已不存在**；此工作流在此 fork 上从未成功运行过。 |
| `.github/workflows/deploy_backend.git.yml` | 依赖服务器 `/root/license-manager` 目录，该目录**不存在**。 |
| `.github/workflows/release_{frontend,backend}.yml` | 推送到 **GHCR**，而生产从 **Docker Hub** 拉取，两者完全不相通。 |
| `docs/github-images-deploy.md` | 假设 GHCR + `docker-compose` + `docker` 命令，三者均与实际不符。 |
| `docs/CI-images-deploy.md` | 同上，且假设镜像由 GitHub Actions 构建；实际为本地构建手动推送。 |

## 变更记录

| 日期 | 内容 |
|---|---|
| 2026-10-27 | 建立本文；记录 Docker Hub + podman 实际流程，标注上述废弃项 |
| 2026-10-27 | 部署 `frontend-20261027112739`（补齐设备管理中/英/日三语文案） |
