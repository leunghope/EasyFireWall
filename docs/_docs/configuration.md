---
layout: default
title: 配置
description: easywall-core 与 easywall-web 的全部 TOML 配置键说明。
---

# 配置

`/etc/easywall/` 下有两个 TOML 文件，每个进程一个，启动时读取。

无法解析的值会让守护进程停止，并在消息中点名对应的键——例如未知的 `ipv6.mode`、缺失的路径。仅仅超出范围的值则会被拉回范围内，并在日志里**明确说出来**。防火墙守护进程拒绝启动，比运行一个有文档说明的默认值更糟。`acceptance.duration` 会被限制在 10–3600。已启用模块上为 0 的速率限制会变成该模块的默认值。

同一个值若从界面传入，则会被拒绝，并点名对应的键。两个方向都不会静默替换——过去有五处速率限制如此，文件与运行中的防火墙可能互相矛盾，却没有任何说明。

| | 所有者与权限 | 内容 |
|---|---|---|
| `easywall.toml` | `root:root` `0600` | 防火墙选项、确认窗口、IPv6、Docker、路由 |
| `web.toml` | `easywall:easywall` `0600` | 监听地址、TLS、会话密钥、凭据 |

这种拆分正是要点所在：web 进程改写自己的文件。初始化向导和密码页面会写入该文件。它绝不能碰 root 守护进程读取的那个文件。

软件包和容器都会安装这两个文件，且已预先填好——来自 `*.toml.template`。两个文件只创建一次，之后不再被触碰：easywall 自己会编辑这两个文件，而程序会改写的文件不能交给包管理器管理。升级只替换模板，你的两个文件保持原样。任一可执行文件也能写出带注释的默认配置——它内置了一份，所以在只装了可执行文件的主机上也能用：

```bash
sudo easywall-core --write-config /etc/easywall/easywall.toml
sudo easywall-web  --write-config /etc/easywall/web.toml
```

**它从不覆盖。** easywall 运行后，两个路径都保存着正在工作的防火墙设置，`web.toml` 还保存会话密钥和密码哈希。若指向已存在的文件，命令会说明这一点，并且不做任何改动。

## 命令行 {#the-command-line}

两个可执行文件都接受三个参数；`easywall-web` 还接受第四个。

| | |
|---|---|
| `-config <path>` | 要读取哪个文件。默认为 `/etc/easywall/easywall.toml` 与 `/etc/easywall/web.toml` |
| `-write-config <path>` | 把带注释的默认配置写到该路径，权限 `0600`，然后退出。文件已存在则拒绝，且不会创建目录 |
| `-version` | 打印版本并退出——即该可执行文件实际构建出的版本 |
| `-healthcheck` | 仅 `easywall-web`：向 `bind_addr` 上的 `/healthz` 发起请求，返回 200 时以退出码 0 结束——即容器的[健康检查]({{ '/docs/features/health/' | relative_url }}) |

`easywall-core` 还有五个子命令：`health` 与 `selftest`（[健康状态]({{ '/docs/features/health/' | relative_url }})）、`status`、`panic` 与 `resume`（[恢复与恐慌模式]({{ '/docs/features/recovery/' | relative_url }})）。`easywall-core -h` 会列出它们。

```bash
easywall-core --version     # easywall-core {{ site.version }}
```

---

## easywall-core（`/etc/easywall/easywall.toml`）

### 顶层键

| 键 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `socket_path` | string | `/run/easywall/core.sock` | Unix socket 路径——必须对 `easywall` 组可访问 |
| `data_dir` | string | `/var/lib/easywall` | `rules.json`、应用状态与恐慌模式标记所在目录。仅归 root 所有：web 进程只写入其中的 `web/` |
| `log_dir` | string | `/var/log/easywall` | 审计日志与规则快照所在目录 |

### `[acceptance]`

两步激活的安全机制。规则集被应用后，core 最多等待 `duration` 秒，以收到明确的确认信号。若没有信号到达，先前的规则集会被自动恢复。

