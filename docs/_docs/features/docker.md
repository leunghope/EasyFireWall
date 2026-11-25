---
layout: default
title: 与 Docker 共存
description: easywall 只拥有一个 nftables 表，不碰其他任何东西，因此 Docker 的链在应用之后依然存在。
---

# 与 Docker 共存

easywall v1 会清空 iptables，把 Docker 的链一并带走。v2 不会：它只拥有一个表，
从不查看另一个表。

{% include themed-figure.html base="/assets/diagrams/docker-coexist" ext="svg"
   alt="easywall 只创建、清空并替换 table inet easywall，从不动 Docker 的表，也不动任何其他表。" %}

每次应用都会清空并重建 `table inet easywall`。`DOCKER`、`DOCKER-USER` 和
`DOCKER-ISOLATION` 位于别处，既不读取、也不写入或删除。

## 启用它

```toml
[docker]
enabled               = true   # detect Docker bridge interfaces
allow_bridge_networks = true   # accept traffic from the detected CIDRs
custom_networks       = []     # anything detection misses
published_ports       = "open" # "filtered": easywall decides who reaches a
                               # published container port — see option 2 below
```

检测会读取名为 `docker*` 或 `br-*` 的接口，并取每个接口的 CIDR。它在每次应用时
运行。在 `easywall-core` 启动后的最多 90 秒内，它还会自己再运行一次。这覆盖了
常见情况：开机时还不存在这样的网桥，而 Docker 是在守护进程之后启动的。超过这个
窗口之后，或者对更晚才添加的网桥，请再应用一次，或者在 `custom_networks` 里
加一条。

那里的条目是 CIDR 网络——`172.20.0.0/16`，而不是单个容器地址——允许 `#` 注释。
其他内容会被点名拒绝，而不是接受之后悄悄跳过。

> **这些网络同时也是打开 `forward` 链的原因。** 容器流量是被路由的，并不发往
> 本机。一条 `policy drop` 的空基础链会在钩子处把它丢弃，无论 Docker 自己的链
> 接受了什么。在 2.5.0 之前，这一点会让下面每一种方案都失效。现在，只要源地址
> *或*目标地址落在这些网络中，流量就会穿越它，**无论 `routing.mode` 怎么设置**：
> 否则升级一次就会把主机上的容器全部踢出网络。

> **无效地址过滤对它们作例外。** 网桥的地址段属于 RFC 1918，正是那个模块会丢弃的
> 地址。在加入这个例外之前，开启它会悄悄破坏共存。

