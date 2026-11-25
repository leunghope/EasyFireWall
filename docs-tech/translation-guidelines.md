# easywall 翻译规范（简体中文）

本文件是 easywall 简体中文本地化的术语与文案规范。界面文案在
`locales/zh.json`，文档在 `docs/` 与 `README.md`。所有新增或修改译文都应
先对照本文件。

## 一、总原则

1. **只改表达，不改事实。** 译者可以自由改写句子结构，但不得改变句子所
   陈述的事实及其方向（谁先于谁被评估、变更是保留还是回滚、端口是打开
   还是关闭）。同一个功能在不同语言里描述成两件不同的事，不是风格差异，
   而是描述了一个不同的防火墙。
2. **技术标识符一律不译。** 变量名、函数名、类名、API 路由、JSON Key、
   数据库表/字段名、环境变量、命令行参数、URL、文件路径、正则表达式、
   日志标识符、第三方库与产品名保持原文。
3. **术语统一。** 同一英文术语在同一业务场景下始终使用本表同一译法；
   本表未覆盖的术语，先补表再翻译。
4. **符合中文软件表达习惯。** 避免逐字直译、避免「的的」堆叠、避免英文
   语序。按钮用动词短语，提示用完整句，错误信息先说结果再说原因。

## 二、核心术语表

| English | 简体中文 | 备注 |
|---|---|---|
| easywall | easywall | 品牌名，不译 |
| nftables / netlink / kernel | nftables / netlink / 内核 | 技术名保留；kernel 可译「内核」 |
| firewall | 防火墙 | |
| Dashboard | 仪表盘 | |
| Ports | 端口 | |
| Blocklist | 拦截列表 | 不译作「黑名单」 |
| Allowlist | 允许列表 | 不译作「白名单」 |
| Port Forwarding | 端口转发 | |
| Custom Rules | 自定义规则 | |
| Options | 选项 | |
| Network | 网络 | |
| Settings | 设置 | |
| System | 系统 | |
| Notifications | 通知 | |
| Audit Log | 审计日志 | |
| Feeds | 列表源 | 指外部 IP 列表订阅，不是「反馈」 |
| Countries | 国家/地区 | |
| Telemetry | 安装计数 / 统计 | 界面原文说的是 counting（计数），不用「遥测」——用户能看懂的是「统计这个安装吗？」 |
| Passkey | 通行密钥 | |
| Two-factor / Second factor | 两步验证 / 第二因子 | |
| TOTP | 动态口令（TOTP） | 首次出现保留缩写 |
| Recovery code | 恢复码 | |
| Setup token | 初始化令牌 | |
| Health | 健康状态 | |
| Demo | 演示 | |
| Provenance | 来源 | 配置项取值来源 |
| Apply | 应用 | 动词；「应用规则」 |
| Staged | 暂存 | 规则集状态 |
| Current | 当前 | 当前生效规则集 |
| Backup | 备份 | |
| Confirm | 确认 | |
| Accepted | 已接受 | 应用窗口状态 |
| Pending | 待处理 | |
| Rolled back | 已回滚 | 固定译法，不用「已撤销」 |
| Unapplied changes | 未应用的更改 | |
| Acceptance window | 确认窗口 | 见「危险语义」 |
| Rate limit | 速率限制 | |
| Session | 会话 | |
| Token | 令牌 | |
| Request / Response | 请求 / 响应 | |
| Enable / Disable | 启用 / 禁用 | |
| Delete / Cancel / Save / Submit | 删除 / 取消 / 保存 / 提交 | |
| rule | 规则 | |
| chain / table | 链 / 表 | nftables 概念，保留层级 |
| packet | 数据包 | |
| interface | 接口 | 网卡接口 |
| inbound / outbound | 入站 / 出站 | |
| loopback | 回环 | |
| bogon filter | 无效地址过滤 | |
| Docker bridge | Docker 网桥 | |
| fragment | 分片 | |
| stateful | 有状态 | |
| protection module | 防护模块 | |
| panic mode | 恐慌模式 | 固定译法 |
| console | 控制台 | |
| source / destination | 源 / 目标 | 端口语境用「目标端口」 |
| CIDR | CIDR | 保留 |

## 三、危险语义：必须逐句核对

以下六类文案一旦译反，会直接导致操作者对自己的防火墙做出错误决策。译文
必须与英文陈述**同一事实、同一方向**。完整英文原文见
`docs-tech/i18n-review.md`。

1. **规则评估顺序**：拦截列表先于允许列表、列表源与端口规则评估；被拦截
   的地址即使命中端口规则也会被丢弃；允许列表在端口规则之前被接受，但
   **防护模块仍先运行**，只有无效地址过滤对允许列表例外。
2. **确认窗口**：应用后若未在超时内确认，**先前规则会自动恢复**。不得译
   成「已保存」「已保留」。确认才保留新规则；什么都不做会回滚。
