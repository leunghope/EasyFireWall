---
layout: default
title: 首次运行
description: 一个页面——账户，以及这台主机如何开始。除账户之外，一切都只是暂存。
---

# 首次运行

第一次打开 `https://<server>:12227` 时，easywall 只提供初始化页面，别的什么都没有。它先要一个初始化令牌，然后问两件事。

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/firstrun" ext="png"
     alt="首次运行页面：左侧是账户区，包含初始化令牌输入框、用户名、密码和确认密码；右侧是初始选择区，包含 SSH 端口、一条说明 Web 端口保持开放的注释、80 和 443 的开关、三个 IPv6 选项，以及统计安装的开关。" %}
  <figcaption>右侧的选择只是暂存。这里没有任何内容会进入防火墙。</figcaption>
</figure>

## 初始化令牌 {#the-setup-token}

这个页面对任何能访问 12227 端口的人开放，所以它先要一份证明：你能读到这台服务器的日志。easywall-web 每次在没有账户的情况下启动时，都会往那里打印一个令牌：

| 安装方式 | 取出令牌 |
|---|---|
| Docker | `docker compose logs easywall \| grep 'setup token' \| tail -1` |
| Debian | `sudo journalctl -u easywall-web -g 'setup token' \| tail -1` |

把取到的值粘贴到 `"token":` 后面——带不带其中的空格都行。

| | |
|---|---|
| 有效期 | 直到账户存在，或 easywall-web 重启。重启会打印新的，旧的随即失效 |
| 丢了 | 重启 easywall-web，读新打印的那一行 |
| 存在哪里 | 只在运行中的进程里，以及那一行日志 |

## 你的账户

| | |
|---|---|
| 几个账户 | 一个。easywall 目前没有用户管理——见[路线图]({{ '/docs/roadmap/' | relative_url }}) |
| 密码 | 至少 12 个字符，包含一个数字和一个符号，用 Argon2id 加每个密码独立的盐值哈希 |
| 找回 | **设计上就没有。** 没有邮件，没有外部服务——见[下文](#if-you-lose-the-password) |

**第二因子是强制的。** 按下「完成」并不会创建账户——它会改为显示一个设置步骤：二维码、可手输的密钥，以及服务器自己的时钟，与[两步验证]({{ '/docs/features/two-factor/' | relative_url }})页面完全一致。只有确认过的六位口令才会写入账户，并同时生成八个只显示一次的恢复码。

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/firstrun-2fa" ext="png"
     alt="首次运行向导的设置步骤：左侧白底上是二维码；右侧是可手输的密钥、服务器自己的时钟，以及一个六位确认口令的输入框和确认按钮。" %}
  <figcaption>确认口令之前，什么都不会保存。没有口令就无法越过这一步。</figcaption>
</figure>