| 键 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `enabled` | bool | `true` | 两步激活的安全机制。安装的 `easywall.toml` 将其设为 `true` |
| `duration` | int | `120` | 未确认时自动回滚前的秒数——10 至 3600 |

把 `duration` 设得足够长，好让你在应用规则后从第二个终端验证连通性。

> **文件省略的键按其随附默认值读取**，所以只写了 `duration` 的文件仍然开启确认窗口。`enabled = false` 不会被限制，也不会被拒绝。关闭时守护进程每次启动都会警告：*applies take effect immediately and nothing will undo them*。

该范围是强制执行的，不只是建议。低于十秒，窗口会在确认页面被读完之前关闭，于是每次应用都会回滚，防火墙再也无法通过界面修改。现有文件中超出范围的值会被带到最近的允许值，并记录警告，而不是让守护进程无法启动。这包括 `0`，它变成 10；关闭窗口靠的是 `enabled = false`。缺失的 `duration` 为 120。通过界面设置的值会直接被拒绝。

### `[usage]`

从内核读取每条规则的数据包计数器的频率。它构成了[端口规则页面]({{ '/docs/features/ports/' | relative_url }})上 **最近使用** 列的数据来源。

| 键 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `interval` | int | `300` | 读取计数器之间的秒数——`0` 会停止定时器 |

`0` 停止的是**定时器**，不是计数。一次应用会在写入之前立即读取计数器，因为写入规则会把每一个计数器清零。所以即使主机上的定时器已关闭，**最近使用** 仍会在每次应用时推进。你失去的是两次应用之间的分辨率：在上次应用一小时后才被使用的端口，会在下一次应用时才被记下时间。

### `[packet_log]`

