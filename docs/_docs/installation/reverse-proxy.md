---
layout: default
title: 反向代理之后
description: 把 easywall 放在 nginx 后面，同时仍然知道是哪个客户端在跟它说话。
---

# 反向代理之后

easywall 自己终止 TLS，所以在它前面放一个代理是可选项，而不是必需项。一旦做了这个选择，有一件事会悄悄坏掉：现在每个请求都来自代理。由此产生三个后果：

- 审计日志记录的是代理的地址
- 应用页无法判断*你*是否即将被锁在门外
- 登录限流的「每 10 分钟 5 次尝试」被代理背后的所有人共用

把代理写进 `trusted_proxies` 能修好第一和第三个。应用页仍然会显示*无法判断*：此时它知道你的地址，但你的浏览器连的是代理的端口，不是 easywall 的，而没有任何请求会说明是哪个端口或哪台主机。按应用步骤里说的做，从第二条连接验证。列表填错，等于把地址伪造的能力交给任何能连上这个端口的东西。

## 1 · 让 nginx 指向 easywall

```nginx
server {
    listen 443 ssl;
    server_name firewall.example.com;

    ssl_certificate     /etc/letsencrypt/live/firewall.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/firewall.example.com/privkey.pem;

    location / {
        # easywall serves HTTPS with its own certificate, generated on first
        # start. It is not in your chain and nginx has no reason to verify it.
        proxy_pass https://127.0.0.1:12227;
        proxy_ssl_verify off;

        proxy_set_header Host              $host;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 2 · 告诉 easywall 该信哪个对端

```toml
trusted_proxies = ["127.0.0.1"]
```

或者 `EASYWALL_WEB_TRUSTED_PROXIES=127.0.0.1`。

这个值填的是**代理发起连接时的地址**，也就是 easywall 看到的那个地址——不是客户端的地址，也不是代理所在的网段。

## 3 · 确认它生效了

登录后打开[审计日志]({{ '/docs/features/audit-log/' | relative_url }})。那次登录的记录里应该是你自己的地址，而不是 `127.0.0.1`。如果写的是 `127.0.0.1`，说明这个头没有被采信：对端不在列表里，或者 nginx 没有发送 `X-Forwarded-For`。

## 在 Docker 里

`trusted_proxies` 里填的是**代理容器自己的地址**，而默认 bridge 并不是 compose 项目所用的那个。直接问 Docker，不要猜：

```bash
docker network inspect <network> --format '{% raw %}{{range .Containers}}{{.Name}} {{.IPv4Address}}{{"\n"}}{{end}}{% endraw %}'
```

默认 bridge 是 `172.17.0.0/16`；compose 项目会得到自己的网络，通常从 `172.18.0.0/16` 起。从上面的输出里取出代理的地址——单个地址，不带掩码。

> **容器的地址在重建之后并不固定。** 在用户自定义网络上给代理一个固定地址，否则下次整套服务起来时列表就对不上了，easywall 会悄悄回到谁都不信的状态。

## 代价是什么

被列进这个列表，意味着对该对端的完全信任。两个错误就会把地址伪造能力交给任何能连上这个端口的人：

- 列了一个实际上**不是** easywall 前置代理的地址；
- 列了**网段**而不是代理本身——那样 `10.0.0.0/8` 里每台主机都能自行选择 easywall 记录的地址、据以判定是否锁定，并影响限流。

列代理本身。不是它们所在的子网，不是 `0.0.0.0/0`，也绝不要是你控制不了的地址。这就是这个设置是*列表*而不是布尔值的原因。「信任这个头」却说不出该信谁的头，对应的就是 GHSA-3fxj-6jh8-hvhx、GHSA-rjr7-jggh-pgcp 和 GHSA-9g5q-2w5x-hmxf。easywall 的任何配置都无法表达这种设定。

`X-Real-IP`、`True-Client-IP` 和 `Forwarded` 无论来自哪个对端，都永不采信。来自不可信对端时，它们仍会翻转审计日志里的 `via-proxy` 标记——它们的存在本身就说明有东西转发了这个请求，即便其内容从不被读取。来自列表中的代理时，它们什么也不标记：在那里只有 `X-Forwarded-For` 说了算，这正是第 1 步恰好设置那个头的原因。

## 通行密钥在这里用不了

上面的示例里 nginx 监听 443，转发到 easywall 自己的 `https://127.0.0.1:12227`。浏览器看到的是 443 端口，easywall 看到的是 `:12227`——它唯一监听的端口。WebAuthn 把通行密钥绑定到浏览器实际使用的确切 origin，两者不一致，于是每次认证流程都失败。

`X-Forwarded-Host` 本可以告诉 easywall 浏览器看到的是什么，但它被拒绝了。`X-Forwarded-For` 只在来自 `trusted_proxies` 中的代理时才被采信；而对于 `X-Forwarded-Host`，没有任何机制能同样地限定*该信哪个*代理，所以这个接口对谁都不读它。

有两条出路，都超出本页范围。让浏览器直接连 443 上的 easywall，前面不放代理。`tls.acme` 正是为此存在的，它消除了「单为证书而跑一个代理」的理由。它仅在 Debian 软件包中提供，容器镜像里没有——见 [Docker → 使用自己的证书]({{ '/docs/installation/docker/' | relative_url }}#your-own-certificate)。或者接受通行密钥是这套配置不提供的又一项功能，就像 `trusted_proxies` 下审计日志和登录限流已经如此一样。无论哪种方式，密码加 TOTP 这个因子都不受影响。

## 出问题时

| 症状 | 原因 |
|---|---|
| 审计日志记录的是代理 | 对端不在 `trusted_proxies` 里，或者 nginx 没有发送 `X-Forwarded-For` |
| easywall 拒绝启动 | 某个条目既不是地址也不是 CIDR 网段。报错信息里会点名它 |
| 应用页的锁定判定显示*无法判断* | 在代理后面属于预期：你浏览器连到的是代理的端口 |
| 本来能用，后来不行了 | 某个容器被重建并拿到了新地址。改地址，不要改列表 |
| 通行密钥卡片可用，但每次认证流程都失败 | 上面的 origin 不匹配——见*通行密钥在这里用不了* |

**下一步：** [配置]({{ '/docs/configuration/' | relative_url }}) ·
[安全]({{ '/docs/security/' | relative_url }})
