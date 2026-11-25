---
layout: default
title: 环境变量
description: easywall 读取的每个环境变量，以及这份清单为何只列到这里。
---

# 环境变量

本页列出 easywall 读取的每个变量、它对应的 TOML 键，以及界面是否有控制同一取值的控件。每个键配置什么、位于哪个文件，见[配置]({{ '/docs/configuration/' | relative_url }})页面。

## 取值优先级

{% include themed-figure.html base="/assets/diagrams/config-precedence" ext="svg"
   alt="存储值优先于环境变量，环境变量优先于内置默认值" %}

| 环境变量 | 存储值 | 实际生效 |
|---|---|---|
| — | — | 内置默认值 |
| 已设置 | — | 环境变量 |
| 已设置 | 不同 | 存储值 |
| 已设置 | 不同，但与环境变量相同 | 存储值——也就是同一个值 |
| 已移除 | 不同 | 存储值 |

**存储值**指文件中的取值与内置默认值不同，而不是指文件里存在这个键。`easywall-core -write-config` 与 `easywall-web -write-config` 会写出全部默认值，容器镜像提供的正是这个文件。若按「键是否存在」来算，本页的每个变量在所有容器化安装中都会失效。

布尔值按解析后的结果比较，而不是按文本比较，因此存储值 `true` 与环境变量 `1` 是同一个取值。这就是上表第四行，界面不必把它标成冲突。

控件对应的键若同时也通过环境变量设置过，页面会说明这一点。当实际生效的是存储值时，控件提供**重置为环境变量值**。

## `easywall-core`

| 变量 | `easywall.toml` 键 | 类型 | 控件 | 用途 |
|---|---|---|---|---|
| `EASYWALL_CORE_SOCKET_PATH` | `socket_path` | string | — | Unix socket 路径——必须对 `easywall` 组可访问 |
| `EASYWALL_CORE_DATA_DIR` | `data_dir` | string | — | 存放 `rules.json`、最近一次应用状态与恐慌标记的目录 |
| `EASYWALL_CORE_LOG_DIR` | `log_dir` | string | — | 存放审计日志与规则快照的目录 |

## `easywall-web`

| 变量 | `web.toml` 键 | 类型 | 控件 | 用途 |
|---|---|---|---|---|
| `EASYWALL_WEB_BIND_ADDR` | `bind_addr` | string | — | 监听地址与端口——例如 `0.0.0.0:12227` |
| `EASYWALL_WEB_SOCKET_PATH` | `socket_path` | string | — | core 的 Unix socket 路径——必须与 `easywall.toml` 一致 |
| `EASYWALL_WEB_SSL_DIR` | `ssl_dir` | string | — | 存放自动生成的 TLS 证书与私钥的目录 |
| `EASYWALL_WEB_DATA_DIR` | `data_dir` | string | — | 数据目录；Web 进程把自身状态保存在其中的 `web/` 下 |
| `EASYWALL_WEB_TLS_CERT` | `tls.cert` | string | — | 自定义 TLS 证书 PEM 文件的路径 |
| `EASYWALL_WEB_TLS_KEY` | `tls.key` | string | — | 与证书匹配的私钥 PEM 文件的路径 |
| `EASYWALL_WEB_LANGUAGE` | `language` | string | 侧边栏切换器，按浏览器生效 | 回退界面语言——`locales/` 中任意语言代码；随附 `en`、`de`、`fr` |
| `EASYWALL_WEB_UPDATE_CHECK` | `update_check` | bool | — | 每天向 github.com 查询一次是否有更新的版本 |
| `EASYWALL_WEB_DEMO_MODE` | `demo_mode` | bool | — | 改用内存中的模拟实现运行，而不是连接 core——仅用于公开演示 |
| `EASYWALL_WEB_COUNTRY_LOOKUP` | `country_lookup` | bool | — | 在**已拦截流量**中标注每条被拒绝来源所属的国家/地区——默认开启，设为 `false` 则不发任何请求 |
| `EASYWALL_WEB_TELEMETRY` | `telemetry` | bool | **系统** | 每天上报一次本安装——默认关闭，除非设置。设置后，更新完成后仪表盘不再询问 |
| `EASYWALL_WEB_TRUSTED_PROXIES` | `trusted_proxies` | list | — | 以逗号分隔的地址与网段，其 `X-Forwarded-For` 头会被采信 |
| `EASYWALL_WEB_HEALTH_ALLOW` | `health_allow` | list | — | 可读取 `/healthz` 的地址与网段，以逗号分隔。默认只有回环地址，除非放宽——并且非回环的 `bind_addr` 必须加入此处，否则容器自身的 HEALTHCHECK 会被拒绝 |

