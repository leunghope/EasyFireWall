---
layout: default
title: 导出与导入
description: 整个规则集存成一个 JSON 文件——用于迁移、高风险改动前的备份，或者版本控制。
---

# 导出与导入

按钮位于仪表盘右上角。导出会下载**暂存的**规则集；
导入会替换它。两者都不触及运行中的防火墙——那仍然需要一次
[应用]({{ '/docs/features/apply/' | relative_url }})。

| | |
|---|---|
| **导出** | 下载 `easywall-rules-<date>.json` |
| **导入** | 先校验，然后整体替换暂存的规则集——它不是合并 |

> **导入是替换。** 导入之后，任何已暂存但不在文件中的内容都会消失。
> 不确定就先导出。

## 格式

每种规则类型对应一个数组。每个字段都是可选的；数组缺失或为空表示没有该类规则。

```json
{
  "tcp": [
    {"port": "22",  "description": "SSH",   "ssh": true},
    {"port": "443", "description": "HTTPS", "ssh": false}
  ],
  "udp":        [{"port": "53", "description": "DNS", "ssh": false}],
  "blocklist":  ["192.0.2.1", "198.51.100.0/24"],
  "allowlist":  ["203.0.113.10"],
  "forwarding": [{"protocol": "tcp", "source_port": 2222, "dest_port": 22}],
  "custom":     ["iif eth0 ip protocol udp udp dport 1194 accept"]
}
```

| 键 | 结构 |
|---|---|
| `tcp`、`udp` | `port` 为字符串，`description` 为字符串，`ssh` 为布尔值 |
| `blocklist`、`allowlist` | 字符串——IPv4、IPv6 或 CIDR |
| `forwarding` | `protocol` 为 `"tcp"` 或 `"udp"`，`source_port` 为整数，`dest_port` 为整数 |
| `custom` | 字符串——原始 nftables 匹配表达式 |

## 哪些内容会被拒绝

导入会先校验，再暂存任何内容。**文件里只要有一个条目不合格，就什么都不会导入。**

| 拒绝的内容 | 示例 | 原因 |
|---|---|---|
| 超出 1–65535 的端口 | `70000` | |
| 格式错误的 CIDR | `192.168.1.1/33` | |
| 无法解析的地址 | `192.0.2.999` | |
| 不是 `tcp` 或 `udp` 的转发协议 | `sctp` | |
| 含有换行或分号的自定义规则 | | nft 会把两者都读作一条命令的结束，因此它会成为**由 root 守护进程执行的第二条命令**。编辑器按换行拆分，从来产生不了这种内容；文件可以 |
| `nft --check` 拒绝的自定义规则 | | 编辑器一直边输入边校验；导入的规则过去完全不校验 |
| 超过 **512 KB** 的文件 | | 足以容纳数千个地址。更大的集合直接放进 `rules.json` 比通过浏览器上传更省事 |

## 值得做的事

| | |
|---|---|
| **高风险改动之前** | 先导出。重新导入胜过手工重建二十条规则 |
| **迁移服务器** | 导出、传输、导入、应用——并在确认之前验证 |
| **版本控制** | JSON 稳定，与你的 Ansible 或 Terraform 放在一起时差异清晰 |

```bash
git add easywall-rules-production.json
git commit -m "chore(firewall): open 8080 for staging"
```