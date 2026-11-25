---
layout: default
title: 自定义规则
description: 手写的原始 nftables 语句，用来表达其他页面无法表达的内容——以及那一个会被拒绝的字符。
---

# 自定义规则

手写的 nftables 语句，追加到 easywall 的 input 链。凡是表单页面无法表达的内容，
都从这里出手：

- 只对单个来源开放的端口
- 对某个特定服务做的速率限制
- 你想要的一条日志记录

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/custom" ext="png"
     alt="自定义规则页面：一个等宽文本框，每行一条 nftables 语句，语句之间可以夹注释；下方是规则计数；旁边是提示卡片，警告写错规则可能把你锁在门外，并说明只需写语句主体，不必写完整的 add rule 命令。" %}
  <figcaption>语法在你输入时就被检查，检查用的正是稍后执行它的同一个 <code>nft</code>。</figcaption>
</figure>

## 文件的形状

每行一条语句。`#` 注释和分组之间的空行都按你写的原样保留——它们是文件的一部分，
保存时不会被去掉。

```
# allow Prometheus node-exporter only from the monitoring host
ip saddr 192.0.2.50 tcp dport 9100 accept

# rate-limit inbound DNS queries to mitigate amplification
udp dport 53 limit rate 50/second accept

# log and drop traffic to a legacy admin port
tcp dport 10000 log prefix "legacy-admin: " drop
```

每一行都会以 `add rule inet easywall input <your line>` 的形式写入 input 链，
因此只写语句本身——不写 `add rule`，不写表，不写链。

## 不允许换行，不允许分号

`nft` 会把两者都读作一条命令的结束和下一条命令的开始。因此，含有其中任何一种的语句
就不是一条规则，而是一条第二条命令：由 root 守护进程执行，能够触及 easywall 并不
拥有的表。两者都会被拒绝——保存时、导入时，以及在 core 内部再拒绝一次。

> **这是实测过的，不是理论推演。** 一条含有换行的导入规则写进了相邻的表，而这正是
> 「easywall 只拥有一个表，从不查看另一个表」声称不可能发生的事。检查针对的是输入的
> *形状*，而不是 nft 的语法。因此它既不依赖子进程是否可用，也不依赖某个包装器是否
> 恰好括号不配对。推理见 [安全]({{ '/docs/security/' | relative_url }})。

每次校验还有 **256 条语句**的上限——远高于任何真实的规则集，也远低于表单请求体
能够携带的量。

## 在哪里被评估

在最后——端口之后，链策略丢弃剩余流量之前。

{% include themed-figure.html base="/assets/diagrams/rule-order" ext="svg"
   alt="入站数据包的判定流程：先按需丢弃分片，然后是回环地址，接着是 IPv6 模式，然后是 ping 与 RST 的速率限制，再是已建立的连接与邻居发现，然后是其余防护模块，接着是 Docker 网桥网络，然后是拦截列表（命中即丢弃），接着是允许列表（接受所有端口），再是你启用的列表源（命中即丢弃），然后是 ping（仅当「回应 ping」开启时），接着是开放端口，再是自定义规则，最后是链的默认策略（丢弃）。" %}

有两个后果值得了解：

- 更早的规则已经接受的数据包，永远不会到达你的语句。要*限制*端口规则放开的内容，
  请删掉那条端口规则——排在它后面的自定义 `drop` 不会被查询。
- 你的语句接受的数据包，正是策略原本会丢弃的那一个；这也通常是写它的理由。

## 输入时即校验

编辑器把文本发送给 core，core 在一个一次性表里让它跑过 `nft --check`，并报告任何
被拒绝内容所在的行号。一组合法的语句只花一次 `nft` 运行；只有被拒绝的那一组才会
逐行复查，以指出是哪一行。

| 状态 | 含义 |
|---|---|
| 无错误 | `nft` 能解析每一条语句。这不代表规则会做你想要的事 |
| 一个行号 | 该语句不是合法的 nftables 语法；消息来自 nft 本身 |
| 「实时校验未运行」 | core 不可达，因此什么都没校验——不代表「一切正常」 |

在[演示模式]({{ '/docs/installation/demo/' | relative_url }})下，界面背后没有 `nft`，
因此它把检查器报告为不可用，而不是声称你的规则合法。

## 在你应用之前，什么都不会发生

保存只是暂存，和其他每个页面一样。这里值得知道的是：通过语法检查的语句，仍然可能
在[应用]({{ '/docs/features/apply/' | relative_url }})时被**内核**拒绝。两个例子：
链不存在，集合未定义。应用失败会回滚到先前的规则。

## 不工作时

| 现象 | 原因 |
|---|---|
| 提示规则里含有换行，nft 会把它读作命令的结束 | 把每条语句单独放一行；这不是格式偏好问题 |
| 语法合法，却没有效果 | 更早的规则先接受了该数据包——见*在哪里被评估* |
| 应用失败并回滚了 | 内核拒绝了该语句。`journalctl -u easywall-core` 里有 nft 的消息 |
| 注释消失了 | 现在不会了。保存会保留 `#` 行和空行 |
| 语法检查什么都不说 | core 不可达，或者你处在演示模式 |