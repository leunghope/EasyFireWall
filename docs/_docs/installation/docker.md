---
layout: default
title: Docker
description: 一个多架构镜像、三个镜像仓库、四种标签——以及为什么它需要 host 网络。
---

# Docker

easywall 过滤的是**宿主机**的防火墙，所以它跑在你想要保护的那台机器上——而那通常不是你浏览网页用的机器。下面的一切都以此为前提：命令在服务器上执行，浏览器在你自己的机器上。

## 首次运行

在服务器上，通过 SSH：

```bash
git clone https://github.com/jp1337/easywall.git
cd easywall
docker compose up -d
```

然后在你自己的机器上打开 `https://<server>:12227`——也就是你平时访问那台主机用的地址。**只有当 easywall 就跑在你面前的机器上时，`localhost` 才有效**，而这正是本页过去唯一给出的指示。

首次打开那个页面时会看到五件事：

| | |
|---|---|
| 一个初始化令牌输入框 | `docker compose logs easywall \| grep 'setup token' \| tail -1`——见[首次运行]({{ '/docs/installation/first-run/' | relative_url }}#the-setup-token) |
| 一条证书警告 | easywall 在首次启动时自己生成证书。接受它，或者[使用你自己的证书](#your-own-certificate) |
| 还没有任何东西被过滤 | 全新容器不带规则，所以在你和 12227 端口之间没有任何 easywall 加的东西 |
| 容器状态是 `unhealthy` | 这是正确的，第一次应用之后就会转好——见[健康检查](#the-health-check) |
| 本机上的容器 | 全新安装时已被信任——自 2.22 起 `docker.enabled = true` 就是默认值；2.21 留下的文件保持它写的内容；要关掉它或改掉它接受的范围，见[与 Docker 共存]({{ '/docs/features/docker/' | relative_url }}) |

如果页面完全打不开，那么拦住它的不是 easywall。检查宿主机上已有的防火墙，以及云厂商安全组里 12227 端口的规则。

接着完成[初始化设置]({{ '/docs/installation/first-run/' | relative_url }})——其中会涉及 SSH 端口，这是唯一可能把你关在门外的答案。动手之前，先看[环境变量]({{ '/docs/environment/' | relative_url }})，了解 `docker-compose.yml` 完全不改 `./easywall-config` 就能设置什么——以及它刻意不能设置什么。

> **`./easywall-config` 归你所有，首次启动时会换掉属主。** Docker 把它建成空的。入口脚本用镜像内的默认值填充它，布局与 Debian 软件包安装的一致。`web.toml` 和 `ssl/` 归容器的 `easywall` 用户（uid 100，gid 101）；`easywall.toml` 归 root。它在 .gitignore 里：账户就存在那儿。**在宿主机上编辑这些文件需要 `sudo`。** 当 `/var/lib/easywall` 是 bind mount 时，入口脚本做同样的事：目录归 root，而 web 进程把自己的通行密钥和 TOTP 重放存储放在其中的 `web/` 里。

> **从 2.21 或更早版本升级 git 检出？** 那时 compose 挂载的是 `./config`，即受版本控制的默认值。拉取之前，先把文件搬一次：
>
> ```bash
> docker compose down
> sudo mv config easywall-config && sudo rm -f easywall-config/embed.go
> git checkout -- config && git pull
> docker compose up -d
> ```
>
> 现有的 `easywall.toml` 原样搬走：它对 `docker.enabled` 写了什么就保留什么，不会被新默认值替换。跳过这一步，容器会开始一次全新的首次运行——而初始化令牌确保别人无法抢先认领。

## 从其他地方访问

直接访问——从你自己的机器打开 `https://<server>:12227`——是推荐的路径。把它开着是安全的：`/firstrun` 现在需要容器自己日志里的初始化令牌，光抢先访问并不够。如果防火墙或云厂商安全组把该端口对除你之外的所有人关闭，还有两条路：

| 方式 | 命令 | 注意 |
|---|---|---|
| 一条 SSH 隧道 | `ssh -L 12227:127.0.0.1:12227 user@server`，然后打开 `https://127.0.0.1:12227` | `administratively prohibited` 说明 `sshd` 设了 `AllowTcpForwarding no`——不是 easywall 的问题 |
| 一个反向代理 | 见[反向代理之后]({{ '/docs/installation/reverse-proxy/' | relative_url }}) | 经 SNAT 或位于 overlay 网络的代理地址，需要从[审计日志]({{ '/docs/features/audit-log/' | relative_url }})读出真实地址，并写进 `trusted_proxies` |

## 必须持久化的内容

| 容器内路径 | 存放 | 没有它会怎样 |
|---|---|---|
| `/etc/easywall` | `easywall.toml`、`web.toml`、证书 | 初始化会从头再来一遍 |
| `/var/lib/easywall` | `rules.json`、应用状态、通行密钥 | 所有规则都没了 |
| `/var/log/easywall` | `audit.log`、`packets.log`（当 `persist = true`） | 审计记录和被拒数据包历史在每次更新后都从空开始 |

```bash
docker run -d --name easywall --network host --cap-add NET_ADMIN \
  -v easywall_config:/etc/easywall \
  -v easywall_data:/var/lib/easywall \
  -v easywall_logs:/var/log/easywall \
  ghcr.io/jp1337/easywall:latest
```

没有这个挂载时，Docker 的 `VOLUME` 指令会给每个重建的容器一个新的空匿名卷，其中的历史在下次更新之后就没了。

## 健康检查 {#the-health-check}

镜像自带一个 `HEALTHCHECK`，每十秒请求一次 `/healthz`；`docker-compose.yml` 在 `healthcheck:` 下有同样的检查。自 2.19 起定义有两份而不是一份，因为 podman 的默认镜像格式没有地方放第一份——见下面的 [Podman](#podman)。有测试保证两者一致。

```bash
docker compose ps                     # healthy / unhealthy
docker exec easywall easywall-core health
```

**只要防火墙没有在过滤，它就报 unhealthy**——全新容器、被删掉的表、死掉的 web 进程，或者不再应答的 core。这是新行为：有整整一年，检查只看 core 的 socket，于是界面已经死掉的容器仍然显示 *Up*。

有两个后果，最好在遇到之前就知道：

- **对 easywall 设 `depends_on: condition: service_healthy`，在一个从未应用过规则的容器上会永远等待。** 先应用一次，或者去掉这个条件。
- **`degraded` 的防火墙仍然算健康。** 它还在过滤，而重启解决不了它的任何成因。[健康检查]({{ '/docs/features/health/' | relative_url }})列出了各种状态、退出码，以及如何让一台监控主机读取 `/healthz`。

## Podman {#podman}

**`podman compose up -d` 不需要额外操作。** 自 2.19 起 `docker-compose.yml` 自己声明了 `healthcheck:` 块，所以无论镜像有没有，容器都会有一份。

**直接 `podman build` 加 `podman run` 仍然需要 `--format docker`：**

```bash
podman build --format docker -t easywall .
podman run -d --network host --cap-add NET_ADMIN easywall
```

**不加它，镜像就不带健康检查，而构建照样成功。** OCI 是 podman 的默认镜像格式，没有 healthcheck 字段，所以 `podman build` 退出码为 `0`，留下 `HealthCheck: null`。直接 `podman run` 时这里也没有 compose 文件可回退。检查你拿到了什么：

```bash
podman image inspect --format '{% raw %}{{ .HealthCheck }}{% endraw %}' easywall
```

需要只停或启动其中一个进程而不重启两个时，`supervisorctl` 在容器内可用，走的是仅限 root 的 socket：

```bash
docker exec easywall supervisorctl status
```

## 从哪里拉取

同一个 `linux/amd64` + `linux/arm64` 镜像，由 CI 推送到这三处。公开，无需认证。

| 镜像仓库 | 镜像 |
|---|---|
| GitHub Container Registry | `ghcr.io/jp1337/easywall` |
| Docker Hub | `docker.io/kermit1337/easywall` |
| Quay.io | `quay.io/jp1337/easywall` |

> **Quay 落后了。** 发布令牌正被拒绝，因此该镜像源在替换之前会被跳过，它也没有最近的版本。请用 GHCR 或 Docker Hub。

## 选哪个标签 {#which-tag}

| 标签 | 何时移动 | 适用于 |
|---|---|---|
| `:latest` | 只在**稳定版**发布时移动——候选版本不会移动它 | **生产环境** |
| `:vX.Y.Z` | 从不移动 | 固定版本，例如 `v{{ site.version }}` |
| `:edge` | `main` 上每次构建通过之后 | 跟踪开发版、[演示模式]({{ '/docs/installation/demo/' | relative_url }}) |
| `:sha-<commit>` | 从不移动 | 回滚与调试 |

```yaml
# docker-compose.override.yml — pin a version
services:
  easywall:
    image: ghcr.io/jp1337/easywall:v{{ site.version }}
```

## 为什么用 host 网络

nftables 作用于宿主机的网络命名空间。在它自己的命名空间里，容器只会过滤自己的流量，那不是你要的。`NET_ADMIN` 让它能发出触达宿主机表的 netlink 调用。

```yaml
network_mode: host
cap_add:
  - NET_ADMIN
security_opt:
  - no-new-privileges:true
```

只要 `NET_ADMIN`，别的都不要。随附的 compose 文件还曾要求 `SYS_MODULE`——加载内核模块的能力。对一个已经共享宿主机网络的容器来说，这项能力就等于换了名字的宿主机 root。本页从未列出过它。它已经去掉了；不加它也能应用完整规则集，这一点验证过。如果 `nf_tables` 没加载，在宿主机上执行 `modprobe nf_tables`——已经在跑 nftables 的主机自然有它。

数据包日志的 `nfnetlink_log` 需要同样的加载方式，也以同样的方式得到。第一次有 NFLOG 组绑定时，内核会按名字请求它，触发条件不过是这个容器已经拥有的 `NET_ADMIN`。只有完全禁用模块加载的主机才会回退到内核日志。

这也是容器里的 easywall 仍能与 Docker 自己的规则共存的原因——它只占有 [`table inet easywall`]({{ '/docs/features/docker/' | relative_url }})，别的都不碰。

## 需求

- Docker Engine 24+、Compose v2
- 一台带 nftables 的 Linux 宿主机

## 使用自己的证书 {#your-own-certificate}

```yaml
volumes:
  - /etc/letsencrypt:/etc/letsencrypt:ro
  - ./easywall-config:/etc/easywall
```

```toml
# ./easywall-config/web.toml
[tls]
cert = "/etc/letsencrypt/live/example.com/fullchain.pem"
key  = "/etc/letsencrypt/live/example.com/privkey.pem"
```

**这个镜像不支持 `tls.acme`。** web 进程会降权到无特权用户，没有任何途径重新获得 `CAP_NET_BIND_SERVICE`。`no-new-privileges:true` 是本项目刻意保留的加固措施，不会为了一个功能放弃。请改为在宿主机上运行你自己的 ACME 客户端（certbot，或 Let's Encrypt 官方容器），并按上面的方式挂载它的输出。

> **通行密钥同样需要 `tls.hostname`。** 它是 WebAuthn 的 Relying Party ID，与上面的证书分开设置——见[配置 → `[tls]`]({{ '/docs/configuration/' | relative_url }}#tls)。

## 更新

```bash
docker compose pull && docker compose up -d
```

重建容器会恢复最后一次应用的规则，**没有确认窗口**——没有任何机制能拦下会让你失联的恢复，而通过界面应用时是有的。拉取和重建要刻意为之，不要放在定时器里。

生产环境请固定 `:vX.Y.Z`（见[选哪个标签](#which-tag)），并以同样方式更新。自动化的常见做法 [Watchtower](https://containrrr.dev/watchtower/) 已归档，自 2023 年起没有发布——跳过它，或者让它指向的标签不比你自己升级的固定标签更宽松。

## 核对拉到的镜像

每个镜像都在一个 OCI 标签里带着源码提交：

```bash
docker pull ghcr.io/jp1337/easywall:latest
docker image inspect \
  --format '{% raw %}{{ index .Config.Labels "org.opencontainers.image.revision" }}{% endraw %}' \
  ghcr.io/jp1337/easywall:latest
```

> **这里以前读的是 `.Manifest.Annotations`，那个值从来不在那儿。** 标签写入的是镜像的 *config*；manifest annotations 是另一个字段，只承载基础镜像与创建相关的键。验证方式是构建一个带 `--label org.opencontainers.image.revision=abc123` 的镜像，再读取推送后的 OCI 布局：`manifest annotations` 里是 `image.base.name` 和 `image.created`，而 `abc123` 在 `config.Labels` 里。所以那条命令什么都不打印，看起来像个没有来源信息的镜像。发布镜像当时确实也没有——现在标签在 `Dockerfile` 里设置，那是三条构建路径共用的唯一文件。

把它与发布标签指向的提交对比。发布这些镜像的工作流是 [`release.yml`](https://github.com/jp1337/easywall/blob/main/.github/workflows/release.yml) 和 [`publish-edge.yml`](https://github.com/jp1337/easywall/blob/main/.github/workflows/publish-edge.yml)。
