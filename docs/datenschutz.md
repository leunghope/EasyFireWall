---
layout: default
title: 数据保护声明
description: easywall-project.org、公开演示以及自愿参与的安装计数会处理你的哪些信息 —— 以及哪些不处理。
permalink: /datenschutz/
# 其中包含控制者的姓名与通信地址（controller.html）。这些信息不进入搜索索引，
# 以免被 Google 等搜索引擎公开：noindex meta（default.html），并从 sitemap.xml
# 中排除（jekyll-sitemap）。
noindex: true
sitemap: false
---

<div lang="zh-Hans" markdown="1">

# 数据保护声明

依据 Art. 13 与 Art. 14 DSGVO 提供的信息。

## 责任主体

{% include controller.html %}

电子邮件：[info@easywall-project.org](mailto:info@easywall-project.org)

## 本声明涵盖什么

- **easywall-project.org**，本文档站点：你的 IP 地址，由 GitHub 接收
- **demo.easywall-project.org**，公开演示：一份 IP 地址经过缩短的服务器日志
- **安装计数**，自愿参与，由 easywall 安装实例发送：一个随机标识符与版本号，绝不含地址
- **电子邮件**，发往上述地址：你所写的内容

**装在你自己服务器上的 easywall 属于你。** 它处理的一切，责任都在你；只要你不
开启安装计数，就不会有任何数据到达我们这里。它的其他出站请求（更新检查、通知、
列表源、国家/地区查询、ACME）从你的服务器发往
[安全]({{ '/docs/security/' | relative_url }})页面所列的服务，而不是发给我们。

其中两个请求出厂即启用，各自每天一次。更新检查会请求 `api.github.com`。国家/地区
查询从 `github.com` 下载一份地址块列表，实际由
`release-assets.githubusercontent.com` 提供，在此过程中会收到你服务器的 IP
地址，以及一个带有 easywall 及其版本的 User-Agent；失败后，从一小时起再
重试。在 `web.toml` 中设置 `update_check = false` 与
`country_lookup = false` 即可关闭它们。

本站没有任何页面设置跟踪 cookie、加载第三方内容或使用分析服务。

## easywall-project.org

由 **GitHub Pages** 提供 —— GitHub, Inc., 88 Colin P. Kelly Jr. St., San
Francisco, CA 94107, USA。GitHub 据其自身说明，会*出于安全目的*记录并保存 Pages
站点每位访客的 IP 地址。我们不会收到这些日志记录。GitHub 已通过 EU-US Data
Privacy Framework 认证；向美国传输数据的依据即在于此（Art. 45 DSGVO）。保存
期限由 GitHub 决定，见
[GitHub Privacy Statement](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)。
法律依据：我们可靠地发布文档的合法权益（Art. 6 Abs. 1 lit. f DSGVO）。

字体与搜索都由本站自身提供。你在浅色与深色之间的选择保存在浏览器的本地存储中，
不会被发送到任何地方。

## demo.easywall-project.org

运行在 Hetzner Online GmbH 的一台服务器上（Industriestr. 25, 91710
Gunzenhausen），并与其签有数据处理协议（Art. 28 DSGVO）。

- **Web 服务器日志** —— IP 地址在写入之前就已缩短（IPv4：去掉最后两个字节；
  IPv6：只保留前 48 位），另有时间、主机、请求的地址、状态、来源页面、浏览器
  标识。用于内容投递，以及识别攻击与故障。在服务器上保存**最多 15 天**，另有一份
  副本存放在我们自行运营的日志服务器上，该服务器位于德国、是我们自己的硬件，日志
  在那里于**一年**后删除。
- **Cookie `easywall`** —— 你的演示会话。在最后一次使用 10 分钟后失效，或在你
  退出登录时失效。
- **Cookies `easywall_pending`、`easywall_passkey`、`easywall_login_passkey`** ——
  仅在你试用第二因子相关页面期间存在。有效期 1 至 5 分钟。
- **Cookie `easywall_lang`** —— 仅在你切换语言时存在。有效期一年。
- **你输入的内容** —— 规则、地址、设置。每 6 小时重置一次。

演示**不会**在其审计日志或自身日志中记录 IP 地址，你输入的任何内容都不会到达
真实防火墙。Web 服务器不会为演示写出错误日志。这些 Cookie 对于你要求演示执行
的操作是绝对必要的（§ 25 Abs. 2 Nr. 2 TDDDG）；其中没有任何一个会跟踪你。法律依据：
展示该软件并安全运行它的合法权益（Art. 6 Abs. 1 lit. f DSGVO）。

## 安装计数

关闭状态，除非 easywall 安装实例的运营者主动开启。安装实例每天一次向
`telemetry.wdkro.de` 发送一个随机的 32 字符标识符及其版本号 —— 就是同一台
Hetzner 服务器。接收端的 Web 服务器会把时间、标识符与版本号写入一份保存 36 天
的日志，**不记录 IP 地址**；留存更久的只是按版本统计的一个数字。标识符在你的
服务器上生成，不指向任何人；删除 `telemetry.json` 的人会得到一个新的。法律依据：
同意（Art. 6 Abs. 1 lit. a DSGVO），可通过在*系统*下关闭来撤回 —— 之后的下一次
上报就不会再发送。在关闭期间，界面会在每次功能更新后询问一次；*不再询问*即可
结束。

## 电子邮件

仅用于回复你。法律依据：如果涉及的是你请求的一项服务，则为
Art. 6 Abs. 1 lit. b DSGVO；否则为我们回复询问的合法权益（lit. f）。保存期限以
该询问所需为限，只有法律要求时才会更久。

## 经 Impressum-Privatschutz 收转的通信地址

寄往上述地址的邮件由 IMPRESSUM-PRIVATSCHUTZ GmbH（Ludwig-Erhard-Str. 18, 20459
Hamburg）代收并转交；我们与其签有数据处理协议。法律依据：拥有可靠通信地址的
合法权益（Art. 6 Abs. 1 lit. f）。
[其数据保护声明](https://impressum-privatschutz.de/datenschutzerklaerung/)。

## 接收方

上述服务提供商，以及法律要求我们披露时的公共机构（Art. 6 Abs. 1 lit. c）。
任何数据都不会被出售，也不会用于广告共享。向欧盟以外传输数据的只有发往 GitHub
的这一次，如上所述。

## 你的权利

查阅（Art. 15 DSGVO）、更正（16）、删除（17）、限制处理（18）、数据可携带（20），
以及**反对**基于合法权益的处理（21）。同意可以随时撤回，仅对未来生效
（Art. 7 Abs. 3）。可写信至上述地址或电子邮件。你也可以向数据保护监管机构投诉
（Art. 77）。

截至 2026 年 10 月 15 日。来源：Impressum-Privatschutz，已针对本网站调整。

同一份声明的另一版本：[隐私政策]({{ '/privacy/' | relative_url }})。

</div>