要在这个页面上完成确认，你必须已经在手边装好了身份验证器应用。而首次运行恰恰是运维人员最不可能准备好的时刻。它发生在安装过程中，所在机器甚至可能没有多余的浏览器标签页去扫二维码。向导不提供跳过这一因子的办法——没有第二因子的密码，正是这个版本要让其无法达成的状态。不过，如果口令怎么都验证不过，仍有别的路可走：见[如果验证码始终不匹配](#if-the-code-never-matches)。

## 首次选择——全部暂存

| 答案 | 它做什么 |
|---|---|
| **SSH 端口** | 作为开放 TCP 端口暂存，并勾选暴力破解防护。默认 22 |
| **12227 端口** | 自动加上，因为防火墙会丢弃所有未被告知允许的内容——包括这个页面 |
| **同时开放 80 和 443** | 再加两个暂存端口，适合托管网站的服务器 |
| **IPv6** | 过滤它（几乎总是对的）、保持不动，或丢弃它。以**设置**的形式保存，而非暂存——它决定之后每条规则如何评估 |
| **统计这次安装** | 除非你打开，否则关闭。它发送的内容完整列在[配置]({{ '/docs/configuration/' | relative_url }}#counting-installations)下 |

**暂存意味着尚未生效。** 登录后在[端口]({{ '/docs/features/ports/' | relative_url }})页面审阅，再用[应用]({{ '/docs/features/apply/' | relative_url }})推送——后者若未确认仍会自行撤销。初始化页面是最不适合破例的时刻：还没有人验证过自己是否仍能连上这台机器。

**SSH 端口是唯一可能把你锁在门外的答案**，所以它在创建账户之前就被校验。校验发生时这个页面还在你面前，因此你可以改正它。

## 按下「完成」之后会发生什么

1. 「完成」会校验账户字段和 SSH 端口，然后显示上面描述的那个设置步骤——此时还没有写入任何东西。
2. **确认**口令会写入账户，然后暂存初始选择，并把你带到恢复码页面：只显示一次，之后不再显示。从这里登录需要主动点击，不会自动跳转。
3. 如果在暂存选择时 core 守护进程没有响应，失败的只是这一部分，恢复码页面会说明这一点：*「账户和第二因子已创建……初始选择无法暂存。」* 你可以先登录，再自己设置这些选择。

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/firstrun-codes" ext="png"
     alt="首次运行向导确认第二因子之后立刻显示的八个一次性恢复码，带一个「复制恢复码」按钮和一个「继续登录」按钮。" %}
  <figcaption>这八个恢复码唯一一次显示。现在就复制；两步验证页面可以签发新的，但无法再次显示这些。</figcaption>
</figure>

<figure class="docs-shot">
  {% include themed-figure.html base="/assets/img/screens/login" ext="png"
     alt="easywall 登录页：一张卡片上有用户名和密码输入框、「登录」按钮，以及页脚里的语言菜单。" %}
  <figcaption>语言切换也出现在登录页上——读不懂界面的人仍然能进去。</figcaption>
</figure>

如果向导拒绝了某项内容，除了两个密码，其他每个答案都会随页面一起回填。重新输入密码不会悄悄把你的 SSH 端口重置为 22。

## 如果验证码始终不匹配 {#if-the-code-never-matches}

口令始终验证不过，几乎总是说明这台机器的时钟不对，而不是输错了数字。「确认」能察觉出最多五分钟的时钟偏差，但它只接受与该时钟相差三十秒以内的口令。偏差超过三十秒，被拒绝的方式与完全错误的口令一模一样。没有实时时钟的单板机开机时可能差上好几年，直到 NTP 校正过来，而这两种情况都不会被接受。

失败一次之后，设置步骤会提供第二条通路：

| | |
|---|---|
| 何时可见 | 只有在一个口令已经失败过之后——第一次渲染时页面上没有它 |
| 确认方式 | 一个复选框，不是链接：必须勾选「我明白我的口令无法通过验证……」 |
| 它写入什么 | 账户、上方已经显示过的密钥，以及八个恢复码——与确认口令时完全相同的写入 |
| 它不做什么 | 跳过这一因子。它创建的账户仍然登记了一个因子，与本版本移除的跳过路径不同 |

用八个恢复码之一登录，然后修正时钟。已经配对的身份验证器无需重新登记即可继续使用：你选择这条路而非匹配口令时，存储的就是当时屏幕上已经显示的同一个密钥。

## 之后修改密码

界面里的**系统 → 密码**。在你执行修改的那台设备上，你保持登录状态。其他所有会话立即失效，因为每个会话都携带签发时所用密码哈希的指纹。

### 如果忘记密码 {#if-you-lose-the-password}

没有重置链接。找回意味着能访问持有 `web.toml` 的那台主机，具体做法按安装方式不同：

| | Debian / systemd | Docker |
|---|---|---|
| 编辑 | `sudo sed -i -E 's/^password[[:space:]]*=.*/password = ""/' /etc/easywall/web.toml` | `sudo sed -i -E 's/^password[[:space:]]*=.*/password = ""/' ./easywall-config/web.toml` |
| 重启 | `sudo systemctl restart easywall-web` | `docker compose restart easywall` |

清空 `password` 那一行会让这个页面重新打开——带着新的初始化令牌，因为清空账户正是让这台主机重新回到首次运行的原因。用读第一个令牌的同样方式读新的：`journalctl -u
easywall-web -g 'setup token' | tail -1`，或 `docker compose logs easywall |
grep 'setup token' | tail -1`。规则、审计日志和其他所有设置都不受影响——只有账户被重建。

## 反向代理之后

请读[反向代理之后]({{ '/docs/installation/reverse-proxy/' | relative_url }})。把代理写进 `trusted_proxies`，才能解决两个问题：共用的登录限流，以及审计日志把代理自己的地址记成客户端地址。

## 出问题时

| 症状 | 原因 | 检查 |
|---|---|---|
| 「该初始化令牌不匹配」 | 打错了，或者你复制之后 easywall-web 又重启了 | 取最新的一行——见[初始化令牌](#the-setup-token) |
| 初始化页面 404 | 账户已经存在 | 去 `/login`；清空 `password` 那一行即可重来 |
| 「这不是端口号」 | SSH 端口不在 1–65535 之间 | |
| 「初始选择无法暂存」 | core 守护进程无法连接 | `systemctl status easywall-core`（Debian）或 `docker compose logs easywall`（Docker），然后手动设置端口 |
| 浏览器提示证书有问题 | 首次启动时用的是自签名证书 | 接受它，或配置你自己的证书——[Debian]({{ '/docs/installation/debian/' | relative_url }}#your-own-certificate) · [Docker]({{ '/docs/installation/docker/' | relative_url }}#your-own-certificate) |
| 已登录，但每个页面都说连不上 core | web 用户访问不到那个 socket | `ls -l /run/easywall/core.sock`——必须是 `root:easywall` |
| 登录时出现「Too Many Requests」 | 十分钟内从你的地址失败了五次 | 等一等；每十分钟才返还一次尝试机会 |

---

**下一步：** [应用规则]({{ '/docs/features/apply/' | relative_url }}) ·
[端口]({{ '/docs/features/ports/' | relative_url }}) ·
[配置]({{ '/docs/configuration/' | relative_url }})
