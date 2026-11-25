---
layout: default
title: 健康检查
description: 把监控接到一个真正测量自身行为的防火墙上，而不是一个只会断言的防火墙。
---

# 健康检查

2.17 之前 easywall 报告的一切都只是关于*意图*的说法：表存在、守护进程在跑、规则集已应用。有五个版本，input 链中负责状态跟踪的那一半一个数据包都没匹配到，而上述每一项都说防火墙处于活动状态。

健康检查报告的是*效果*。本页讲怎么把它接起来。

## 三个入口

| | 回答方式 | 用在哪 |
|---|---|---|
| `easywall-core health` | 三行文本加一个退出码 | 主机上的 cron 或 Nagios 式检查 |
| `GET /healthz` | Web 端口上的 JSON | 跨网络的监控系统 |
| Docker 与 systemd | 无需配置 | 容器重启与单元启动顺序 |

### `easywall-core health`

```
$ easywall-core health
health:     ok
reason:     healthy
selftest:   unprovable for 2.17.0 on 6.12.48-1-lts at 2026-09-09T07:41:02Z
```

| 退出码 | 状态 |
|---|---|
| `0` | `ok` |
| `1` | `degraded`，或收到本二进制读不懂的应答 |
| `2` | `fail`，包括守护进程没在运行 |

`2` 涵盖*没有守护进程*是刻意的。无法确认防火墙在运行，和在运行不是一回事。

**在恐慌模式下它与 `easywall-core status` 故意不一致。** `status` 退出码为 `0` —— 有人选择不过滤这台机器，一个询问意图的控制台保持沉默是对的。`health` 退出码为 `2`，因为监控系统问的是另一个问题。

### `GET /healthz`

这是 Web 进程中唯一无需会话即可应答的路由，因为编排器手上没有会话。

`/healthz` 要读内核，所以它会等待正在进行的应用。自定义规则的应用最多能占住它 30 秒。这段时间内端点返回 503，而容器的健康检查会在该窗口内到达第三次失败。什么都不会重启：不健康的容器只是一个标签，不是一个动作。

```bash
curl -sk https://127.0.0.1:12227/healthz
```

```json
{
  "state": "ok",
  "reason": "healthy",
  "selftest": {
    "version": "2.17.0",
    "result": "unprovable",
    "at": "2026-09-09T07:41:02Z"
  }
}
```

**响应体刻意做得很薄。** 没有规则细节，没有计数器，也没有内核版本 —— 这条路由不鉴权，所以它只说监控系统能据以行动的最少信息。内核版本由仪表盘给出；`/healthz` 从不提供。

没有任何记录的时间戳 —— 比如容器，那里证明根本不会运行 —— 会渲染成 `"selftest": {}` 而不是空字符串。不存在就是答案；`result` 只可能是 `passed`、`failed` 或 `unprovable` 之一。

| 状态码 | 何时 |
|---|---|
| `200` | `ok` **以及** `degraded` |
| `503` | `fail`，以及不应答的 core |
| `404` | 调用方不在 `health_allow` 中 |

`degraded` 返回 `200` 是刻意的。降级的防火墙仍在过滤，而它的每一种成因在重启后都原样存在。为它把一个容器标成不健康，会重启一个正常工作的防火墙，并掐断穿过它的每一条连接。改成对 `state` 字段告警。

用 `404` 而不是 `403`：不确认端点是否存在不花任何代价，因为被允许的人反正会得到真实答案。

### Docker 与 systemd，无需配置

| | |
|---|---|
| 镜像 | `HEALTHCHECK` 每 10 秒运行 `easywall-web -healthcheck`，启动期 15 秒，重试 3 次 —— 请求 `bind_addr` 所指地址上的 `/healthz` |
| `docker-compose.yml` | 在 `healthcheck:` 下带同一份检查，因为 podman 的默认镜像格式没地方保存镜像里的那份；有一个测试保证两者一致 |
| `easywall-core.service` | `Type=notify` —— `active (running)` 意味着套接字已存在 |
| | `WatchdogSec=60`，卡死的守护进程会被重启，而不是留在那里监听 |

**全新容器在你第一次应用规则之前都读作 `unhealthy`**，因为它确实还没有在过滤。这是诚实，不是故障，而它有一个后果：easywall 上的 `depends_on: condition: service_healthy` 会一直等下去，直到第一次应用。旧的检查看不到这一点 —— 它看的是 core 的套接字。

