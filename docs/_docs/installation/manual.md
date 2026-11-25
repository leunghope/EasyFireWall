---
layout: default
title: 手动安装
description: 自己构建两个二进制并安装，适用于发行版没有提供软件包的情况。
---

# 手动安装

适用于没有软件包的发行版。先看[环境要求]({{ '/docs/installation/requirements/' | relative_url }})。

## 构建

在**构建**机器上准备 Go 1.27+（外加 Node.js）——目标机不需要。Go 的版本由 `go.mod` 固定；在 Debian trixie 上要从 backports 安装，因为 trixie 自带的是 Go 1.24。

```bash
git clone https://github.com/leunghope/EasyFireWall.git
cd EasyFireWall
npm ci              # Tailwind, which `make build` runs first
make build          # produces bin/easywall-core and bin/easywall-web
```

构建时会嵌入 `git describe` 得到的版本。交叉编译按常规方式即可：

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 make build
```

## 安装

```bash
sudo make install
```

它把二进制放到 `/usr/sbin`，静态资源放到 `/usr/share/easywall`，三个 systemd 单元放到 `/lib/systemd/system`。它**不会**创建服务用户、目录或配置——这些先做一次：

```bash
sudo groupadd --system easywall 2>/dev/null || true
sudo useradd  --system --no-create-home --shell /usr/sbin/nologin \
              --gid easywall easywall 2>/dev/null || true

sudo install -d -m 0750 -o root     -g easywall /run/easywall
sudo install -d -m 0750 -o root     -g easywall /etc/easywall
sudo install -d -m 0750 -o easywall -g easywall /etc/easywall/ssl
sudo install -d -m 0750 -o root     -g easywall /var/lib/easywall
sudo install -d -m 0700 -o easywall -g easywall /var/lib/easywall/web
sudo install -d -m 0750 -o root     -g easywall /var/log/easywall

```

**属主和权限才是关键，不是细节。**

| 路径 | 为什么是这个属主和模式 |
|---|---|
| `/etc/easywall` — `root:easywall` `0750` | web 用户必须能**穿过**它到达自己的配置，但不能在其中写入 |
| `easywall.toml` — `root:root` `0600` | root 守护进程读取它。一个面向网络的进程若能改动 root 读取的内容，两进程隔离就白做了 |
| `web.toml` — `easywall:easywall` `0600` | 向导把密码哈希写在这里，密码页面会重写它 |
| `ssl_dir` — `easywall:easywall` `0750` | easywall 在这里生成自己的证书，并在过期前替换 |
| `/var/lib/easywall` — `root:easywall` `0750` | core 的：`rules.json`、应用状态、恐慌模式标记。属主为 root，core 因此不需要 `CAP_DAC_OVERRIDE`；组不可写，web 用户就无法替换 root 读取的内容 |
| `/var/lib/easywall/web` — `easywall:easywall` `0700` | web 进程自己的状态：通行密钥、TOTP 重放防护、版本缓存 |
| `/run/easywall` — `root:easywall` `0750` | 存放 web 进程所连接的控制 socket |

> **这里以前把 `/var/lib/easywall` 写成 `easywall:easywall`，那样行不通。** 该单元把守护进程的权限收缩到 `CAP_NET_ADMIN`，对 root 服务而言这同时会去掉 `CAP_DAC_OVERRIDE`。于是 root 无法进入不属于自己的目录。`rules.json` 永远写不进去。每个请求都会在调用深处抛出的权限错误上失败。软件包也有同样的 bug，上面的 `root:easywall` 正是它的修复。

> **在 2.22 之前这样装过？** `/var/lib/easywall` 当时是 `0770` 且共享，web 用户能替换其中 core 的文件。启动新二进制之前，先做一次：
>
> ```bash
> sudo install -d -m 0700 -o easywall -g easywall /var/lib/easywall/web
> sudo chmod 0750 /var/lib/easywall
> ```
>
> easywall-web 下次启动时会把自的文件复制进 `web/`。在模式改好之前，easywall-core 每次启动都会记录一条错误。新的 `easywall-web.service` 没有 `web/` 根本无法启动。

## 配置

两个二进制都自带带注释的默认配置，并能把它写出来。两者都不会覆盖已存在的文件。

```bash
sudo easywall-web  --write-config /etc/easywall/web.toml
sudo chown easywall:easywall      /etc/easywall/web.toml

sudo easywall-core --write-config /etc/easywall/easywall.toml
```

两者的权限都是 `0600`；正是那条 `chown` 让 `web.toml` 可被需要重写它的进程写入。每个键的说明见[配置]({{ '/docs/configuration/' | relative_url }})。

会话密钥出厂是一个明显的占位符，easywall 会在首次启动时替换它。自己生成更好——这样在端口开通之前它就已经存在：

```bash
sudo sed -i "s|CHANGE_ME[A-Z0-9_]*|$(openssl rand -hex 32)|" /etc/easywall/web.toml
```

> **不需要 `csrf_key`。** CSRF 防护用的是 Go 1.25 的 `net/http.CrossOriginProtection`，它校验 `Origin` 和 `Sec-Fetch-Site`，而不是发放令牌。旧配置和本页旧版本要求的那把密钥，没有任何代码读取。

## 启动

```bash
sudo systemctl enable      easywall-selftest
sudo systemctl enable --now easywall-core easywall-web
systemctl status easywall-core easywall-web     # both should be active (running)
```

`easywall-selftest` 只启用，从不手动启动：

| | |
|---|---|
| 它是什么 | 一个 `oneshot` 单元，排在 `Before=easywall-core` |
| 它做什么 | 针对运行中的内核对规则构建器做验证 |
| 多久运行一次 | 仅在版本或内核变化时 |
| 为什么单独拆开 | 它是唯一持有 `CAP_SYS_ADMIN` 的单元，构建一次性的网络命名空间需要该能力 |
| 失败会怎样 | 没有影响。core 照常启动；`easywall-core health` 会说明它发现了什么 |

打开 `https://<server>:12227`，接受自签名证书，完成初始化向导。

## 卸载

```bash
sudo systemctl disable --now easywall-core easywall-web
sudo systemctl disable easywall-selftest
sudo rm -f /lib/systemd/system/easywall-core.service \
           /lib/systemd/system/easywall-web.service \
           /lib/systemd/system/easywall-selftest.service
sudo systemctl daemon-reload

sudo rm -f  /usr/sbin/easywall-core /usr/sbin/easywall-web
sudo rm -rf /usr/share/easywall

# the rules stay in the kernel until you say otherwise
sudo nft delete table inet easywall 2>/dev/null || true

# config, rules and logs — back them up first if you want them
sudo rm -rf /etc/easywall /var/lib/easywall /var/log/easywall
sudo userdel  easywall 2>/dev/null || true
sudo groupdel easywall 2>/dev/null || true
```
