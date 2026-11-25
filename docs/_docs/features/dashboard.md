---
layout: default
title: 仪表盘
description: 这道防火墙此刻正在执行什么——向内核询问，而不是假设。
---

# 仪表盘

登录后的落地页。它优先回答一个问题：
**防火墙正在做的，是不是你以为它在做的事？**

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/dashboard" ext="png"
     alt="easywall 仪表盘：防火墙状态显示为「生效中」，并附一句解释其含义的说明；然后是确认状态、待应用的更改、上次应用与上次自检；还有统计 TCP 端口、UDP 端口、拦截列表、允许列表、自定义规则和端口转发的多个卡片；最后是近期活动列表。" %}
  <figcaption>计数来自内核当前加载的规则集，而不是暂存的内容。</figcaption>
</figure>

## 状态卡片

| 显示 | 含义 |
|---|---|
| **生效中** | 规则已生效，有状态的那一半正在匹配数据包 |
| **降级** | 仍在过滤，但有某项指标可以测出不对劲 |
| **未生效** | 内核没有承载 easywall 的规则 |
| **待处理的更改** | 有内容已暂存，但运行中的防火墙没有——请前往 [应用]({{ '/docs/features/apply/' | relative_url }})。仅一条注释或调整过的列表顺序不改变任何规则，也不计入；它会被保存，并随下一次应用生效 |
| **确认状态** | 确认[窗口]({{ '/docs/features/system-settings/' | relative_url }})是否开启，以及此刻是否有一个正开着 |
| **上次应用** | 运行中的规则集最近一次被推送的时间 |
| **自检** | 关于规则构建器的四项声明，最近一次在这个内核上证明了什么 |

状态下方的那一行是**原因**，也是更有用的那一半：没有原因的状态无法据以行动。
每一种状态和原因都在
[健康检查]({{ '/docs/features/health/' | relative_url }})页面上，连同监控系统读取的
退出码。

「生效中」是**测出来的**，不是从守护进程还活着推断出来的。在 2.5.0 之前，它报告的
是守护进程自己的看法，因此一张被别的什么东西清空的表仍然显示为生效。自 2.17 起它还
会读取内核的计数器，因为曾经有五个版本，有状态的那一半一个数据包都没匹配到，而所有
界面都显示生效中。

## 计数卡片

六个计数——TCP 端口、UDP 端口、拦截列表、允许列表、自定义规则、端口转发——各自
链接到对应页面。它们描述的是**已加载的**规则集。如果你已经暂存了一处修改，卡片
显示的仍是正在执行的规则；这个数字的意义正在于此。

## 最近活动

来自[审计日志]({{ '/docs/features/audit-log/' | relative_url }})的最近几条记录，
最新的在前，并附有指向完整页面的链接。全新安装时为空。

## 导出与导入

右上角。导出把**暂存的**规则集下载为 JSON；导入会替换它。
两者都不触及运行中的防火墙——见
[导出与导入]({{ '/docs/features/export-import/' | relative_url }})。

## 更新提示

存在更新的版本时会出现一条横幅。答案来自磁盘上的缓存，每天在后台刷新一次，因此
从不会拖慢页面。在没有外网出口的主机上，失败结果会被记住一小时，而不是每次加载
都重试。`update_check = false` 会把它完全移除——见
[配置]({{ '/docs/configuration/' | relative_url }}#the-update-check)。

## 看起来不对劲时

| 现象 | 原因 | 检查 |
|---|---|---|
| 「无法连接 core 守护进程」 | `easywall-core` 没有运行，或 web 用户访问不到该 socket | `systemctl status easywall-core`，然后 `ls -l /run/easywall/core.sock`——它必须是 `root:easywall` |
| 状态显示未生效，规则却明显在工作 | 是另一个表在过滤；easywall 只报告自己的表 | `sudo nft list tables` |
| **降级**，且不知道原因 | 原因那一行会说明是三种原因中的哪一种 | `easywall-core health` 会打印同样的原因，外加自检详情 |
| **自检：此处无法验证** | 正常。要证明任何一项都需要 `CAP_SYS_ADMIN`，而守护进程没有这个权限 | 无需处理——见 [健康检查]({{ '/docs/features/health/' | relative_url }}) |
| 某个计数和你编辑的内容不一致 | 卡片显示的是**已加载的**规则集，而你的编辑还在暂存 | 应用它 |
| 出现了不是你做的待处理更改 | 另一个浏览器里的保存，或一次导入 | [审计日志]({{ '/docs/features/audit-log/' | relative_url }})会点出是哪一个 |
| 活动列表为空 | 尚无记录，或 core 无法写入 `log_dir` | `journalctl -u easywall-core` |

---

**下一步：**[应用规则]({{ '/docs/features/apply/' | relative_url }}) ·
[审计日志]({{ '/docs/features/audit-log/' | relative_url }}) ·
[导出与导入]({{ '/docs/features/export-import/' | relative_url }})