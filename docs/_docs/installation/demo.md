---
layout: default
title: 演示模式
description: 整个界面跑在内存 mock 上——无需 root，不用 nftables，也没有 core 守护进程。
---

# 演示模式

只跑 `easywall-web`，用内存代替特权守护进程。每个页面都能用，每次保存都会被记录，每次应用都会走一遍确认状态机。没有任何内容进入防火墙。

{% include demo-callout.html %}

| | |
|---|---|
| 需要 | 一个二进制文件 |
| **不需要** | root、`CAP_NET_ADMIN`、nftables、`easywall-core` |
| 状态 | 在内存中；重启即消失 |

## 自己跑一个

```bash
sudo mkdir -p /etc/easywall/ssl /var/lib/easywall

sudo tee /etc/easywall/web.toml > /dev/null <<'EOF'
bind_addr   = "0.0.0.0:12227"
ssl_dir     = "/etc/easywall/ssl"
data_dir    = "/var/lib/easywall"
session_key = "REPLACE_WITH_openssl_rand_hex_32"
demo_mode   = true
username    = "demo"
password    = ""      # empty: the wizard runs, and asks for the setup token below
[tls]
cert = ""
key  = ""
EOF

sudo easywall-web -config /etc/easywall/web.toml
```

启动日志会确认这一点，并且在还没有账户时，带上向导要求的初始化令牌。没有它，没人能完成首次运行——包括你自己：

```
demo mode active — using in-memory mock instead of core socket
{"time":"…","level":"WARN","msg":"first run: enter this setup token at /firstrun to create the account","token":"ABCD EFGH …"}
```

自己把向导走完。访客永远不需要账户：演示站的登录卡片只有一个**进入演示**按钮。把首次运行留给第一个访客的公开演示，正是这个令牌要避免的情况。

> **自定义规则的语法无法校验。** 这里没有 `nft` 二进制，所以页面会说明实时校验没有在运行，而不是给出一个它毫无依据的结论。它以前无论输入什么都回答「没有错误」——恰恰是在输错就会把人锁在门外的那个页面上，给出了假的绿灯。地址列表校验在 web 进程里运行，一切正常。

## 访客看到什么

- 每个页面顶栏里一个中性的**演示**标签：*不会进入真实防火墙；状态会周期性重置*
- 登录卡片上只有一个**进入演示**按钮，没有密码表单

这个按钮是一个没有密码的会话。演示模式放宽的只有这一点：按钮背后的一切都不会触及防火墙，任何会写入凭据的页面在演示模式下都会拒绝。CSRF 防护、CSP 以及其余部分保持原样。这个路由在正式安装里并不存在——没有 `demo_mode` 时，除了密码没有别的入口。由于没有特权进程在跑，最坏情况也只局限在无特权的 web 进程及其数据目录里。

## 定时重置

重启进程就会清空状态。一个定时器就够了：

```ini
# /etc/systemd/system/easywall-web-reset.timer
[Timer]
OnCalendar=hourly
Persistent=true
[Install]
WantedBy=timers.target
```

```ini
# /etc/systemd/system/easywall-web-reset.service
[Service]
Type=oneshot
ExecStart=/bin/systemctl restart easywall-web.service
```

## 出问题时

| 症状 | 原因 |
|---|---|
| 「无法连接 core 守护进程」 | 缺少 `demo_mode = true`，或者二进制文件早于这个选项。检查启动日志 |
| 状态消失了 | 设计如此——它存在内存里。想要持久化，你需要一次真实安装 |
| 试了几次后登录被拒 | 限流在这里同样生效：每 10 分钟 5 次尝试。重启进程即可复位 |
