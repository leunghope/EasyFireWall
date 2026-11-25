---
layout: default
title: 防火墙过滤器
description: 这些防护模块 —— 每一个丢弃什么、如何调整，以及哪些默认开启。
---

# 防火墙过滤器

可选模块，在开关端口之外进一步加固主机。它们全都是 `table inet easywall` 里的原生 nftables 规则 —— 没有子进程，没有可注入的东西。

选项会立即保存到守护进程的配置中，并在下次[应用]({{ '/docs/features/apply/' | relative_url }})时进入内核。自 2.10 起这算作未应用的更改：仪表盘上的*未应用的更改*标签会计入它，应用页面会按它的键名列出，如 `drop_fragments off → on`。

## 该开启哪些 {#which-to-turn-on}

大多数人带着这个问题来。每个模块的细节在下面。

| 主机 | 开启 | 保持关闭 |
|---|---|---|
| 公网服务器，固定地址 | 攻击防护下的全部模块（分片丢弃除外），外加无效地址过滤 | 分片丢弃 —— 它会破坏大型 DNS 应答，[见下文](#what-fragment-drop-breaks) |
| 位于 NAT 之后，或在局域网中 | SSH 暴力破解、SYN 泛洪、端口扫描检测、无效数据包。自己的网络加入允许列表之后，再加上无效地址过滤 | 丢弃广播/组播/任播 |
| 容器主机 | 默认值即可。开启 Docker 共存时无效地址过滤是安全的 —— 网桥网络会被豁免 | — |

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/options" ext="png"
     alt="防火墙选项页面：一张卡片说明三类主机各自需要打开哪些开关，其下是两张卡片，攻击防护与流量过滤，每个防护模块一行：名称与作用、参数、最右侧的开关，以及一个收起的「这会改变什么？」披露面板。已开启的模块沿左侧带一道描边。" %}
  <figcaption>在这里拨动开关会立即暂存改动 —— 内核要到下次应用才会看到它。</figcaption>
</figure>

## 模块所在的位置

在拦截列表之前，在允许列表之前，在考虑任何端口之前。某个模块丢弃了一个数据包，无论你还放行了什么，它都会被丢弃。

还有两件事更早：回环，永远最早，以及 [IPv6 模式]({{ '/docs/features/system-settings/' | relative_url }}) —— 设为 `passthrough` 或 `block` 时，IPv6 在任何模块看到它之前就已经被决定。