见[防火墙过滤器]({{ '/docs/features/filters/' | relative_url }})；如果这台主机
还因其他原因做路由，另见
[`[routing]`]({{ '/docs/configuration/' | relative_url }}#routing)。

## 一起运行的四种方式

| | 方案 | 入站到已发布端口 | 从容器出站 | 适用场景 |
|---|---|---|---|---|
| **1** | `enabled = true`——出厂默认 | 是，由 Docker 发布它们 | 正常 | 大多数主机 |
| **2** | `enabled = true`、`published_ports = "filtered"` | 只有存在**转发**范围的[端口规则]({{ '/docs/features/ports/' | relative_url }})时才放行 | 正常 | 容器主机上只有一道防火墙 |
| **3** | `enabled = true`，Docker 配置为 `{"iptables": false}` | 每个端口都要有一条[端口规则]({{ '/docs/features/ports/' | relative_url }}) | **需要你自己写一条 masquerade 规则** | 一道防火墙，一处查看 |
| **4** | `enabled = false`——有意如此 | 否 | **否** | 不运行任何容器的主机 |

### 方案 2 取代的是云防火墙

已发布的端口是被路由到容器的，并不发往本机，因此它永远到达不了 `input` 链，
任何普通端口规则都看不到它。在 `published_ports = "filtered"` 下，easywall 的
`forward` 链会拒绝发往容器地址的入站流量，而范围设为 `forwarded` 的端口规则才是
打开某个端口的途径。这样，主机上的全部端口清单就集中在一处——这正是云厂商防火墙
此前替你做的事。拦截列表、列表源和国家/地区也先于这些规则被检查；允许列表只对
列表源作例外，不打开任何容器端口。

> **开启它会关闭每一个没有转发规则的已发布端口。** 先写好规则，再改这个键，
> 然后应用一次。120 秒的确认窗口在这里救不了你：它验证的是你自己的连接，而你的
> 连接走 `input` 链，容器的不走。

### 带 IPv6 的 Docker 主机

用户自定义网桥可以在它的 IPv4 子网之外再携带一个 ULA `/64`（`fd…::/64`）：

```bash
docker network create --ipv6 --subnet 172.20.0.0/16 --subnet fd00:20::/64 mynet
```

一条普通的 `-p 993:993` 本来就会在两个地址族上发布。不指定主机 IP 时，Docker 会
同时绑定 `0.0.0.0` *和* `[::]`——`docker ps` 两个都会显示。在一个启用了 IPv6 且
打开了 `ip6tables` 的网络上，它会以写 `ip` 规则同样的方式写入对应的 `ip6` DNAT
规则。`-p [::]:993:993` 只发布 IPv6；请有意识地用它把某个端口排除在 IPv4 之外，
而不是用它给一次普通发布「添加」IPv6。Ansible 的
`community.docker.docker_container` 是例外：它的 `default_host_ip` 是 `0.0.0.0`，
因此 `"25:25"` 只绑定 IPv4，直到把它设为 `""`。Docker 需要在自己的守护进程配置里
启用 `ip6tables` 才能写入 `ip6` DNAT 规则；自 Docker 27 起这已是默认值。

网桥一旦存在，检测就会在识别其 IPv4 网络的那次应用里一并识别出它的 IPv6 网络。
已经为 IPv4 打开某个端口的转发端口规则，也会为 IPv6 打开它——一条规则，两个地址族，
无需另写 IPv6 端口规则。网桥自身的 `fe80::` 链路本地地址永远不算容器网络，因此它
本身不会打开任何东西。

`ipv6.mode = "block"` 会把 IPv6 从转发链所基于的容器网络中移除：无论网桥携带什么，
都不会生成 IPv6 的对应项、拒绝规则或例外。这与在转发链里给出一个 IPv6 判决并不是
一回事——`block` 只会在 input 链上添加判决。此后发往容器的 IPv6 流量会遇到什么，
改由 `routing.mode` 决定。在 `routing.mode = "closed"` 下，它遇到的是转发链的策略
丢弃；`"networks"` 下也一样，除非 `routing.networks` 点名了一个覆盖它的网络。在
`routing.mode = "open"` 下，转发策略是接受，因此发往容器的 IPv6 流量会被原样放行，
不做过滤，交给 Docker 自己的 `ip6tables` 去处理。

IPv6 网桥上的容器可以像走 IPv4 那样，通过 IPv6 访问主机自身的服务——input 链的
网桥放行规则覆盖两个地址族。

### 发布在网桥网关上的端口仍然是已发布端口

`-p 172.17.0.1:53:53` 发布到网桥，而不是发布给整个世界，所以它读起来像是一个只有
容器才会访问的服务，`forward` 链的拒绝规则似乎管不着它。其实管得着。*另一座*
网桥上的容器访问它时，会被 DNAT 进第一座网桥，于是它到达时目标地址在这座网桥内、
源地址在网桥外。这正是来自互联网的数据包的形状——因为到那时它已经是了。这样的
端口和任何其他已发布端口一样，需要一条**转发**规则。

2.20.1 所源自的那台主机，只要当时有这句话就能避免故障。它的解析器就是这样发布的，
第一次应用后所有容器都失去了 DNS，而十五个外部探针全部保持正常。

> **指定了来源的转发规则只覆盖这些来源。** 拒绝规则关闭的是整个世界*加上*其他
> 每一座网桥。把其他网桥的网络写进去，可以恢复其中的容器访问，在只允许容器访问
> 该端口时这就是正确的规则。但它不会恢复整个世界：发往 `0.0.0.0` 已发布端口的
> 数据包会被 DNAT 进某座网桥，并以同样的形状遇到拒绝。这类规则仍会被日志点名，
> 因为没有任何来源列表是针对整个集合来校验的。

### 每次应用时日志会说什么

自 2.20.1 起，在 `filtered` 下应用时，日志会点名每一个没有转发规则覆盖的已发布
端口，这些信息读自内核中 Docker 自己的 DNAT 规则：

```
53 published on 172.17.0.1 with no forwarded rule: the forward chain drops
everything that reaches it from outside its own bridge — the world, and
containers in another bridge. Give it a port rule with scope "forwarded" —
no sources, or sources naming the other bridge networks if only containers
should reach it — or set docker.published_ports = "open"
```

一次重新映射端口的发布——`-p 8080:80`——会同时点名两个数字，因为规则必须指名
第二个。转发链在 Docker 的 DNAT 之后运行，因此数据包到达那里时携带的是 `80`，
而针对 `8080` 的规则什么也匹配不到：

```
8080 published on 0.0.0.0 reaches the container on 80: no forwarded rule for
80: the forward chain drops everything that reaches it from outside its own
bridge — the world, and containers in another bridge. Give it a port rule with
scope "forwarded" — no sources, or sources naming the other bridge networks if
only containers should reach it — or set docker.published_ports = "open"
```

| | |
|---|---|
| 每次应用都输出一次，不省略 | 重复的内容不折叠。你正在读日志的那次应用，正是必须说出这句话的那次 |
| 指定了来源的规则也会被点名 | 它只覆盖它列出的来源，而这些来源从来不是拒绝规则关闭的整个集合——见上面的提示框 |
| **两个地址族、两个后端** | 会读取 Docker 的 `ip nat` 和 `ip6 nat` DNAT 规则——如果用的是它的 nftables 后端，则读取 `ip docker-bridges` 和 `ip6 docker-bridges`——因此只发布在 IPv6 上的端口也会被点名 |
| 回环发布不会被点名 | `-p 127.0.0.1:8081:80` 只能从主机访问，而这条连接从不经过转发链 |
| 不需要 Docker socket，也不需要客户端库 | Docker 已经停止、规则却还留在内核里的主机，正是这里说的那种主机 |

| 另外值得知道的 | |
|---|---|
| 界面上没有任何地方能设置这个键 | 只能在 `easywall.toml` 里编辑。一个按下就可能把所有容器踢出网络、而确认窗口对此毫无察觉的按钮，算不上控制手段 |
| 改由[端口页面]({{ '/docs/features/ports/' | relative_url }})读取它 | 当 `published_ports` 为 `open` 时，转发规则在那里被标记为不生效。它只警告，不提供修改这个键的操作 |
| `filtered` 需要 `enabled = true` | 这个组合会在启动时和收到 `SIGHUP` 时被拒绝，并点明是哪一项，而不是接受后忽略 |
| 没有检测到网桥网络，就什么都不渲染 | `forward` 链保持原样，只在日志里留一条警告。一条旁边没有任何例外的拒绝规则，会把主机的容器流量整个关掉 |
| 端口页面看不到这个状态 | 当 `published_ports` 为 `open` 时，它把转发规则标记为不生效，因为这个键是它读得到的。有没有检测到网桥只有 core 知道，因此在 `filtered` 下没有网桥时，转发规则在那个页面上显示为已生效，而实际上什么都没渲染。日志里的警告才是信号；每次应用后请检查它 |
| 容器之间和出站不受影响 | 拒绝规则只匹配目标地址是容器地址、且来源地址不是的流量 |
| `routing.networks` 里的对端会失去对已发布端口的访问 | 拒绝规则先于那些 CIDR 例外被评估，因此你在那里放行的对端仍然需要一条转发端口规则。如果设置了 `routing.networks`，请把这种情况当作一次中断来预期 |
| IPv6 和 IPv4 一样被过滤 | 一旦网桥携带 ULA 或全局 IPv6 网络，拒绝规则和转发规则就会覆盖它。在 `ipv6.mode = "block"` 下这些不会渲染；之后发往容器的流量按 `routing.mode` 处理——`closed`/`networks` 下遇到转发策略丢弃，除非 `routing.networks` 点名了它；`open` 下则原样交给 Docker 自己的 `ip6tables` |
| 生效规则不认识的网桥会给出警告 | 执行 `docker network create` 之后，仪表盘和 `easywall-core status` 会显示 *Docker 网络 … 存在，但不在当前生效的规则中*，直到下一次应用 |

> **方案 4 是「不运行容器」，而不是「安静运行的容器」。** `forward` 链保持关闭，
> 容器的出站流量也和任何流量一样经由它路由。本页过去读起来像是只做出站的容器在
> 这里没问题。并不是。

> **方案 3 是实测的，不是假设的。** Docker 的 NAT *就是* iptables，而这个设置把
> 它去掉了。针对 Docker 29.7.2，用一个只有离开网桥才能访问的对端来验证：原装
> 守护进程 → 一条 `MASQUERADE` 规则、七条 `DOCKER` 过滤规则，对端可达。使用
> `{"iptables": false}` → 两者都没有，对端不可达；数据包以 `172.17.x.x` 为源
> 地址离开，什么也回不来。easywall 无法替你提供 masquerade：
> [自定义规则]({{ '/docs/features/custom-rules/' | relative_url }})进入的是它的
> `input` 链，而这里需要在你自己拥有的表里有一条 `postrouting` 链。

## 检查是否生效

```bash
# easywall's table — rebuilt on every apply
sudo nft list table inet easywall

# Docker's — should be untouched
sudo nft list tables | grep -i docker
```

对于不回应任何东西的已发布端口——解析器、syslog 接收端——任何基于回包的探测都
分不清*被丢弃*和*没有回应*。改看数据包计数器：
[证明一个不回应任何东西的端口]({{ '/docs/features/health/' | relative_url }}#proving-a-port-that-answers-nothing)。