[`*_log` 开关]({{ '/docs/features/filters/' | relative_url }}#logging)所记录的数据包保存在哪里——即[已拦截流量页面]({{ '/docs/features/blocked-traffic/' | relative_url }})读取的内容。启动时读取；`SIGHUP` 会报告此处的变更并忽略它。

| 键 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `nflog_group` | int | `12227` | `easywall-core` 绑定的 NFLOG 组 |
| `entries` | int | `20000` | 保留的数据包数量，最新的在前——`1000` 至 `200000` |
| `persist` | bool | `false` | 同时写入 `log_dir/packets.log`，使页面在重启后仍有数据 |

一个组在每个网络命名空间中只能绑定一次（若以 host 网络方式安装，则每台主机一次）。如果 ulogd2 已占用 `12227`，core 会记录 *could not bind its NFLOG group*，回退到 2.21 之前那样的内核日志，页面也会给出同样的说明。选一个空闲的组。

`entries` 超出范围时会被带到最近的一端，并记录警告。`nflog_group` 超出 `0`–`65535` 时，守护进程会停止并点名该键。

### `[ipv6]`

| 键 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `mode` | string | `"filter"` | IPv6 如何处理：`filter` 让它经过每一条规则，`passthrough` 在任何规则之前接受它，`block` 丢弃它（回环除外）。`block` 还会把 IPv6 从容器网络中移除，而转发链正是建立在这些网络之上的——此后到达容器的流量遵循 `routing.mode`（在 `closed`/`networks` 下丢弃，除非 `routing.networks` 列出了它；在 `open` 下透传） |
| `icmp_allow_router_advertisement` | bool | `true` | 允许 ICMPv6 类型 133 和 134——SLAAC 地址自动配置所需 |
| `icmp_allow_neighbor_advertisement` | bool | `true` | 允许 ICMPv6 类型 135/136——邻居发现协议所需 |

两个 ICMPv6 键仅在 `mode = "filter"` 下生效。在 `passthrough` 下流量已被接受，在 `block` 下则已经不存在。

> **`enabled` 已废弃。** 它曾被记录为「关闭意味着完全不过滤 IPv6 流量」，实际效果却相反。表是 `inet`，因此每一条规则和丢弃策略仍然作用于 IPv6，被移除的只有 ICMPv6 豁免——结果是 IPv6 既被过滤，*又*无法工作。仍带有该键的配置可以加载，两个旧取值都会变成 `mode = "filter"`。

关闭任一 ICMPv6 键都会在多数网络上破坏 IPv6——SLAAC 与邻居发现正是地址得以获取并保持可达的方式。它们是为那些确实两者都不需要的静态地址主机而存在的。

### `[docker]`

| 键 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `enabled` | bool | `true` | 自动检测 Docker 网桥接口并将其加入允许列表 |
| `allow_bridge_networks` | bool | `true` | 将自动检测到的网桥网络 CIDR 加入允许列表 |
| `custom_networks` | list | `[]` | 无条件加入允许列表的额外 CIDR（在 `enabled = true` 时处理） |
| `published_ports` | string | `"open"` | `open` 或 `filtered`。在 `filtered` 下，只有作用域为 `forwarded` 的端口规则才让任何内容到达已发布的容器端口——自 2.24 起也覆盖 IPv6 容器网络 |

> **自 2.22 起 `enabled` 随附为 `true`。** 没有 `docker*`/`br-*` 接口的主机在两种取值下都不会有任何规则变化。而拥有此类接口、又随附 `false` 的主机，会在第一次应用时失去每个容器的网络。确认窗口对此视而不见：它验证的是操作者自己在 `input` 链上的连接，而容器流量经过的是 `forward` 链。现有文件保留它们原本的取值；只有全新安装或 `--write-config` 才会看到新的默认值。
>
> 非 Docker 的 `br-*` 接口——OpenWrt 的 `br-lan` 是一个，OpenStack 的 `br-ex` 是另一个——同样会被信任。自 2.24 起，其全局或唯一本地 IPv6 前缀也以同样方式被信任；在这样的主机上应设置 `enabled = false`。

> **`published_ports` 刻意不在界面中提供控件。** 一次点击就可能让本机上的每个容器脱离网络。确认窗口抓不住这种情况：它验证的是你自己的连接，而你的连接到达的是 `input` 链。在这里编辑它，应用一次，并先阅读 [Docker 共存]({{ '/docs/features/docker/' | relative_url }})。当 `enabled = false` 时，`filtered` 会在启动时和 `SIGHUP` 时被拒绝；任何既不是 `open` 也不是 `filtered` 的值同样会被拒绝。

完整设置指南见 [Docker 共存]({{ '/docs/features/docker/' | relative_url }})。

这个列表与下面的 `routing.networks` 都存放 **CIDR 网络**——`192.168.1.0/24`，而不是 `192.168.1.5`。空行或以 `#` 开头的行是注释，与地址列表编辑器中一样，会被跳过。其他任何内容都会让守护进程停止并点名该条目，在启动时和 `SIGHUP` 时都是如此。

> **这两个列表写入文件时，没有任何东西检查它们。** 只有网络页面会检查，而且只在输入时检查——于是手工编辑进去的条目到达内核时根本不成其为规则。当 `routing.mode = "networks"` 且 `networks = ["10.8.0.0/24", "10.9.0.0-24"]` 时，守护进程启动时毫无警告。`forward` 链建立后，第一个网络有接受规则，第二个没有，其流量随后被丢弃策略摧毁。页面则出现了相反的失败：它用地址列表的规则校验，而那种规则接受裸地址。于是 `192.168.1.5` 在那里通过，却被 core 拒绝，并被报告为 *保存更改失败。请检查 core 连接。*

### `[routing]`

本机转发而非接收的流量——接口之间、从容器出去、进入已发布的容器端口。

| 键 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `mode` | string | `"closed"` | `closed`、`networks` 或 `open`——见下文 |
| `networks` | list | `[]` | 允许被路由的 CIDR，双向皆可。仅在 `mode = "networks"` 下读取 |

| `mode` | 什么会经过 `forward` 链 |
|---|---|
| `closed` | 除 `[docker]` 允许的内容外，什么都不通过。适用于普通服务器 |
| `networks` | 此外，源*或*目标位于 `networks` 中的任何内容 |
| `open` | 所有内容。easywall 只过滤到达本主机的内容 |

> **关闭的 `forward` 链与未过滤的链并不相同。** 规则不给出裁决的基础链会落到它的策略上，因此带有 `policy drop` 的空链会摧毁每一个被路由的数据包。这包括另一个表的转发链已经接受的数据包。直到 2.5.0，这是唯一的行为，而且没有任何说明，这意味着 easywall 一应用，每个 Docker 容器就会失去网络。现在无论 `mode` 如何，Docker 的网络都会通过；其他需要路由的内容则依赖这个键。

### `[firewall]` —— 防护模块

每个模块一行：开启它的开关、可用来调优它的键，以及它自己的日志开关。每个阈值都是**按源地址**计。所有布尔值的默认值如「开关——默认」列所示；数字则是表中显示的默认值。

每个模块实际丢弃什么，见[防火墙过滤器]({{ '/docs/features/filters/' | relative_url }})；这里只是键的参考。

| 模块 | 开关——默认 | 调优 | 日志 |
|---|---|---|---|
| SSH 暴力破解 | `ssh_brute_force` —— **开启** | `ssh_brute_force_connection_limit` `5`/分钟 | `ssh_brute_force_log`、`ssh_brute_force_log_limit` `60`/分钟 |
| 响应 ping | `icmp_allow_echo_request` —— **开启** | —— | —— |
| ICMP 洪泛 | `icmp_flood` —— **开启** | `icmp_flood_connection_limit` `10`/秒 | `icmp_flood_log`、`icmp_flood_log_limit` `60`/分钟 |
| SYN 洪泛 | `syn_flood` —— **开启** | `syn_flood_limit` `100`/秒 | `syn_flood_log` |
| 端口扫描 | `port_scan` —— **开启** | —— | `port_scan_log` |
| 无效数据包 | `drop_invalid_packets` —— **开启** | —— | `drop_invalid_packets_log` |
| 分片（IPv4） | `drop_fragments` —— 关闭 | —— | `drop_fragments_log` |
| 无效地址过滤（IPv4） | `bogon_filter` —— 关闭 | —— | `bogon_filter_log` |
| 连接数限制 | `connection_limit_per_ip` —— 关闭 | `connection_limit_max` `100` | —— |
| TCP RST 洪泛 | `tcp_rst_flood` —— 关闭 | `tcp_rst_flood_limit` `100`/秒 | `tcp_rst_flood_log` |
| 广播 | `drop_broadcast` —— 关闭 | —— | —— |
| 组播 | `drop_multicast` —— 关闭 | —— | —— |
| 任播 | `drop_anycast` —— 关闭 | —— | —— |

不含 `icmp_allow_echo_request` 的文件——2.25 之前的每一个 `easywall.toml`——会将其读作开启。

上面的每个数字都有允许范围，而把守范围的是守护进程。选项页面上的 `max` 和 JSON Schema 中的 `maximum` 只是给浏览器和编辑器的提示。两者都管不到 `curl` 或手工编辑的文件：

| 键 | 范围 | 默认值 |
|---|---|---|
| `ssh_brute_force_connection_limit` | 1–100 | 5 |
| `ssh_brute_force_log_limit` | 1–10000 | 60 |
| `icmp_flood_connection_limit` | 1–1000 | 10 |
| `icmp_flood_log_limit` | 1–10000 | 60 |
| `syn_flood_limit` | 1–10000 | 100 |
| `tcp_rst_flood_limit` | 1–10000 | 100 |
| `connection_limit_max` | 1–100000 | 100 |
| `log_blocked_connections_limit` | 1–10000 | 60 |
| `log_blocklist_connections_limit` | 1–10000 | 60 |
| `log_feed_connections_limit` | 1–10000 | 60 |

文件中超出范围的值会被限制并记录日志；从界面传入的超出范围的值会被拒绝并点名该键——与 `acceptance.duration` 相同的分流方式。

> **直到现在都没有任何上限，而这些数字会进入 32 位字段。** 所以过大的值不会失败，而是回绕。对着内核实测：`connection_limit_max = 5000000000` 到达时变成 `ct count over 705032704`，`4294967296` 到达时变成 `ct count over 0`——一条匹配来自每个源的每个连接并丢弃它的规则。一个数字，在宣称不会把你锁在门外的产品的页面上输入，却没有任何日志。

还有三个日志开关不属于任何模块，也在这里设置：

| | 记录什么 | 速率 |
|---|---|---|
| `log_blocked_connections` | 最终策略丢弃的一切 | `log_blocked_connections_limit` `60`/分钟 |
| `log_blocklist_connections` | 拦截列表命中，在丢弃之前 | `log_blocklist_connections_limit` `60`/分钟 |
| `log_feed_connections` | 列表源命中，在丢弃之前 | `log_feed_connections_limit` `60`/分钟，每个列表源和地址族 |

---

## easywall-web（`/etc/easywall/web.toml`）

### 顶层键

| 键 | 类型 | 说明 |
|---|---|---|
| `bind_addr` | string | 监听地址与端口——例如 `"0.0.0.0:12227"` 或 `"127.0.0.1:12227"` |
| `socket_path` | string | core Unix socket 的路径——必须与 `easywall.toml` 一致 |
| `ssl_dir` | string | 自动生成的 TLS 证书与私钥存放目录 |
| `data_dir` | string | 数据目录，默认 `/var/lib/easywall`。本进程把状态保存在其中的 `web/`：通行密钥、TOTP 重放防护、版本缓存、安装标识、国家/地区表 |
| `language` | string | 备用界面语言——任何 `locales/` 中存有语言目录的语言代码；随附 `en`、`de`、`fr`、`zh`，即 English, Deutsch, Français, 简体中文。仅当浏览器请求 easywall 不支持的语言、且界面中尚未做出选择时使用 |
| `session_key` | string | 用于签名会话 cookie 的十六进制密钥——`openssl rand -hex 32`，即 64 个字符。可选：若密钥缺失、短于 32 个字符，或仍是随附的占位值，首次启动时会生成一个并写回这里 |
| `username` | string | 登录用户名——通过初始化向导设置 |
| `password` | string | Argon2id 哈希——通过初始化向导设置，不要手工编辑 |
| `totp_secret` | string | 第二因子的 Base32 共享密钥，由界面写入——为空表示未登记任何第二因子。要实现仅密码登录，清空此项、`recovery_codes` 和 `<data_dir>/web/passkeys.json`，然后重启 |
| `recovery_codes` | array of strings | 八个一次性恢复码的 Argon2id 哈希——绝不是恢复码本身，恢复码只显示一次。每使用一个恢复码，就删除一个条目 |
| `update_check` | bool | 每天向 github.com 询问一次是否有更新的发布版本——默认 `true`。六种可能的外发请求之一，也是默认开启的两种之一；见下文 |
| `telemetry` | bool | 遥测：本次安装是否可以被计入统计——除非开启，否则关闭；首次运行时会询问。见下文 |
| `notify_kind` | string | 通知传输方式——`"webhook"`、`"ntfy"`，或 `""` 表示关闭，即默认值。见[每一次离开主机的请求](#every-request-that-leaves-the-host) |
| `notify_url` | string | 通知发布到哪里。这是一份凭据：知道 ntfy 主题的人都能读取它。会拒绝重定向 |
| `notify_on_rolled_back` | bool | 一次应用未被确认、自行撤销时通知——默认 `false` |
| `notify_on_accepted` | bool | 一次应用被确认时通知——默认 `false` |
| `notify_on_panic` | bool | 恐慌模式启用或结束时通知——默认 `false` |
| `notify_on_failed_logins` | bool | 同一地址反复登录失败时通知——默认 `false` |
| `country_lookup` | bool | 在**被拦截的流量**页面上标出每个被拒绝来源登记在哪个国家/地区——默认 `true`。开启时每天下载 ipverse 的列表；设为 `false` 则不发请求。见[每一次离开主机的请求](#every-request-that-leaves-the-host) |
| `demo_mode` | bool | 针对内存中的模拟对象运行，而不是 core。仅用于公开演示——绝不要用在你正在保护的主机上 |
| `trusted_proxies` | array of strings | 其 `X-Forwarded-For` 头会被采信的地址与网络。默认空，表示以 TCP 对端为准。列出条目要付出什么代价，见[位于反向代理之后](#behind-a-reverse-proxy) |
| `health_allow` | array of strings | 可以读取 `/healthz` 的地址与网络。默认仅回环。见[谁可以读取 /healthz](#who-may-read-healthz) |

### 谁可以读取 `/healthz` {#who-may-read-healthz}

```toml
health_allow = ["127.0.0.1/8", "::1/128", "10.20.0.0/24"]
```

`/healthz` 是唯一无需会话即可应答的路由，因为询问 easywall-web 是否活着的编排器并不持有会话。它以 JSON 返回 `ok`、`degraded` 或 `fail`，不涉及你的任何规则。

| 应答 | HTTP | 含义 |
|---|---|---|
| `ok` | 200 | 正在过滤，且自检结果一致 |
| `degraded` | 200 | 仍在过滤，但有需要查看的地方 |
| `fail` | 503 | 未在过滤，或无法联系到 core |

`degraded` 刻意返回 200：若因为它重启容器，原本正常工作的防火墙会丢弃穿过它的每一条已建立连接。应改为对响应体中的 `state` 告警。

除非你放宽，否则仅限回环。列表匹配的是 TCP 对端，绝不匹配 `X-Forwarded-For`，所以 easywall 前面的代理无法把监控主机带进来——要把代理列进去。空列表会关闭该端点；省略该键则意味着回环默认值，因此升级后仍可继续工作。

### 位于反向代理之后 {#behind-a-reverse-proxy}

```toml
trusted_proxies = ["127.0.0.1", "10.1.0.5"]
```

每个条目都是一个地址或 CIDR 网络，easywall 采信其 `X-Forwarded-For`。默认空，表示以 TCP 对端为准。

**列入此列表就是对那个对端的完全信任。** 列出代理本身，绝不要列出它们所在的网络：否则 `10.0.0.0/8` 中的每台主机都能自选 easywall 记录、据以判定锁定并施加速率限制的地址。

整套做法——在前面放 nginx、该用哪个值、Docker 的情形，以及如何确认它已生效——见[位于反向代理之后]({{ '/docs/installation/reverse-proxy/' | relative_url }})。

**如果列出的代理不发送 `X-Forwarded-For`，**请求仍然会被解析——解析为代理自己的地址。审计日志中会标记为 `via-proxy`，应用页面则会报告为 *cannot tell*。这个标记正是该版本最可能出现的配置错误的症状：代理被加入 `trusted_proxies`，但代理自身的配置里没有对应的 `proxy_set_header X-Forwarded-For`。此后它背后每个客户端都共享那一个地址的速率限制桶，与列表存在之前一样——设置该头，问题就会消失。

### 界面如何选择语言

优先级从高到低：

1. **界面中的明确选择。** 侧边栏页脚的切换器——登录页上也有，这样读不懂当前语言的运维人员仍能登录——会把 `easywall_lang` 保存一年。它的优先级高于下面所有项：浏览器头描述的是机器，而它描述的是使用机器的人。
2. **浏览器的 `Accept-Language` 头。**
3. **配置中的 `language`**，即上面的设置。
4. **English**。

界面提供哪些语言，取决于 `locales/*.json` 里有什么；每个文件通过自己的 `language_name` 键声明自身的名称。因此 `Deutsch` 读作 `Deutsch`，`Français` 读作 `Français`，无论界面当前使用哪种语言。尚未有人审校的语言会在切换器中注明这一点，覆盖率报告则会说明它已有多少内容。要让它出现，只需添加一个语言区域文件；见[添加一种语言]({{ '/docs/contributing/' | relative_url }}#adding-a-language)。

```bash
openssl rand -hex 32     # session_key
```

**`session_key` 必须保密。** 持有它的人可以伪造有效的会话 cookie——无需密码。若密钥缺失、过短，或仍是示例配置随附的占位值，easywall 会在首次启动时生成一个，并写回 `web.toml`。

> **不存在 `csrf_key`。** CSRF 防护由 Go 1.25 的 `net/http.CrossOriginProtection` 提供，它检查 `Origin` 和 `Sec-Fetch-Site`，而不是签发令牌。旧配置遗留下来的 `csrf_key` 不会被任何东西读取。

### `[tls]`

两个键都留空，就会使用 `ssl_dir` 中自动生成的自签名证书。

| 键 | 说明 |
|---|---|
| `cert` | 自定义 TLS 证书 PEM 文件的绝对路径（例如 Let's Encrypt fullchain） |
| `key` | 匹配的私钥 PEM 文件的绝对路径 |
| `hostname` | 访问本次安装所用的名称。一个字段，两个用途：ACME 签发证书所用的域名，以及通行密钥的 WebAuthn Relying Party ID |
| `acme` | 通过 ACME 自动获取并续期证书（默认为 Let's Encrypt）。需要 `hostname` 与 `acme_agree_tos`；与 `cert`/`key` 互斥 |
| `acme_email` | 可选的联络地址，证书颁发机构会向它发送到期警告 |
| `acme_agree_tos` | 你对证书颁发机构用户协议的同意。**没有默认值**——easywall 绝不会替你同意第三方的条款，因此未设置此项时 `acme = true` 会在启动时被拒绝 |
| `acme_directory` | ACME 目录 URL 覆盖项。留空即 Let's Encrypt 生产环境；在把 DNS 和 80 端口弄对期间，可指向预发布端点 |

`acme = true` 还会在 80 端口上开启一个固定的 HTTP-01 监听器。关于它要能应答必须满足哪些前提——包括 easywall 不会替你做的那一件事——见[安全 → 传输]({{ '/docs/security/' | relative_url }}#transport)。

自动生成的证书有效期为一年，一旦进入到期前 30 天内就会被替换。检查发生在启动时，以及服务运行期间每天两次，因此一台运行时间超过自身证书有效期的服务器仍能继续工作。

自定义证书**绝不**被覆盖。文件变化时它会重新读取，因此在原地续期的 ACME 客户端无需重启，会在下一次连接时生效。

`cert` 与 `key` 要么都设置，要么都不设置。只设置其一会在启动时被拒绝：否则 easywall 会把你的文件与它自己生成的那对中的另一半配对。TLS 随后会因密钥不匹配而失败，错误里点名的证书你从未配置过。

## 每一次离开主机的请求 {#every-request-that-leaves-the-host}

共六项，这就是完整清单。

| | 更新检查 | 安装计数 | 通知 | 列表源 | 国家/地区查询 | 证书 |
|---|---|---|---|---|---|---|
| 键 | `update_check` | `telemetry` | `notify_kind` | 无——界面中的**拦截列表** | `country_lookup` | `tls.acme` |
| 默认值 | **开启** | **关闭**，直到你开启 | **关闭**，直到你开启 | **关闭**，每个列表源各自 | **开启** | **关闭**，直到你开启 |
| 目标 | `api.github.com` | `telemetry.wdkro.de` | `notify_url`——你的，不是我们的 | 每个列表源自己的地址，某个国家的在 `raw.githubusercontent.com` 上；自有列表源的地址归你 | `github.com`，重定向到 `release-assets.githubusercontent.com` | `tls.acme_directory`，未设置时为 Let's Encrypt |
| 频率 | 每天一次 | 每天一次 | 有事发生时 | 按列表源自己的时间表，从每小时到每天 | 每天一次 | 首次需要时，之后在到期前 |
| 携带内容 | 与你无关——只为获取最新版本发起的一次普通 GET | 一个随机标识和版本号，详见下文 | 你的防火墙发生了什么 | 一个 GET；自有列表源的用户名与密码发往它自己的 URL | 一个 GET | `tls.hostname`、在此生成的账户密钥，以及设置了的话 `tls.acme_email` |
| 关闭方式 | `update_check = false` | `telemetry = false`，或界面中的**系统** | `notify_kind = ""`，或界面中的**通知** | 该列表源的开关 | `country_lookup = false` | `acme = false`——easywall 随后自行签发证书 |

默认开启的有两项：更新检查与国家/地区查询。

`notify_url` 和自有列表源的 URL 是 easywall 唯一没有指定的目标。其中的主机部分是你可以替换的默认值，其余都是固定的。正因此，通知地址被当作凭据，以 `0600` 权限保存在 `web.toml` 中，与 `session_key` 并列；也正因此，自有列表源的密码绝不会被回显。

前五项从不会拖慢页面。更新检查由磁盘上的缓存提供，并在后台刷新。失败会被记住一小时，这样没有对外路由的主机就不会在每次加载时重试。计数在后台运行，十秒后放弃。列表源同样在后台刷新，刷新失败会保留已经加载的副本；国家/地区查询也是如此。唯一可能阻塞页面的是证书，因为页面正是由它提供的。

### 更新检查

存在更新的发布版本时会出现横幅。关掉它不会改变其他任何东西——无论开关如何，都会显示 easywall 正在运行的版本。

### 安装计数 {#counting-installations}

初始化向导会询问，而不是替你假定。严重缺陷在十个安装和一万个安装中的分量不同。计数是唯一能知道属于哪种情况的办法——也是唯一能说明某个修复已覆盖其中大多数的办法。

它发送的内容，完整如下——每天一次，除此之外永远没有别的：

```
GET https://telemetry.wdkro.de/v1/count?id=<32 hex characters>&v=<version>
```

| | |
|---|---|
| **不**发送 | 主机名、任何地址、任何规则、你配置内容的任何数量 |
| 标识符 | 在你机器上生成的 16 个随机字节，位于 `<data_dir>/web/telemetry.json`。删除该文件，下一次上报在任何人看来都会是一个新安装 |
| 为何随机 | 由主机名或 machine-id 派生的值，任何知道该主机的人都能复现，这会把计数变成一次查询 |
| 在远端 | 一行——时间戳、标识符、版本——以及一个 204。**不记录你的地址**：它只用于对该端点做速率限制，永远不会落盘。记录保留 35 天，之后只剩汇总后的数字 |

关闭它并不需要 core 进程正在运行——只有在另一个守护进程可联系时才能撤回的同意，算不上同意。

#### 更新后再次询问

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/telemetry-ask" ext="png"
     alt="仪表盘顶部的一张卡片，询问本次安装是否可以被计入统计，并展示了它会发送的请求，以及「可以，计入统计」「不用了，谢谢」两个按钮和一个链接样式的「不再询问」。" %}
</figure>

| | |
|---|---|
| 何时 | **次要或主要**更新（2.27 → 2.28）后的第一个仪表盘，且计数处于关闭状态。补丁版本（2.28.0 → 2.28.1）不会询问 |
| 从不询问 | 计数已开启 · 已选择 *不再询问* · 已设置 `EASYWALL_WEB_TELEMETRY`（由部署决定）· 演示模式 |
| 一个回答 | *可以，计入统计*、*不用了，谢谢*、初始化向导以及**系统**下的开关，都算作针对此版本的回答 |
| 保存在哪里 | `<data_dir>/web/telemetry_prompt.json`。删除它，下一次仪表盘会再次询问 |

它是一张卡片，不是挡路的页面：登录以回滚某项更改时从不需要等它。

> 这个数字是下界，无法做到防篡改：端点是开放的，任何人都能编造标识。它足以区分十个安装与一万个安装，也足以看出某个修复是否已经扩散。它不适用于其他任何用途，也没有被这样宣称。

---

## 编辑器自动补全

两个文件都随附一份 JSON Schema。把 [Taplo](https://taplo.tamasfe.dev/) 指向它们，即可在 VS Code、Neovim 以及任何使用 LSP 的编辑器中获得内联校验：

```toml
# taplo.toml (project root)
[[rule]]
include = ["config/easywall.toml"]
url = "https://easywall-project.org/schemas/easywall.schema.json"

[[rule]]
include = ["config/web.toml"]
url = "https://easywall-project.org/schemas/web.schema.json"
```

Schema 直接链接：

- [easywall.schema.json]({{ '/schemas/easywall.schema.json' | relative_url }})
- [web.schema.json]({{ '/schemas/web.schema.json' | relative_url }})
