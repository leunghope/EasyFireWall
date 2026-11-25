---
layout: default
title: 系统与网络设置
description: 让一条坏规则可恢复的确认窗口，以及决定 IPv6、路由流量和 Docker 能做什么的网络页面。
---

# 系统与网络设置

界面里的两个页面，它们都决定着规则页面无法决定的事。**系统**承载确认窗口；**网络**承载在任何规则被查询之前就生效的三种处理方式。

## 确认窗口

一个概念：easywall 在撤销你刚刚应用的东西之前，等你确认多久。

{% include themed-figure.html base="/assets/diagrams/apply-flow" ext="svg"
   alt="状态机：编辑后进入「暂存」；应用后进入「已生效」；在窗口内确认后进入「已确认」；窗口过期则进入「已回滚」，而暂存的编辑仍可从这取用。" %}

## 启用

| | |
|---|---|
| **开启** —— 安装的 `easywall.toml` 就是这样设置的 | 每次应用都会启动计时器。不确认就会恢复先前的规则 |
| **关闭**，键不存在也算关闭 | 应用是最终的。没有任何自动回去的方式 |

> **不要在远程主机上关闭它。** 一条关掉你自己 SSH 端口的规则，会让控制台访问成为唯一的补救。这个设置是为你能物理接触到的机器准备的。

> **只写了 `duration` 的文件根本没有窗口。** `enabled` 缺失会读作关闭，所以 duration 设定的是一个永远不会打开的窗口的时长。从 2.20.1 起，守护进程每次启动都会说明这一点 —— *应用立即生效，没有任何东西会撤销它们*。如果不是这个意思，请加上 `enabled = true` 并重新加载。

## 时长

10 到 3600 秒。该字段会拒绝此范围之外的值 —— 太短会把自己锁死，太长则是一段敞开的暴露窗口。

| | 适用 |
|---|---|
| 10–30 秒 | 几乎没有任何场景。在慢速链路上，你来不及确认 |
| **60–300 秒** | **大多数情况。从 120 开始** |
| 300–3600 秒 | 自动化流水线，或者已知链路缓慢的维护窗口 |

长到足以打开第二个连接并测试你改了什么；短到一次锁死会在它毁掉你一下午之前自行解决。

## 安装计数

默认关闭，除非你在这里或首次运行向导里打开它 —— 完整的请求见[安装计数]({{ '/docs/configuration/' | relative_url }}#counting-installations)。

[{{ site.demo_url | remove: 'https://' }}]({{ site.demo_url }}) 处的公开演示也会上报 —— 那里是通过 `EASYWALL_WEB_TELEMETRY` 打开的。所以这个数字的含义是*安装数，包含演示站*，而不是*有人在运行的安装数*。

## 网络页面 {#the-network-page}

三种处理方式，它们都在其他任何页面的任何规则被查询之前就已确定。

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/settings" ext="png"
     alt="网络设置页面：一张 IPv6 卡片，提供过滤、不处理或丢弃；一张路由流量卡片，提供不路由任何流量、路由这些网络（配一个 CIDR 列表），或不处理路由流量；以及一张 Docker 卡片，含网桥检测、允许检测到的网桥，以及一份附加网络列表。" %}
  <figcaption>每张卡片都决定规则页面无法表达的事。</figcaption>
</figure>

| 卡片 | 决定 | 细节 |
|---|---|---|
| **IPv6** | IPv6 是走同一套规则、绕过它们，还是无处可去 | [`[ipv6]`]({{ '/docs/configuration/' | relative_url }}#ipv6) |
| **路由流量** | 什么可以穿过 `forward` 链 —— 什么都不行、一份指定列表，还是全部 | [`[routing]`]({{ '/docs/configuration/' | relative_url }}#routing) |
| **Docker** | 允许哪些容器网络，入站*和*转发都算 | [Docker 共存]({{ '/docs/features/docker/' | relative_url }}) |

> **这两份列表不是同一份列表。** 无论路由模式怎么说，Docker 的网络都会穿过 `forward` 链，因为打开共存本身就已经是那个声明。`routing.networks` 是给所有会路由且不是 Docker 的东西用的 —— 一个 VPN 网关、第二个接口。

两者都接受**每行一个 CIDR 网络** —— `192.168.1.0/24`，而不是 `192.168.1.5` —— 并且和地址列表一样，把空行或以 `#` 开头的行当作注释。不是网络的条目会在保存时被指出，而不是被存下来。读取文件时会运行同一套检查，所以手工编辑的网络不会以空值的身份到达内核。

## 保存

设置会立即保存到守护进程的配置里，并会在下一次[应用]({{ '/docs/features/apply/' | relative_url }})时到达内核。从 2.10 起，这算作一项待处理的更改：仪表盘上的*未应用的更改*标签会包含它，应用页面会按它的键列出它，`ipv6.mode filter → block`。保存本身由一条 toast 确认。不需要重启，而且对禁用了 JavaScript 的浏览器，保存按钮仍然保留。两个页面的行为相同，[选项]({{ '/docs/features/filters/' | relative_url }})也是如此。

直接编辑文件也可以。`SIGHUP` 会重新加载 `[firewall]`、`[acceptance]`、`[ipv6]`、`[docker]`、`[routing]` 和 `[usage]`，而不会断开套接字。路径在启动时绑定，改动其中之一会被记录并忽略，直到重启。无法解析或无法通过校验的文件会被拒绝，运行中的配置保持不变 —— 一个拼写错误不应该解除任何防护。

重启同样不会应用任何东西。启动、`easywall-core resume` 和 Docker 协调器执行的都是上次应用过的配置，所以编辑过的 `[firewall]`、`[ipv6]`、`[docker]` 或 `[routing]` 会一直处于待处理状态，直到一次带确认窗口的应用。

```toml
[acceptance]
enabled  = true
duration = 120   # seconds
```

```bash
sudo systemctl reload easywall-core   # or: kill -HUP $(pidof easywall-core)
```

> **这在 2.5.0 之前不起作用。** 没有任何代码处理 `SIGHUP`，而未处理的默认处置是终止 —— 所以照着这页做，会把 core 关掉而不是重新加载它。

## 行为不符合预期时

| 现象 | 原因 |
|---|---|
| 应用用的是旧时长 | 正在运行的计时器保留它启动时的值。下一次应用才会读取当前值 |
| 窗口在你确认之前就过期了 | 提高时长。回滚在[审计日志]({{ '/docs/features/audit-log/' | relative_url }})里记为 `apply_rolledback`，detail 为 `timeout` |
| 字段拒绝了某个值 | 超出 10–3600 |