有两个模块紧跟在 IPv6 模式之后、在返回流量之前：**ICMP 泛洪**和 **TCP RST 泛洪**。对 conntrack 来说，某个来源的第一个 ping 之后的每一个 ping、活跃连接的每一个 reset，*都*是返回流量，而那条 accept 会让它们不受计量地通过。**分片丢弃**在这一切之前运行，位于自己的链中 —— 见[它会破坏什么](#what-fragment-drop-breaks)。

{% include themed-figure.html base="/assets/diagrams/rule-order" ext="svg"
   alt="入站数据包的决策流程：先是分片丢弃（开启时）；然后是回环；然后是 IPv6 模式，除非设为 filter，否则它直接接受或丢弃全部 IPv6；然后是 ping 与 reset 的速率限制，然后是已建立的连接与邻居发现，然后是其他防护模块，然后是 Docker 网桥网络，然后是丢弃的拦截列表，然后是接受所有端口的允许列表，然后是你开启的列表源执行丢弃，然后是 ping（响应 ping 开启时），然后是开放端口，然后是自定义规则，最后是链策略丢弃。" %}

## 始终开启 {#always-on}

编译进每一套规则。这些没有开关。

| 规则 | nftables | 原因 |
|---|---|---|
| 默认丢弃 | `input` 上的 `policy drop` | 默认拒绝 |
| 回环 | `iif lo accept` | 本机进程必须能互相访问 |
| 返回流量 | `ct state {related, established} accept` | 对你发起的连接的回包，以及关于它的每一个 ICMP 错误 —— 不可达、需要分片、超时 |
| ICMPv6 发现 | 类型 133–136，启用时 | 邻居发现与地址自动配置 —— 见[网络设置]({{ '/docs/features/system-settings/' | relative_url }}) |

除此之外没有其他 ICMP 类型会被单独接受，[响应 ping](#answering-pings) 除外。只有当本机发出了那个 echo 时，echo 应答才算返回流量；也只有当错误是关于 conntrack 跟踪的连接时，它才算。主动送来的应答、与任何连接无关的错误、时间戳请求，都会撞上策略丢弃。

## 响应 ping {#answering-pings}

| 开关 —— 默认 | 接受 | 排在它之前 |
|---|---|---|
| `icmp_allow_echo_request` —— **开启** | ICMP 类型 8 与 ICMPv6 类型 128 —— 后者仅在 [IPv6]({{ '/docs/features/system-settings/' | relative_url }}) 设为 filter 时 | *ICMP 泛洪*，会先把超出速率的来源丢弃；拦截列表与列表源，会在任何 ping 被应答之前就拒绝某个来源 |

关闭后，本机不应答任何 ping —— 但 `ipv6.mode = "passthrough"` 下的 IPv6 除外，那种模式会在任何规则之前接受全部 IPv6。它自己发出的 ping 仍会收到应答，路径 MTU 发现和 traceroute 也仍然可用：那些属于返回流量。

**只让监测系统 ping：** 关掉它，并把监测地址加入[允许列表]({{ '/docs/features/allowlist/' | relative_url }})。允许列表会接受来自该地址的一切 —— 每一个端口，而不只是 ping。

## 三条链

以上全部属于 `input` 链 —— 发往本机的流量。easywall 还会创建另外两条链，它们的行为有一些值得了解的后果。

| 链 | 策略 | 谁会进入 |
|---|---|---|
| `input` | **丢弃** | 发往本机的流量。本页讲的全部内容。 |
| `output` | **接受** | 本机发出的流量。easywall 不过滤它。 |
| `forward` | 默认 **丢弃** | 本机要*路由*的流量 —— 两个接口之间、从容器出来、进入已发布的容器端口。由 `[routing]` 支配 |

> **关闭和不加过滤不是一回事。** 基本链的规则若未给出判决，就会落到它的策略上。一条空的 `forward` 链配上 `policy drop`，会在该钩子处销毁每一个被路由的数据包。这包括已经被另一张表接受的包。这个丢弃是最终的：你自己的 forward 链也无法推翻它，而[自定义规则]({{ '/docs/features/custom-rules/' | relative_url }})会进入 `input`。在普通服务器上毫无代价；在 2.5.0 之前，它让每个 Docker 容器都彻底停摆。

有两类流量会穿过这条链：

- 你已放行的 [Docker]({{ '/docs/features/docker/' | relative_url }}) 网络，无论 `routing.mode` 怎么说。开启共存本身就是在声明这台主机承载容器流量。
- `[routing]` 点名的任何东西。它的三个取值 —— 不路由任何东西、只路由这些网络、不去管被路由的流量 —— 见[网络页面]({{ '/docs/features/system-settings/' | relative_url }}#the-network-page)与[配置]({{ '/docs/configuration/' | relative_url }}#routing)。

## 攻击防护 {#attack-protection}

| 模块 | 丢弃 | 调参 | 默认 |
|---|---|---|---|
| **SSH 暴力破解** | 单个来源超过速率的新建 SSH 连接。作用于[端口页面]({{ '/docs/features/ports/' | relative_url }})上标记为*SSH 保护*的端口；若一个都没有标记，则作用于 22 | `ssh_brute_force_connection_limit` —— 5/min | **开启** |
| **响应 ping** | 不丢弃任何东西 —— 它接受 ICMP 类型 8 与 ICMPv6 类型 128；[ICMP 还会经历什么](#answering-pings) | — | **开启** |
| **ICMP 泛洪** | 单个来源超过速率的 echo 请求 —— ICMP 类型 8 与 ICMPv6 类型 128 | `icmp_flood_connection_limit` —— 10/s | **开启** |
| **SYN 泛洪** | 单个来源超过速率的新建 TCP 连接 | `syn_flood_limit` —— 100/s | **开启** |
| **端口扫描检测** | 七种不可能的 TCP 标志组合：NULL、单独 FIN、SYN+FIN、RST+FIN、SYN+RST、XMAS 与全标志 —— 真实客户端一种都不会发 | — | **开启** |
| **无效数据包** | conntrack 无法匹配到连接的数据包 | — | **开启** |
| **分片丢弃** | 发往本机的分片 **IPv4** 数据包，在重组之前 —— [它会破坏什么](#what-fragment-drop-breaks) | — | 关闭 |
| **无效地址过滤** | 非回环接口上不可能的 **IPv4** 源地址 | — | 关闭 |
| **连接数限制** | 单个来源超过上限的并发连接 | `connection_limit_max` —— 100 | 关闭 |
| **TCP RST 泛洪** | 单个来源超过速率的入站 RST 数据包，无论对应连接是否存在 | `tcp_rst_flood_limit` —— 100/s | 关闭 |

> **每个速率都按源地址分别计数** —— 每个地址一个内核计数器，放在一个集合里，该来源安静下来后条目就过期。来自一台主机的泛洪花不掉让你保持连接的那份额度。
>
> 2.5.0 之前并非如此：四个模块整台机器共用一个计数器，于是来自任何地方每分钟五次 SSH 尝试，也会把管理员自己锁在外面。

每个模块都在允许列表之前运行，因此对已加入允许列表的地址也照常生效。只有无效地址过滤会豁免其中一些。

### 分片丢弃会破坏什么 {#what-fragment-drop-breaks}

内核会在 `input` 链看到数据包之前先把它重组。所以这个模块有自己的链 `fragments`，位于 prerouting 钩子上，在重组之前。它丢弃发往本机的 IPv4 分片。IPv6、回环和本机路由的流量不受影响。转发到本机地址的端口转发算作发往本机的流量。

这样一来，任何大到装不进一个数据包的应答都会被破坏。多数主机会遇到的是承载 DNSSEC 的 UDP DNS 应答，而 [RFC 8900](https://www.rfc-editor.org/rfc/rfc8900#section-6.5) 要求运维者不要过滤进出 DNS 服务器的分片。2.22 之前这条规则待在 `input` 里，什么都匹配不到。

另一种是隧道。WireGuard 走 UDP，当路径 MTU 小于它的数据包时，每个包到达时都是分片，随即被丢弃。一台生产主机在 2.23 上量到了这一点：一秒钟内来自它自己的 WireGuard 对端的 60 个分片。如果隧道是你进入这台主机的方式，就保持关闭，或调低隧道的 MTU。

### 无效地址过滤会丢弃什么 {#what-the-bogon-filter-drops}

一个声称来自下列网段、却从真实接口进来的数据包是伪造的 —— 公网上没有任何东西合法地拥有这样的源地址。

| 网段 | | 网段 | |
|---|---|---|---|
| `0.0.0.0/8` | 「本网络」 | `172.16.0.0/12` | 私有 |
| `10.0.0.0/8` | 私有 | `192.0.2.0/24` | 文档 |
| `100.64.0.0/10` | 运营商 NAT | `192.168.0.0/16` | 私有 |
| `127.0.0.0/8` | 回环 | `198.51.100.0/24` | 文档 |
| `169.254.0.0/16` | 链路本地 | `203.0.113.0/24` | 文档 |
| | | `240.0.0.0/4` | 保留 |

**仅限 IPv4。** 没有对应的 IPv6 版本，而且那也不会是这个过滤的翻译版本：`fe80::/10` 是链路本地，IPv6 必须要靠它上面的邻居发现才能运作。

### 它不会丢弃什么 {#what-it-does-not-drop}

允许列表上的任何东西，以及任何 Docker 网桥网络，都会被豁免。两者都是 RFC 1918 地址列表，而这恰恰是这个过滤要丢弃的东西。它在两者之前运行，所以过去一开启它，就等于不声不响地关掉了这两个功能。把 `192.168.1.0/24` 加入允许列表毫无作用，让 Docker 的 `172.17.0.0/16` 通过也一样。

豁免的范围很窄：它只覆盖你放行的部分，别的都不管。允许列表里写 `192.168.1.0/24`，`192.168.0.0/16` 的其余部分仍会被丢弃。端口规则的[来源]({{ '/docs/features/ports/' | relative_url }}#who-may-reach-it)同样豁免，但只针对该规则的协议和端口。

```
# nft list chain inet easywall bogon
ip saddr 192.168.1.0/24 return      ← allowlisted
ip saddr 172.17.0.0/16 return       ← Docker bridge
ip saddr 10.9.0.0/24 tcp dport 2222 return   ← a port rule's source
ip saddr 10.0.0.0/8 drop
ip saddr 192.168.0.0/16 drop        ← the rest of the range, still dropped
...
```

| 开启之前 | 原因 |
|---|---|
| 把你用来管理主机的地址加入允许列表（若它属于 RFC 1918） | 豁免只覆盖规则应用时已在列表上的内容 |
| 不要在 DHCP 服务器上开启 | 请求租约的客户端还没有地址，会从 `0.0.0.0` 发出，而它正好被这个过滤丢弃 |

## 流量过滤 {#traffic-filtering}

| 模块 | 丢弃 | 默认 |
|---|---|---|
| **丢弃广播** | 发往广播地址的流量 | 关闭 |
| **丢弃组播** | 发往组播组的流量 | 关闭 |
| **丢弃任播** | 发往任播目标的流量 | 关闭 |

> **不要在局域网里开。** 在地址固定的对外主机上丢弃是安全的；在其他几乎所有场景都有破坏性。

各自会破坏什么：

- **广播** —— DHCP 服务器以广播形式发给续租客户端的应答。`systemd-networkd` 上首次获取租约仍然能成功：它是从包套接字读的，在防火墙看到之前。其他 DHCP 客户端未实测。
- **组播** —— mDNS（Avahi、`.local` 名称）、SSDP 和 DLNA，以及其他一切组播。IPv6 邻居发现不受影响：它更早就被决定了，见[始终开启](#always-on)。
- **任播** —— 只影响发往内核归类为任播的地址的流量。多数主机没有这样的地址，对它们来说这个开关什么也不会改变。

## 日志 {#logging}

每个模块都有自己的 `*_log` 开关，外加三个全局开关。所有日志都限速，消息每分钟条数里的那些 `*_limit` 值就是为此 —— 泛洪不能把磁盘写满。

| 开关 | 记录 | 前缀 |
|---|---|---|
| `ssh_brute_force_log` | 超出 SSH 速率的连接 | `easywall ssh:` |
| `icmp_flood_log` | 超出速率的 echo 请求 | `easywall icmp-flood:` |
| `syn_flood_log` | 超出速率的 SYN | `easywall syn-flood:` |
| `tcp_rst_flood_log` | 超出速率的 RST | `easywall tcp-rst:` |
| `port_scan_log` | 扫描标志组合 | `easywall portscan:` |
| `drop_invalid_packets_log` | INVALID 状态的数据包 | `easywall invalid:` |
| `drop_fragments_log` | 分片数据包 | `easywall fragment:` |
| `bogon_filter_log` | 无效源地址 | `easywall bogon:` |
| `log_blocklist_connections` | 拦截列表命中，在丢弃之前 | `easywall blocklist:` |
| `log_feed_connections` | 列表源命中，在丢弃之前 | `easywall feed: <id>` |
| `log_blocked_connections` | 最终策略丢弃的一切 | `easywall drop:` |

这些开关记录的内容都会出现在[被拦截的流量]({{ '/docs/features/blocked-traffic/' | relative_url }})页面上，最新的在最前，可筛选，点一下就能得到接下来要写的那三条规则。

**自 2.21 起，这不是内核日志。** 规则通过 NFLOG 把每个数据包发给 easywall-core，组默认是 `12227`，而 `journalctl -k | grep easywall` 什么也返回不了。取而代之：

| 你想 | 就 |
|---|---|
| 看看 | 在界面中打开**拦截记录** |
| 拿一个文件来 `jq` 或运到别处 | 设置 [`[packet_log] persist = true`]({{ '/docs/configuration/' | relative_url }}#packet_log)，然后 `tail -f /var/log/easywall/packets.log \| jq .` —— 每行一个 JSON 对象 |
| 让 ulogd2 拿到 | 给 ulogd2 换一个组；一个组在每个网络命名空间中只能绑定一次（对 host 网络安装而言是每台主机一次） |

两项代价，明说而不是等你去发现：

- **easywall-core 停止时，被记录的数据包会被丢弃。** 内核日志无论是守护进程在不在都会写；NFLOG 只投递给监听者。
- **如果这个组无法绑定** —— ulogd2 已经占用了它 —— core 会在启动时说明，拦截记录页面也会说明。此时日志回落到内核日志，上面那条 `journalctl` 命令又管用了。把 `nflog_group` 设成一个空闲的组并重启。

你自己[自定义规则]({{ '/docs/features/custom-rules/' | relative_url }})里的 `log` 语句仍然写内核日志，除非它点名 `group 12227`。

每条日志规则都紧挨着它所属的丢弃规则，并且带同样的匹配条件，所以页面上出现的正是被丢弃的东西。

> **2.5.0 之前这一整套都不工作。** 八个开关根本没生成规则，唯一生成的那个也没带前缀。日志表达式的 `Key` 字段是个位掩码，而它被设成了一个裸属性编号。于是内核拿到的是空的日志组。无论开了什么，上面那条命令都匹配不到东西。

这是*内核*日志 —— 数据包。管理性改动在[审计日志]({{ '/docs/features/audit-log/' | relative_url }})里。

界面写出的就是你手工要编辑的同一段 `[firewall]`；每个键都列在[配置]({{ '/docs/configuration/' | relative_url }}#firewall--protection-modules)下。