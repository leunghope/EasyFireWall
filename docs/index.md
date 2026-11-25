---
layout: default
title: easywall —— 带 Web 界面的 nftables 防火墙
description: 带 Web 界面的 Linux 防火墙管理。用 Go 编写，经 netlink 操作 nftables，双进程权限隔离；应用规则后若不确认，120 秒后自动回滚。
seo: software
# 这是落地页，不是参考页：不带页内目录，原本留给目录的那一栏还给正文。
toc: false
---

<div class="docs-hero">
  <div class="docs-hero-glow"></div>

  <div class="docs-hero-pill">
    <span class="docs-hero-pill-dot"></span>
    <span>一台服务器防火墙，<em>不会把你锁在门外。</em></span>
  </div>

  <h1 class="docs-hero-title">
    你的防火墙。<br>
    你的规则。<br>
    <span class="docs-hero-mark">没有意外。</span>
  </h1>

  <p class="docs-hero-desc">
    easywall 是 Linux nftables 防火墙的 Web 界面。每次应用都会在 120 秒后自动
    回滚，除非你确认它 —— 所以，隔着网络编辑防火墙，永远不可能把你关在门外。
    可运行在家用服务器、Raspberry Pi 或租用的 VPS 上。
  </p>

  <div class="docs-hero-buttons">
    <a href="{{ '/docs/installation/debian/' | relative_url }}" class="btn btn-primary btn-lg">安装</a>
    <a href="{{ site.demo_url }}" class="btn btn-soft btn-lg">在线演示 ↗</a>
    <a href="{{ '/docs/' | relative_url }}" class="btn btn-ghost btn-lg">文档</a>
  </div>

  <div class="docs-hero-meta">
    <span><strong>Go 1.27</strong></span>
    <span class="docs-hero-meta-sep">·</span>
    <span><strong>GPL-3.0</strong></span>
    <span class="docs-hero-meta-sep">·</span>
    <span><strong>nftables</strong> 经 netlink</span>
    <span class="docs-hero-meta-sep">·</span>
    <span><strong>Argon2id</strong> 认证</span>
  </div>
</div>

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/dashboard" ext="png"
     alt="easywall 仪表盘：防火墙状态，含确认状态、待处理的更改与最近一次应用；统计 TCP 与 UDP 端口、拦截列表、允许列表、自定义规则与端口转发的卡片；最近活动列表。" %}
  <figcaption>仪表盘先回答一个问题：这道防火墙此刻正在执行什么？</figcaption>
</figure>

<section class="docs-landstrip">
  <div class="docs-cardgrid">
    <div class="docs-card">
      <h2>Debian / Ubuntu</h2>
      <p class="mono">sudo dpkg -i easywall_amd64.deb</p>
      <p>amd64 与 arm64 共用一个安装包，内置 systemd 单元与 HTTPS。</p>
      <a href="{{ '/docs/installation/debian/' | relative_url }}">安装 →</a>
    </div>
    <div class="docs-card">
      <h2>Docker</h2>
      <p class="mono">docker compose up -d</p>
      <p>独占一张自己的 nftables 表，不触碰主机上的其他任何东西。</p>
      <a href="{{ '/docs/installation/docker/' | relative_url }}">运行 →</a>
    </div>
    <div class="docs-card">
      <h2>从源码构建</h2>
      <p class="mono">make build && sudo make install</p>
      <p>Go 1.27, no runtime dependencies —— 运行时只需内核、`nft` 与 `openssl`。</p>
      <a href="{{ '/docs/installation/manual/' | relative_url }}">构建 →</a>
    </div>
  </div>
</section>

## 防火墙始终由你掌控

一切都留在你自己的机器上：TCP/UDP 端口、拦截列表与允许列表、端口转发、自定义
nftables 规则、防护模块，以及防篡改的审计日志。没有云、没有账号，也没有你未同意
的遥测。

* 端口 —— 单个端口或端口区间，可对每条规则单独启用 SSH 暴力破解防护
* 拦截列表与允许列表 —— IPv4、IPv6、CIDR，先于任何端口规则评估
* 端口转发 —— 带协议选择的 NAT 重定向
* 自定义规则 —— 原始 nftables，应用前先做语法检查
* 与 Docker 共存 —— 独占 `table inet easywall`，不触碰其他任何东西
* English, Deutsch, Français, 简体中文，浅色与深色

[它能做什么 →]({{ '/docs/' | relative_url }})

<section class="docs-cta" markdown="0">
  <div class="docs-cta-glow"></div>
  <h2>先试用，再安装</h2>
  <p>完整界面运行在内存中的模拟之上。任何操作都不会触及真实防火墙。</p>
  <div class="docs-cta-buttons">
    <a href="{{ site.demo_url }}" class="btn btn-primary btn-lg" target="_blank" rel="noopener">打开演示 ↗</a>
    <a href="{{ '/docs/installation/requirements/' | relative_url }}" class="btn btn-soft btn-lg">环境要求</a>
  </div>
  <p class="docs-cta-note">无需账号：按一下按钮就能进去。</p>
</section>

## 以开源方式构建，面向 2026

Go、GPL-3.0、经 netlink 操作 nftables。应用规则永远不会把你锁在外面。看看
[架构]({{ '/docs/architecture/' | relative_url }})与
[安全模型]({{ '/docs/security/' | relative_url }})。在
[Discord]({{ site.discord }}) 和 [GitHub](https://github.com/jp1337/easywall) 上找到我们。