---
layout: default
title: 架构
description: 两个进程、一条类型化 socket、三套规则集，以及一个会撤销你误操作的计时器。
---

# 架构

两个进程。暴露在网络上的那个，没有任何途径触碰防火墙。

{% include themed-figure.html base="/assets/diagrams/architecture" ext="svg"
   alt="浏览器通过 HTTPS 与 easywall-web 通信，后者以无权限身份运行；easywall-web 通过 Unix socket 以类型化 JSON 与 easywall-core 通信，后者以 root 运行，并通过 netlink 与 nftables 表 inet easywall 通信。" %}

## 谁能做什么

| | `easywall-web` | `easywall-core` |
|---|---|---|
| 运行身份 | 无权限用户 | root，位于 `easywall` 组，能力集裁剪到只剩 `CAP_NET_ADMIN` |
| 监听 | HTTPS，端口 12227 | 仅 Unix socket |
| 能否访问 nftables | **不能** | 能，经 netlink |
| 能否执行 shell | **不能** | 仅 `nft`：用 `--check --file -` 校验一条自定义规则，用 `-f -` 应用一条 |
| 持有 | 会话、模板 | 规则状态、确认计时器 |

表单解析或模板渲染中的缺陷无法触及内核，因为持有该缺陷的进程没有可被滥用的内核访问权限。要触达防火墙，还需要一条 JSON 命令：它既要被类型化协议接受，也要通过 core 的校验。

> **为什么这样构建。** 最初的 easywall —— Python、Flask、通过 `subprocess`
> 调用 `iptables` —— 把所有东西都跑在一个 root 进程里，并把用户字符串当作
> 命令行参数传递。Web 层的 bug 就是防火墙的 bug，参数还可能被注入。它在一次
> CVE 之后于 2022 年归档。这里两个根因都不存在了：权限在另一个进程里，每条
> 类型化规则都以 Go 结构体而非命令行的形式抵达内核。自定义规则是唯一仍会与
> `nft` 见面的字符串，保障其安全的机制写在
> [安全]({{ '/docs/security/' | relative_url }})中。

## 三套规则集

{% include themed-figure.html base="/assets/diagrams/rule-states" ext="svg"
   alt="编辑写入暂存。应用把当前复制到备份，并把暂存提升为当前。若确认窗口过期，备份会被恢复为当前。" %}

编辑从不触碰内核。保存写入 **暂存**；只有*应用*会把它提升为**当前**，而**备份**是在你没有确认时回来的那一套。

## 应用，以及回来的路

{% include themed-figure.html base="/assets/diagrams/apply-flow" ext="svg"
   alt="状态机：编辑通向暂存，应用通向生效，在窗口内确认通向已确认，让窗口过期则通向已回滚；从已回滚状态仍可取用暂存的改动。" %}

**应用是可逆的：什么都不做就会回滚。** 没有确认 —— 无论是有意为之，还是因为再也打不开页面 —— 都会恢复上一套规则。

| | |
|---|---|
| 窗口 | 默认 120 秒，可设 10 到 3600 |
| 计时器在哪里 | core 守护进程。关掉标签页或重启 `easywall-web` 都不算确认 |
| 窗口中途停掉 core | 视为未确认：规则会在它退出之前回滚 |
| 同样的思路 | Cisco 路由器上的 `commit confirmed`，或者有经验的操作者在高风险变更前敲下的 `at now + 5 minutes; iptables -F` —— 在这里它是默认行为，而不是习惯 |

分步说明：[应用规则]({{ '/docs/features/apply/' | relative_url }})。

## socket 协议

二十五种消息类型，两端都声明为 Go 结构体。新增一个操作，意味着两端都要添加一个常量。

有一个例外值得知道：`SaveRulesPayload.Rules` 是一个 `interface{}`，core 会把它重新编码，再按 `rule_type` 指明的类型解码。未知的 `rule_type` 会被拒绝。解码后的规则在存储之前会经过校验。这个字段本身在协议层面并不是类型化的 —— 本页曾经声称整个协议都是类型化的。

| 命令 | 用途 |
|---|---|
| `GET_RULES` · `SAVE_RULES` | 读取全部三套规则 · 写入**暂存** |
| `APPLY_RULES` · `ACCEPT` | 提升暂存并启动计时器 · 确认 |
| `CANCEL_ACCEPTANCE` | 立即结束当前窗口，不必等到超时 |
| `GET_OPTIONS` · `SAVE_OPTIONS` | 防护模块 |
| `GET_SETTINGS` · `SAVE_SETTINGS` | IPv6、Docker 与路由 |
| `GET_SYSTEM` · `SAVE_SYSTEM` | 确认窗口 |
| `GET_STATUS` · `GET_LOG` | 仪表盘 · 最近 200 条审计记录 |
| `GET_PACKET_LOG` | 防火墙拒绝了什么 —— 来自 easywall NFLOG 组的已解码数据包，经过过滤，最新的在前 |
| `EXPORT_RULES` · `IMPORT_RULES` | 以 JSON 备份与恢复 |
| `VALIDATE_CUSTOM` | 为实时编辑器执行 `nft --check` |
| `GET_APPLIED_CONFIG` | 随内核中当前那套规则一同进入内核的选项与网络设置 |
| `GET_USAGE` | 每条端口规则承载了多少流量，按规则 id 索引，以及这些数字是何时读取的 |
| `GET_HEALTH` | 防火墙是否说到做到 —— 按顺序评估的三项事实，外加最近一次自检的标识 |
| `PANIC` · `RESUME` | 拆除整张表并记录 · 结束该状态并恢复 |
| `LOG_EVENT` | 十三种登录事件之一，取自固定枚举，写入审计日志 |
| `UPDATE_FEED` | 某个列表源的新版本，由 Web 进程抓取；core 在存储或加载任何条目之前都会重新校验每一条 |
| `GET_FEEDS` | 每个列表源：条目数量、何时变更与何时校验、丢弃的数据包 —— 从不包含条目本身 |

每条消息单程最大 8 MiB。更长的请求会得到 *request too large* 的响应。

完整列表：[`internal/shared/protocol.go`](https://github.com/leunghope/EasyFireWall/blob/main/internal/shared/protocol.go)。

## 它能防什么、不能防什么

| 防御对象 | 方式 |
|---|---|
| Web 层漏洞 | Web 进程没有内核访问权限 |
| 对 `nft` 的参数注入 | 应用路径构建的是结构体，而不是命令行 |
| CSRF | 每个非安全方法都检查 `Origin` 与 `Sec-Fetch-Site` |
| 把自己锁在门外 | 确认窗口会自行回滚 |

| **不**防御的对象 | 原因 |
|---|---|
| 被攻陷的 root 账户 | root 拥有 core |
| 内核 nftables 漏洞 | 完全在 easywall 之下 |
| 管理员自己写了一条坏规则 | [审计日志]({{ '/docs/features/audit-log/' | relative_url }})会记录它，但没有什么能阻止它 |

---

**下一步：** [应用规则]({{ '/docs/features/apply/' | relative_url }}) ·
[配置]({{ '/docs/configuration/' | relative_url }}) ·
[安全]({{ '/docs/security/' | relative_url }}) ·
[规则的评估顺序]({{ '/docs/features/filters/' | relative_url }})