## 放监控主机进来

`/healthz` 初始只对回环开放。它会向任何读得到的人公布防火墙的状态。这道门读的是 TCP 对端，绝不是转发头，所以代理后面的任何东西都无法声称自己是回环。

> **改 `bind_addr` 在这里不需要做任何事。** 容器的检查在绑定的地址上、并从该地址请求 `/healthz`，而端点会放行本身是它自己本地地址的对端。要为监控主机放宽列表，而不是为这个检查。`health_allow = []` 仍然会把端点对所有人关闭，包括容器的检查。
>
> **这不是回环例外。** 本机上任何连到该绑定地址的进程都会以同样方式被放行，包括同机反向代理转发远程流量的情形。除非列表是 `[]`，否则它并不是被放行对端的完整集合。

```toml
# /etc/easywall/web.toml
health_allow = ["127.0.0.1/8", "::1/128", "10.20.0.0/24"]
```

`EASYWALL_WEB_HEALTH_ALLOW` 设置同一个列表，以逗号分隔。

| 你写 | 你得到 |
|---|---|
| 这个键，带条目 | 只有那些地址和网段 |
| 这个键，空列表 | 端点完全关闭 |
| 完全不写这个键 | 上面的回环默认值 |

**列表只与 TCP 对端匹配，绝不与 `X-Forwarded-For` 匹配。** easywall 前面的代理无法借此把监控主机放进来 —— 列出代理自己的地址。见[安全]({{ '/docs/security/' | relative_url }})。

## 三种状态，以及各自携带的原因

{% include themed-figure.html base="/assets/diagrams/health-states" ext="svg"
   alt="决策流程：恐慌模式给出 fail 与原因 panic；没有生效的规则给出 fail not_enforcing；端口在计数而已建立连接计数为零给出 degraded stateful_dead；表达式检查发现问题给出 degraded build_findings；自检失败给出 degraded selftest_failed；其余情况是 ok healthy。" %}

| 状态 | 原因 | 含义 |
|---|---|---|
| `ok` | `healthy` | 规则已生效，负责状态跟踪的那一半正在匹配数据包 |
| `degraded` | `stateful_dead` | 主机自身的端口在计数，而已建立连接那条规则一个都没数到 |
| `degraded` | `build_findings` | 有一条规则进了内核，而表达式检查无法相信它 |
| `degraded` | `selftest_failed` | 自检证伪了关于规则构建器的一项判断 |
| `fail` | `not_enforcing` | 内核没有承载 easywall 的规则 |
| `fail` | `panic` | 有人运行了 `easywall-core panic`，而且它在重启后依然存在 |
| `fail` | `core_unreachable` | Web 进程问了，但没得到回答 |

`stateful_dead` 就是这个版本要抓的缺陷。它需要两半同时成立：在没有流量的主机上 `established == 0` 是诚实的，所以告警还要求端口计数器动过。两侧都只统计 input 链：发往容器的流量走 `forward` 链，两侧都不动。2.25.1 之前它会动端口那一侧，于是只提供容器服务的主机会读作 `stateful_dead`。

`core_unreachable` 来自 Web 进程，绝不来自 core。core 崩溃说明不了内核的任何事，内核可能仍在完美地过滤。

## `unprovable` 是正常的，不是故障

自检在大多数安装上报告 `unprovable`，而健康状态保持 `ok`。

要证明任何事都需要 easywall 自己的网络命名空间，而那需要 `CAP_SYS_ADMIN`。守护进程只持有 `CAP_NET_ADMIN`，别的都没有，所以它建不出来 —— 普通容器也不能。无法证明某件事，和它坏了，不是一回事。

你会在哪里遇到它：**在一台在线主机上手动运行 `easywall-core selftest`。** 这是预期行为 —— 有一项判断需要一条在主机侧被接受的入站连接作为它的对照，而已经就位的表会拒绝它。正因如此，随包提供的证明从一个 `oneshot` 单元运行，排在守护进程*之前*。

在容器里你会遇到另一种情况。镜像只请求 `CAP_NET_ADMIN`，所以证明在那里跑不起来，也永远不会。自检行读作*此处不可用*，而不是*从未记录* —— 前者指的是一个能力，后者指的是有件事该有人去做。`/healthz` 渲染成 `"selftest": {}`，仪表盘略去这一项，状态保持 `ok`。