**list** 类型变量以逗号分隔：`EASYWALL_WEB_TRUSTED_PROXIES=127.0.0.1,10.1.0.0/24`。条目两侧的空白会被忽略，空字段会被丢弃，因此结尾的逗号不算一个条目。既不是地址也不是 CIDR 网段的条目会中止启动，并在错误信息中指明自身。

反向代理本身的配置见[置于反向代理之后]({{ '/docs/installation/reverse-proxy/' | relative_url }})。

侧边栏的语言切换器只为当前浏览器设置 cookie。它不写入 `web.toml`，因此既不覆盖 `EASYWALL_WEB_LANGUAGE`，也不被该变量覆盖：变量决定没有 cookie 的浏览器看到哪种语言。

## `TZ`

它不在这两张表里，因为 easywall 根本不读取它——Go 运行时通过 tzdata 读取，与其他任何 Go 程序一样。它决定界面渲染时间戳时使用的时区：审计日志、`Applied` 与 `Confirmed` 时间，以及所有带时钟的内容。不设置它，容器按 UTC 运行，其他时区的运维人员看到的每条记录都会与本地时间相差一个偏移。

## 无法在此设置的内容及原因

一句话概括这条规则：环境变量配置 easywall 在哪里运行，界面配置防火墙做什么。

- **规则设置与确认窗口**——`[firewall]` 的每个开关与限额、`[ipv6]`、`[docker]`、`[routing]`、`acceptance.duration`——都由界面写入。若用变量覆盖其中一项，下次容器重启并重新读取 `easywall.toml` 时它会被静默还原，而界面此前已把这次更改报告为已保存。`docker.published_ports` 是界面同样不写入的一个键，原因见 [Docker 共存]({{ '/docs/features/docker/' | relative_url }})。它依然不是一个变量：带着该值启动的容器，曾在任何人来得及查看之前关闭了所有已发布端口。
- **凭据、会话密钥、动态口令（TOTP）密钥、恢复码**——`username`、`password`、`session_key`、`totp_secret`、`recovery_codes`——都是机密，而环境变量不是。它能被 `docker inspect` 看到，能被任何读取 `/proc/<pid>/environ` 的东西看到，也能被有人贴进 issue 的日志看到。`web.toml` 的权限是 `0600`，运行中容器的环境变量则不是。
- **是否参与统计**——`telemetry`——*可以*在这里设置，它也是界面会写入的唯一一个键。公开演示完全由环境变量配置，并与其他安装一样上报。在界面中给出的答案属于存储值，优先于变量；系统页面上的**重置为环境变量值**是回到环境变量取值的方式。

## 行为

空变量视为未设置——`-e EASYWALL_WEB_LANGUAGE=` 会保留文件中的取值，而不会把它清空。布尔变量用 `strconv.ParseBool` 读取，因此 `1`、`t`、`T`、`true`、`TRUE` 与 `True` 都表示真，`0`、`f`、`F`、`false`、`FALSE` 与 `False` 都表示假。其他任何取值都会在启动时中止进程，并在错误信息中指明该变量。建议使用 `true` 与 `false`——其余写法可以接受，但不推荐。

## Compose 示例

以下示例基于仓库自带的 `docker-compose.yml`，并加入了上文中的若干变量：

```yaml
services:
  easywall:
    image: ghcr.io/jp1337/easywall:latest
    container_name: easywall
    restart: unless-stopped

    environment:
      - TZ=${TZ:-UTC}
      - EASYWALL_CORE_DATA_DIR=/var/lib/easywall
      - EASYWALL_WEB_BIND_ADDR=0.0.0.0:12227
      - EASYWALL_WEB_LANGUAGE=de
      - EASYWALL_WEB_UPDATE_CHECK=false

    network_mode: host
    cap_add:
      - NET_ADMIN
    security_opt:
      - no-new-privileges:true

    volumes:
      - ./easywall-config:/etc/easywall
      - easywall_data:/var/lib/easywall
      - easywall_logs:/var/log/easywall

volumes:
  easywall_data:
  easywall_logs:
```