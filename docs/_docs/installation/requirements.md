---
layout: default
title: 环境要求
description: 主机需要什么、不需要什么，以及该走哪条安装路径。
---

# 环境要求

{% include themed-figure.html base="/assets/diagrams/install-choice" ext="svg"
   alt="决策树：只是看看，进入演示模式；Debian 或 Ubuntu，使用 .deb 软件包；已经在跑容器，使用 Docker；其余情况从源码构建。" %}

## 主机

| | 最低要求 | |
|---|---|---|
| 内核 | 3.13 | nftables netlink API 问世的版本 |
| nftables | 任意版本 | `apt install nftables` |
| init | systemd——Debian 软件包 | 用于服务单元 |
| init | 无——Docker | 容器内由 `supervisord` 运行两个进程 |
| 架构 | amd64 或 arm64 | 两种架构都提供二进制 |
| 内存 | 约 32 MB | 两个进程空闲时 |
| 内存，登录时 | 峰值约 200 MB | 同一时间只做一次密码校验（Argon2id，64 MiB），无论同时来多少请求 |
| 磁盘 | 约 20 MB | 二进制、静态资源、配置 |
| 权限 | `CAP_NET_ADMIN` | core 需要它来写规则。web 进程不需要 |

数据包日志的 NFLOG 组需要 `nfnetlink_log` 内核模块。这里没有任何地方手动加载它——第一次有持有 `CAP_NET_ADMIN` 的进程绑定 NFLOG 组时，内核会自动加载，方式与自动加载 `nf_tables` 相同。只有完全禁用模块加载的主机才需要自己执行 `modprobe nfnetlink_log`，并且在加载成功之前，日志会回退到内核日志。

## 测试过的环境

| 发行版 | 架构 |
|---|---|
| Debian 12 · Debian 11 | amd64, arm64 |
| Ubuntu 24.04 LTS · 22.04 LTS | amd64, arm64 |
| Raspbian（Debian 12） | arm64 |

其他带 nftables 的 systemd 发行版——Arch、Fedora、openSUSE——应当可用，但不在 CI 中。

## 端口

| 端口 | 方向 | 用途 |
|---|---|---|
| 12227/tcp | 入站 | Web 界面，仅 HTTPS |
| 80/tcp | 入站 | 仅当 `tls.acme = true` 时——证书颁发机构用 HTTP-01 challenge 证明对 `tls.hostname` 的控制权。不由 easywall 自己的规则开放；见[安全 → 传输]({{ '/docs/security/' | relative_url }}#the-one-exception-acmes-port-80) |

没有其他明文端口。两个进程通过 Unix socket 通信，而不是端口。

## 不需要

- Python、Node.js，或两个二进制之外的任何运行时
- 数据库
- 目标机上的 Go 工具链——用 `.deb` 或发布版二进制
- **出站互联网访问。** 字体、样式表和脚本都由 easywall 自身提供，因此在隔离网络的主机上也能正确渲染。有两处会向外连接，且两者都静默失败。更新检查默认开启，可以关闭。安装计数默认关闭，除非有人开启。两者的完整说明都在[安全]({{ '/docs/security/' | relative_url }})下

## 从 easywall v1 迁移

v1 用 iptables 和 YAML；v2 用 nftables 和 JSON。规则文件无法导入——请重建规则，或在 v2 运行后使用[导出/导入]({{ '/docs/features/export-import/' | relative_url }})。

```bash
sudo systemctl disable --now easywall easywall-web   # stop v1
pip uninstall easywall                               # if installed via pip

# optional: v1's iptables rules are independent of nftables and can stay,
# but if you want them gone
sudo iptables -F && sudo iptables -X
sudo ip6tables -F && sudo ip6tables -X
```

**下一步：** [Debian / Ubuntu]({{ '/docs/installation/debian/' | relative_url }}) ·
[Docker]({{ '/docs/installation/docker/' | relative_url }}) ·
[从源码构建]({{ '/docs/installation/manual/' | relative_url }}) ·
[演示模式]({{ '/docs/installation/demo/' | relative_url }}) ·
[首次运行]({{ '/docs/installation/first-run/' | relative_url }})