*从未记录*仍然会出现在它属实的场合：一台**本可以**运行证明却没有运行的主机。在那种主机上，`easywall-core selftest` 会把它清掉。

## 证明一个不做任何应答的端口

如果问题只是*这条规则匹配过吗*，端口页面的**最近使用**列已经用同一批计数器回答了。下面讲的是它回答不了的那个问题：`forward` 链的丢弃不是你写的规则，那里没有对应行。

`unprovable` 是 easywall 拒绝测量自己的断言。同一个问题也会从另一侧出现。一个从不应答的 UDP 服务，无论防火墙丢弃了数据包还是服务忽略了它，看起来都一样。任何等待应答的探测都分不开这两者。

规则会数数据包。问规则，而不是问服务。

```bash
sudo nft list table inet easywall > /tmp/before
#   from the outside host, send what needs no reply
#   dig @198.51.100.7 example.com   ·   nc -u host 514   ·   nmap -sU -p 53 host
sudo nft list table inet easywall > /tmp/after
diff /tmp/before /tmp/after
```

中间不要应用规则。一次应用会重写整张表，那会把表里每一个计数器归零。

| 动了的计数器 | 解读 |
|---|---|
| 该端口上的一条 `accept` | 数据包到达了，easywall 放行了 —— 服务拿它做了什么不是 easywall 的事 |
| `forward` 链的 `drop` | 数据包到达了，easywall 丢弃了它。已发布的容器端口却没有对应的**转发**[规则]({{ '/docs/features/ports/' | relative_url }})，是最常见的原因 |
| 什么都没动 | 什么都没到达。检查 easywall 前面的路径 —— 服务商防火墙、NAT、写错的地址 |

[Docker 共存]({{ '/docs/features/docker/' | relative_url }})里发布在网桥网关上的那个端口，就是这样查出来的：每个容器的 DNS 都没了，而十五个外部探测一直是绿的。

## 自检

四项判断，每一项都用真实数据包经一对 veth 测量，而不是从规则里读出来。

**它在你的主机上创建什么，在每次升级后的启动时。** 一对 veth，名为 `ewst-r` 和 `ewst-p`，地址为 `10.77.9.1` 和 `10.77.9.2`，位于 `10.77.9.0/24`，其中 `ewst-p` 被移进一个用完即弃的网络命名空间。证明结束时这对 veth 会被删除。运行之前有两点值得知道：

- **网段是固定的。** 如果你自己的局域网就是 `10.77.9.0/24`，证明会在它运行的这几秒里加一条与之冲突的路由。
- **名字是固定的，而且安装阶段按名字清除它们。** 任何名字字面为 `ewst-r` 的主机接口都会在构建这对 veth 之前被删除 —— 无论谁建的 —— 而删掉 veth 的一端会把对端一起带走。

四项判断：

1. 已建立连接上的应答能通过
2. 开放的端口接受连接
3. 关闭的端口不接受
4. 被拦截的地址到不了开放端口

被证伪的判断记为 `failed`，读作 `degraded` —— 从来不是拒绝过滤。这一整套里没有任何一层会拒绝把规则放进内核。

**每个版本和每个内核只运行一次。** `easywall-selftest.service` 在守护进程启动之前运行 `selftest --if-stale`，而已经同时匹配两者的时间戳不会重新证明。它抓的那类缺陷是构建缺陷：同一二进制每次启动都完全一样，所以证明一次就够。

```bash
systemctl start easywall-selftest    # honour the stamp
easywall-core selftest               # force it, and rewrite the stamp
```

不带参数的子命令总是运行。在改动规则构建器之后就用它 —— 手动运行它的理由，正是不信任已记录的结果。它会打印一行 `detail`，点出失败的判断，而 `/healthz` 刻意从不携带它。

`unprovable` 退出码为 `0`，被证伪的判断退出码为 `1`。该单元列了 `SuccessExitStatus=0 1 2`，所以两者都不会让启动失败。

两种结果都会以 `selftest_passed` 或 `selftest_failed` 进入[审计日志]({{ '/docs/features/audit-log/' | relative_url }})，而[仪表盘]({{ '/docs/features/dashboard/' | relative_url }})会带着状态和文字说明的原因。

---

**下一步：**[恢复与恐慌模式]({{ '/docs/features/recovery/' | relative_url }}) · [仪表盘]({{ '/docs/features/dashboard/' | relative_url }}) · [配置]({{ '/docs/configuration/' | relative_url }})