3. **恐慌模式**：它停止的是**过滤**，不是它自己；重启后仍然生效，且只能
   在主机控制台用 `easywall-core resume` 结束，界面没有对应按钮。
4. **第二因子与恢复码**：恢复码只显示一次；磁盘写入失败可能导致同一恢复
   码可再次使用；没有邮件找回；密码与第二因子同时丢失只能改主机上的配置
   文件。
5. **拦截/允许语义**：拦截即丢弃；允许列表跳过端口规则、可达所有端口，
   但不跳过防护模块。
6. **演示模式**：演示中的任何操作都**不会**触及真实防火墙，状态会周期性
   重置。

## 四、占位符与标记

- go-i18n 模板动作 `{% raw %}{{.Name}}{% endraw %}`、`{% raw %}{{.N}}{% endraw %}`、
  `{% raw %}{{.Err}}{% endraw %}` 等**原样保留**，不得改写、不得翻译字段名、
  不得增删空格以外的内容。
- 纯文本占位符 `{name}`、`{}`、`%s`、`%d` **原样保留**，数量与顺序不变。
- Markdown 标记：反引号 `` `code` `` 与星号 `*强调*` 必须保留，且含标记的
  键必须仍由模板的 `richText` 渲染（`TestMarkupStringsAreRenderedThroughRichText`）。
- HTML 标签及其属性、class 名、Liquid 标签（`{% raw %}{% link %}{% endraw %}`、
  `{% raw %}{% include %}{% endraw %}`、`{% raw %}{{ site.* }}{% endraw %}`）原样保留。
- 富文本中可译的只有标签之间的自然语言文本。

## 五、文案风格

| 类型 | 规范 | 示例 |
|---|---|---|
| 按钮 | 动词或动宾短语，2–4 字，不加句号 | 保存、应用、取消、删除规则 |
| 导航 | 名词短语，简短 | 仪表盘、端口、审计日志 |
| 表单标签 | 名词短语，首字母不大写 | 端口、来源、协议 |
| 占位符 | 说明期望输入 | 例如 443 或 8000:9000 |
| 提示/说明 | 完整句，句末加句号 | 更改会先暂存，应用后才会生效。 |
| 错误 | 先说结果，再说原因/下一步 | 未保存：端口格式无效。 |
| 空状态 | 陈述现状 + 下一步 | 尚未打开任何端口。添加一个以接受入站连接。 |

- 中文与英文、数字之间加一个半角空格（如「Go 1.27」「端口 80」）。
- 标点用中文全角（，。：；「」），代码与标识符内的标点保持英文。
- 不用感叹号；不堆叠「请」「您」；界面统一用「你」或不出现主语。
- 同一概念不要中英混排；确需保留英文缩写时首次出现给出中文。

## 六、格式规范

- 日期：`2026-11-24`；时间：24 小时制 `14:30`；时区保留原文标识。
- 数字：千分位用逗号 `1,024`；范围用 `–` 或「至」。
- 端口：纯数字，不译；区间 `8000:9000` 保留冒号写法。
- IP/CIDR：原样保留，不加中文标点。
- 百分比：`5%`；时长：`120 秒`、`30 天`。
- 货币与账单：本项目不涉及。

## 七、语言枚举约定

界面语言切换器中的语言名使用各语言自身的写法（endonym）：English、
Deutsch、Français、简体中文。文档中列举支持语言时，保留这些 endonym 并
追加「简体中文」，例如：

> English, Deutsch, Français, 简体中文

不要遗漏任何一种已提供的语言、只写其中前两种（守卫测试
`TestNoDocumentStillClaimsOnlyTheStrictPair` 会判定为过时声明）。新增语言
时，`internal/web/i18n_docs_test.go` 的 `endonyms` 必须同步增加该语言。

## 八、审校状态

`locales/status.json` 里 `zh` 是 `reviewed: true`：简体中文已由母语者逐句通读并
修正，界面语言切换器只显示「简体中文」，不再带「未审校」。

标成 `true` 不是一枚奖章，而是一条持续生效的守卫：`TestTranslationsAreNotCopiedEnglish`
从此把 `zh.json` 纳入检查，任何与 `en.json` 逐字节相同、且英文原文不短于 12 个
字符的条目都会让它失败——新增字符串必须真的翻译，不能先粘贴英文占位。九个 id
在允许列表里（`demo_label`、`nav_group_system`、`nav_system`、`ports_tcp`、
`ports_udp`、`log_col_user`、`forwarding_protocol`、`system_title`、
`audit_title`），它们的「译文」本来就该是同一个词。

译文大改、或一次改动之后来不及通读，可以把 `zh` 回退成 `false`，但要在 PR 里
说明为什么。`en`、`de` 由 `internal/web/locales.go` 的 `StrictLangs` 强制为
`true`，不在可以回退的范围里；`fr` 的状态与简体中文无关，各语言各自记录。
