---
layout: default
title: 列表源
description: 用别人发布的列表拦截已知攻击者 —— 在你的允许列表之后检查，每个列表源都默认关闭，直到你开启它。
---

# 用列表源拦截已知攻击者

列表源是一份由别人发布、由 easywall 抓取的攻击地址列表。列表源在你的允许列表**之后**被检查，所以只要某个地址被你放行，无论列表源怎么说，它都保持可达。列表源还会丢弃 `forward` 链所过滤的转发数据包，即发往容器或路由网络的那些。要拦截整个国家/地区，见[国家/地区]({{ '/docs/features/countries/' | relative_url }})。

{% include themed-figure.html base="/assets/diagrams/feed-order" ext="svg"
   alt="一行四个方框：你的拦截列表丢弃，接着允许列表接受，接着列表源丢弃，最后是端口规则。" %}

> **列表源作用于每一个端口，包括 Web 端口。** 来自 fail2ban 的列表会拦截它点名的任何地址，无论该地址从哪里连接。**blocklist.de (all) 收录了搜索引擎爬虫、一家在线监测服务和 Cloudflare**：2026-10-15 那天，它收录了 Googlebot、Google 的专用爬虫、Bingbot、Applebot、StatusCake 和 Cloudflare 的地址。自 6 月以来的 9 次采集中，Bingbot 和 StatusCake 9 次全在，Googlebot 出现 8 次；Cloudflare 只在那一天被观测到。无法访问你站点的爬虫会让站点从搜索中掉出去；Bing 的索引还供给 ChatGPT 搜索。这就是它带 ✗ 的原因。**SSH 列表在 11 次采样中一次都没有**，是推荐使用的那个。已经开启 blocklist.de (all) 的安装保持现状即可；如果这台主机对外提供网站，就换成 SSH 列表。若要保留完整列表，就把厂商公布的网段加入允许列表（[Googlebot](https://developers.google.com/static/crawling/ipranges/common-crawlers.json)、[Bingbot](https://www.bing.com/toolbox/bingbot.json)）。列表源在你的允许列表之后被检查，因此已加入允许列表的网段会通过每一个列表源。把美国作为一个[国家/地区]({{ '/docs/features/countries/' | relative_url }})拦截，会把这两个爬虫整体拦掉。

## 从 Spamhaus DROP 和 DShield 开始

1. 打开**拦截列表**，滚动到**列表源**。
2. 启用 *Spamhaus DROP* 和 *DShield top block list*。
3. 按**保存**。这个开关和其他任何规则一样只是暂存。
4. 打开**应用规则**，按**立即应用**。如果判定结果点出某个列表源命中了你自己的地址，先把该地址加入[允许列表]({{ '/docs/features/allowlist/' | relative_url }})。
5. 在窗口内按**确认规则**，否则这个开关会回滚。

之后的刷新会载入新副本，**不会**开启确认窗口。把你用来管理这台主机的地址加入允许列表，这样任何一份副本都锁不住你。

没有任何列表源会替你开启。要查看某个列表源拒绝了谁，在[选项页面]({{ '/docs/features/filters/' | relative_url }})上启用 *记录列表源*；它的命中记录会带着列表源名称出现在[被拦截的流量]({{ '/docs/features/blocked-traffic/' | relative_url }})页面。

## 再多选几个

列表源卡片里的每一行都带三种标记之一。它们沿用 firebog 的做法。

| 标记 | 含义 | 如何开启 |
|---|---|---|
| ✓ | 最不可能把合法客户端锁在外面 | 同上 |
| • | 有用，但有明确代价 —— 该行会写明 | 同上 |
| ✗ | 需要特意选择，不是默认项 | 勾选开关下方的复选框，那里写明会锁住谁 |

这里的误报意味着某个合法客户端再也连不上：一位访客、一个 webhook 发送方、一台邮件服务器、你自己的监测。出站流量永远不会受影响。

## 自有列表源

最多三个按地址添加的自有列表。任何纯文本列表都可以：每行一个地址或网段，`#` 和 `;` 表示注释开始。

1. 在**拦截列表**页面，在**自有列表源**下填写一个槽位：名称、地址，以及列表需要时的用户名和密码。
2. 按**保存**。
3. 在**列表源**中开启该列表源，保存、应用并确认。

**移除**只有在对应列表源已关闭并应用之后才会清空槽位。

| 地址可以是 | 会被拒绝 |
|---|---|
| `https://` 指向任意主机 | URL 内含用户名或密码 —— 改用下面两个字段 |
| `http://` 指向字面回环地址，如 `127.0.0.1` | `http://localhost`，以及 `http://` 指向其他任何主机 |

自有列表源每 24 小时抓取一次。它不包含在[导出]({{ '/docs/features/export-import/' | relative_url }})中；密码存放在哪里见[安全页面]({{ '/docs/security/' | relative_url }})。

### 示例：CrowdSec 的 Raw IP List

CrowdSec 用一个 URL 提供你在其 Console 中订阅的拦截列表，无需安装 CrowdSec 引擎。

1. 在 CrowdSec Console 中创建一个 **Raw IP List** 集成，并订阅你想要的拦截列表。
2. 复制端点地址和 Basic 凭据。凭据只显示**一次**。
3. 把它们作为自有列表源填入。

| | |
|---|---|
| 地址 | `https://admin.api.crowdsec.net/v1/integrations/<id>/content` |
| 用户名、密码 | 该集成的 Basic 凭据 |
| 免费套餐 | 每 24 小时拉取一次 —— 也就是 easywall 对自有列表源的间隔 —— 且最多三个列表 |
| 条款 | 数据适用 [CrowdSec EULA](https://booking.crowdsec.net/crowdsec-eula) |

CrowdSec 默认的*拉取上限*是每次 10,000 个 IP，而 easywall 只抓取一页，因此三个订阅列表可能只到达一部分。这个上限由 CrowdSec 一侧设定：见其[集成文档](https://docs.crowdsec.net/u/integrations/intro)。

已经在运行 CrowdSec 引擎？它的 `cs-blocklist-mirror` 在本地提供你自己的判定结果：使用 `http://127.0.0.1:41412/security/blocklist`。

## 列表源失败时

每一次失败期间，上一份副本都保持生效。刷新永远不会打开确认窗口；开启或关闭列表源才会。

| 上次刷新 | 含义 |
|---|---|
| 从未抓取 | 尚无副本；即使已开启，也不拦截任何流量 |
| 已更新 | 上次刷新带来了新版本 |
| 未变化 | 服务器上没有新内容 |
| 失败 —— 上一份副本仍然生效 | *错误*行会说明原因 —— 服务器返回 HTML 时是*是网页，不是列表*；下一次尝试会自行进行 |
| 失败 —— 尚无副本 | 同上，且在副本到达之前不拦截任何流量 |

| 该行上的警告 | 该怎么办 |
|---|---|
| *已加载进内核但集合为空* | 无需处理 —— 第一份副本到达后就会拦截 |
| *30 天未变化* | 检查该列表的来源页面；停止变化的列表可能已被放弃 |
| *包含允许列表条目（N）* | 无需处理 —— 那些地址仍然可达，这正是这个顺序的意义 |
| *连续失败 N 次* | 阅读它上方的*错误*行 |
| *已拒绝：从 N 缩减到 M* | core 会拒绝小于它所存副本 70% 的列表。列表若以这个规模维持一天、并连续三次抓取都是如此，就会被采用，审计日志会说明这一点。要提前采用，先关闭该列表源并应用，再开启 —— 开启会立即抓取 —— 然后应用 |
| *N 行不是地址* | 只显示前五行 —— 通常是服务器夹在地址中的一条通知，例如一句限速提示 |

core 还会整体拒绝超过 100,000 条的列表，或包含比 /8（IPv4）或 /16（IPv6）更宽网段的列表。私有与保留网段会从每个列表源中静默剔除，*条目数*统计的是剩下的部分。

`nft list table inet easywall` 显示某个列表源的集合元素数少于*条目数*，但什么都没少。内核里的集合存的是网段：相邻或重叠的条目在写入之前会合并成一条。在一台生产主机上，CINS 的 15,000 个地址成了 10,410 个网段，blocklist.de 的 23,680 个成了 14,786 个。

在没有窗口打开时出现*另一项更改正在写入*，意味着某次列表源刷新短暂占用了应用槽位。即使面对 core 接受的最大列表，这也在一秒以内。再按一次**立即应用**；`easywall-core resume` 会自行等待它。

## 参考：列表源目录

### 实测于 2026-10-15

这些数字是一份快照。若有版本改动了某个列表源的结论，会重新实测。

| 列表源 | 拦截什么 | 条目数 | 轮询 |
|---|---|---|---|
| [Spamhaus DROP](https://www.spamhaus.org/blocklists/do-not-route-or-peer/) | 曾被劫持或移交给犯罪分子的网段 | 1,710 + 91（IPv4 + IPv6） | 12 小时 |
| [DShield top block list](https://isc.sans.edu/api/) | 过去三天内扫描源最多的 20 个 /24 网段 | 20 × /24 | 1 小时 |
| [blocklist.de (SSH)](https://www.blocklist.de/en/export.html) | fail2ban 在过去 48 小时内报告的 SSH 攻击地址 | 4,078 | 1 小时 |
| [blocklist.de (IMAP, POP3)](https://www.blocklist.de/en/export.html) | fail2ban 在过去 48 小时内报告的 IMAP、POP3 和 SASL 攻击地址 | 4,538 | 1 小时 |
| [CINS Army](https://cinsscore.com/) | CINS 传感器观测到正在攻击的地址；始终是最新的 15,000 个 | 15,000 | 1 小时 |
| [Emerging Threats compromised](https://rules.emergingthreats.net/OPEN_download_instructions.html) | 已知被入侵的主机 | 686 | 1 小时 |
| [blocklist.de (all)](https://www.blocklist.de/en/export.html) | fail2ban 在过去 48 小时内报告的攻击地址 | 25,945 | 1 小时 |
| [IPsum，出现在 ≥ 3 个列表上](https://github.com/stamparm/ipsum) | 在 30 多个公开列表中至少三个列表上出现的地址 | 19,549 | 24 小时 |
| [Hagezi threat intelligence IPs](https://github.com/hagezi/dns-blocklists) | 来自威胁情报源、为 DNS 拦截而收集的地址 | 36,487 | 24 小时 |
| [Tor exit nodes](https://metrics.torproject.org/collector.html) | 每一个 Tor 出口中继 | 1,370 | 1 小时 |

每个列表源的代价，以及它的条款在哪里：

| 列表源 | 误报 | 结论 | 条款 |
|---|---|---|---|
| Spamhaus DROP | 几乎没有 | ✓ | [条款](https://www.spamhaus.org/drop/terms/) |
| DShield top block list | 几乎没有 | ✓ | [条款](https://isc.sans.edu/api/) |
| blocklist.de (SSH) | 有一些 —— 动态和重复使用的地址，部分云主机；自 2026 年 1 月起没有出现爬虫或监测服务 | ✓ | [条款](https://www.blocklist.de/en/export.html) |
| blocklist.de (IMAP, POP3) | 有一些 —— 与 SSH 列表相同；只在邮件主机上才值得 | • | [条款](https://www.blocklist.de/en/export.html) |
| CINS Army | 有一些 —— 大约每十一个地址中就有一个是云地址 | • | [条款](https://cinsscore.com/) |
| Emerging Threats compromised | 有，且明显 —— 四分之一是 AWS 和 Google Cloud 地址；不适合要从云接收 webhook 或监控数据的主机 | • | [条款](https://rules.emergingthreats.net/open/suricata-7.0.3/LICENSE) |
| blocklist.de (all) | 有 —— 动态地址、云主机，以及 Googlebot、Bingbot、Applebot、StatusCake 和 Cloudflare | ✗ | [条款](https://www.blocklist.de/en/export.html) |
| IPsum，出现在 ≥ 3 个列表上 | 有 —— 一个聚合列表，部分来源相当激进，并包含 Tor 出口节点 | ✗ | [条款](https://github.com/stamparm/ipsum) |
| Hagezi threat intelligence IPs | 有 —— 云主机和 Tor 出口节点 | ✗ | [条款](https://github.com/hagezi/dns-blocklists) |
| Tor exit nodes | 设计如此 —— 每一个 Tor 用户，无论是否正当 | ✗ | [条款](https://metrics.torproject.org/collector.html) |

### 国家/地区 {#countries}

如何添加： [拦截一个国家/地区]({{ '/docs/features/countries/' | relative_url }})。

| | |
|---|---|
| 来源 | [ipverse/country-ip-blocks](https://github.com/ipverse/country-ip-blocks)，CC0 1.0，由五个区域注册管理机构的委派文件构建 |
| 抓取自 | `raw.githubusercontent.com`，每个国家/地区一个文件，两种地址族 |
| 刷新 | 每 24 小时；列表每天约 01:25 UTC 重建 |
| 宽度限制 | IPv4 最宽到 /4，IPv6 最宽到 /16（列表本身的限制是 /8 和 /16） |
| 上限 | 最多同时开启 30 个国家/地区 |
| 默认 | 每一个都关闭 |

设这个上限是因为每次刷新都会重写已存储的副本：30 个国家/地区会让该文件从 0.5 MB 涨到约 4 MB。