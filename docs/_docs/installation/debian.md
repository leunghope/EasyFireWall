---
layout: default
title: Debian / Ubuntu
description: .deb 软件包——两个 systemd 单元，配置位于 /etc/easywall，到此为止。
---

# Debian / Ubuntu

```bash
ARCH=$(dpkg --print-architecture)      # amd64 or arm64
wget https://github.com/jp1337/easywall/releases/latest/download/easywall_$ARCH.deb
sudo dpkg -i easywall_$ARCH.deb && sudo apt-get install -f
```

| 架构 | 文件 | 典型主机 |
|---|---|---|
| `amd64` | `easywall_amd64.deb` | 一台 VPS、一台物理服务器 |
| `arm64` | `easywall_arm64.deb` | 一台树莓派、一台 arm64 VPS、Apple 芯片的虚拟机 |

文件名从不带版本号——版本号在软件包内部，`dpkg -l easywall` 与 `apt policy easywall` 从那里读取，所以这条命令不必为每个版本改动。

完整清单就这两个。如果 `dpkg --print-architecture` 输出的是别的值——例如 32 位树莓派镜像上的 `armhf`——那就没有对应的软件包，下载会 404；请改为[从源码构建]({{ '/docs/installation/manual/' | relative_url }})。

然后打开 `https://<server>:12227`，完成[初始化设置]({{ '/docs/installation/first-run/' | relative_url }})。它索要的初始化令牌在 `sudo journalctl -u easywall-web -g 'setup token'` 里——安装过程也会打印这条命令。证书在首次启动时是自签名的，所以浏览器会警告——接受它，或者[配置你自己的证书](#your-own-certificate)。

两个软件包都在各自**本机**架构的 runner 上构建，从不交叉编译，而且每个包在发布前都会在该架构上安装并启动一次。若某个包的 `Architecture` 字段或其二进制与构建它的那条流水线不一致，发布流程就会拒绝发布它。

> **2.5.0 之前的每一次发布都缺少这个文件。** 没有任何一次发布附带过 `.deb`。它曾是一个七天后过期的 CI 产物，还需要登录 GitHub，所以文档里写的安装路径对任何人都从未生效。从 2.5.1 起，发布流程会发布它，并拒绝发布其中没有二进制的包。

## 各部分都在哪里

| 路径 | |
|---|---|
| `/etc/easywall/easywall.toml` | core：防火墙选项、确认窗口、Docker |
| `/etc/easywall/web.toml` | web：身份认证、TLS、语言、绑定地址 |
| `/var/lib/easywall/rules.json` | 三个规则集 |
| `/var/lib/easywall/web/` | web：通行密钥、TOTP 重放防护、版本缓存 |
| `/var/log/easywall/audit.log` | [审计日志]({{ '/docs/features/audit-log/' | relative_url }})，每天轮转 |
| `/etc/easywall/*.toml.template` | 软件包自带的带注释默认配置 |

完整的键名参考：[配置]({{ '/docs/configuration/' | relative_url }})。

**升级永远不会改动你的两个 `.toml` 文件。** 软件包安装的是模板，只有真实文件缺失时才会创建它。因此两者都不是 dpkg 的 conffile，升级不会就它们发问，也不会覆盖它们。两个文件都由 easywall 自己编辑。设置页面写 `easywall.toml`；向导和密码页面写 `web.toml`。一个由程序自己改写的文件，本就不该交给包管理器管理。模板也是查找新版本新增键的地方：升级时模板会被替换，你的文件不会。

想从默认值重新开始，删除该文件并重装软件包，或者自己复制模板：

```bash
sudo cp /etc/easywall/easywall.toml.template /etc/easywall/easywall.toml
sudo chown root:root /etc/easywall/easywall.toml && sudo chmod 600 /etc/easywall/easywall.toml
sudo systemctl restart easywall-core
```

## 运行

```bash
systemctl status  easywall-core easywall-web
journalctl -u easywall-core -f
journalctl -u easywall-web  -f
systemctl restart easywall-core easywall-web
```

`easywall-core` 以 root 运行；`easywall-web` 以无特权的 `easywall` 用户运行。两者都随开机启动，而 core 在启动时会把它最后一次确认的规则重新写入内核——见[恢复与恐慌模式]({{ '/docs/features/recovery/' | relative_url }})。

## 使用自己的证书 {#your-own-certificate}

```toml
# /etc/easywall/web.toml
[tls]
cert = "/etc/letsencrypt/live/example.com/fullchain.pem"
key  = "/etc/letsencrypt/live/example.com/privkey.pem"
```

```bash
systemctl restart easywall-web
```

`easywall` 用户必须能读取这两个文件——Let's Encrypt 的 `privkey.pem` 默认是 `root:root 0600`。

## 卸载

```bash
sudo apt remove easywall          # keeps /etc/easywall, the rules and the audit log
sudo apt purge  easywall          # removes /etc/easywall, /var/lib/easywall, /var/log/easywall
```

执行 purge 之后，`easywall` 系统用户和组仍会保留。

**purge 不会**清除内核中的规则。如果你想让这台主机重新开放，请先执行这一步：

```bash
sudo nft delete table inet easywall
```
