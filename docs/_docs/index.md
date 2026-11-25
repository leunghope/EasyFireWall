---
layout: default
title: easywall —— 工作原理与各部分的位置
description: easywall 文档 —— 安装、配置并理解这套不会把你自己锁在门外的 nftables 防火墙。
permalink: /docs/
---

## 设计思路

通过网络编辑防火墙，可能把你锁在正在编辑的那台机器之外。easywall 让这种操作默认可恢复。

{% include themed-figure.html base="/assets/diagrams/apply-flow" ext="svg"
   alt="状态机：编辑通向暂存，应用通向生效，在窗口内确认通向已确认，让窗口过期则通向已回滚；从已回滚状态仍可取用暂存的改动。" %}

编辑不改变任何东西。应用会改变一切 —— 为期 120 秒。*不*确认 —— 无论是有意为之，还是因为再也打不开页面 —— 正是让旧规则回来的原因。

## 由什么组成

| | |
|---|---|
| **两个进程** | Web 界面以无权限身份运行，没有任何通往内核的路径 |
| **用 netlink，而不是 shell** | 规则是 Go 结构体，因此不存在可供注入的命令行 |
| **三套规则集** | 暂存、当前、备份 —— 编辑与执行是分开的 |
| **审计日志** | 每次变更都记录了什么发生了变化、发生在何时，每行一个 JSON 对象 |
| **与 Docker 共存** | easywall 只拥有 `table inet easywall`，不触碰其他任何东西 |
| **English, Deutsch, Français, 简体中文** | 可在界面内切换，登录之前也可以 —— 未翻译的部分会渲染为 English 原文 |

[工作原理 →]({% link _docs/architecture.md %})

## 规则的评估顺序

在写第一条规则之前，唯一值得知道的事。允许列表中的地址可以到达所有端口；被拦截的地址在允许列表被查询之前就已丢弃。

{% include themed-figure.html base="/assets/diagrams/rule-order" ext="svg"
   alt="入站数据包的处理流程：先做分片丢弃（该功能开启时）；然后是回环、IPv6 模式、ping 与 reset 速率限制、已建立的连接与邻居发现、其他防护模块、Docker 网桥网络、拦截列表、允许列表、已开启的列表源、被应答的 ping、开放端口、自定义规则、链策略丢弃。" %}

## 从哪里开始

| 你想要 | 前往 |
|---|---|
| 在 Debian 或 Ubuntu 上安装 —— amd64 或 arm64 | [.deb 安装包]({% link _docs/installation/debian.md %}) |
| 在容器中运行 | [Docker]({% link _docs/installation/docker.md %}) |
| 从源码构建 | [手动安装]({% link _docs/installation/manual.md %}) |
| 不安装任何东西先试一下 | [在线演示]({{ site.demo_url }}) · [演示模式]({% link _docs/installation/demo.md %}) |
| 看懂设置页面在问什么 | [首次运行]({% link _docs/installation/first-run.md %}) |
| 安全地把规则送进内核 | [应用规则]({% link _docs/features/apply.md %}) |
| 理解设计 | [架构]({% link _docs/architecture.md %}) · [安全]({% link _docs/security.md %}) |
| 了解计划中的内容 | [路线图]({% link _docs/roadmap.md %}) |
| 参与贡献或提问 | [贡献指南]({% link _docs/contributing.md %}) · [Discord]({{ site.discord }}) |