<p align="center">
  <img src="web/static/icon.svg" alt="" width="88" height="88">
</p>

<h1 align="center">easywall</h1>

<p align="center"><em>你的防火墙。你的规则。没有意外。</em></p>

<!-- 统一使用同一家徽章服务，这一组徽章因此共享形状、字体和 logo 处理。
     三个状态徽章按名字读取工作流；版本、许可证和 Go 徽章读取仓库与
     go.mod，所以没有哪一枚会像手写的版本号那样走样。 -->
<p align="center">
  <a href="https://github.com/jp1337/easywall/actions/workflows/test.yml"><img src="https://img.shields.io/github/actions/workflow/status/jp1337/easywall/test.yml?branch=main&label=tests&logo=github&logoColor=white" alt="Tests"></a>
  <a href="https://github.com/jp1337/easywall/actions/workflows/build.yml"><img src="https://img.shields.io/github/actions/workflow/status/jp1337/easywall/build.yml?branch=main&label=build&logo=github&logoColor=white" alt="Build"></a>
  <a href="https://github.com/jp1337/easywall/actions/workflows/security.yml"><img src="https://img.shields.io/github/actions/workflow/status/jp1337/easywall/security.yml?branch=main&label=security&logo=github&logoColor=white" alt="Security"></a>
  <a href="https://codecov.io/gh/jp1337/easywall"><img src="https://img.shields.io/codecov/c/github/jp1337/easywall/main?logo=codecov&logoColor=white&label=coverage" alt="Coverage"></a>
</p>

<p align="center">
  <a href="https://github.com/jp1337/easywall/releases/latest"><img src="https://img.shields.io/github/v/release/jp1337/easywall?logo=github&logoColor=white&label=release" alt="Latest release"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/github/go-mod/go-version/jp1337/easywall?logo=go&logoColor=white&label=go" alt="Go version"></a>
  <a href="https://www.gnu.org/licenses/gpl-3.0"><img src="https://img.shields.io/github/license/jp1337/easywall?logo=opensourceinitiative&logoColor=white&label=license&color=blue" alt="GPL-3.0"></a>
  <a href="https://discord.gg/3zJMvChvUA"><img src="https://img.shields.io/badge/discord-join-5865F2?logo=discord&logoColor=white" alt="Discord"></a>
  <a href="https://ko-fi.com/jp1337"><img src="https://img.shields.io/badge/ko--fi-support-13C3FF?logo=kofi&logoColor=white" alt="Support on Ko-fi"></a>
</p>

<p align="center">
  <a href="https://demo.easywall-project.org"><strong>在线演示</strong></a> ·
  <a href="https://easywall-project.org"><strong>文档</strong></a> ·
  <a href="CHANGELOG.md">更新日志</a>
</p>

通过一个不会把你锁在门外的 Web 界面管理 nftables：**每次应用都会自动回滚，除非你确认。**

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/img/screens/dashboard-dark.png">
  <img src="docs/assets/img/screens/dashboard-light.png" alt="easywall 仪表盘：防火墙状态，含确认状态、待处理更改和上次应用；还有统计 TCP 端口、UDP 端口、拦截列表、允许列表、自定义规则和端口转发的卡片；以及最近活动列表。">
</picture>

## 设计思路

编辑一条规则不会改变任何东西。应用会改变一切 —— 为期 120 秒。如果新规则切断了你的连接，你就无法点击确认，而*不*确认正是让旧规则回来的原因。

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/diagrams/apply-flow-dark.svg">
  <img src="docs/assets/diagrams/apply-flow-light.svg" alt="状态机：编辑通向暂存，应用通向生效，在窗口内确认通向已确认，让窗口过期则通向已回滚；从已回滚状态仍可取用暂存的改动。">
</picture>

## 架构

两个进程。暴露在网络上的那个不持有任何值得窃取的权限。

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/diagrams/architecture-dark.svg">
  <img src="docs/assets/diagrams/architecture-light.svg" alt="浏览器通过 HTTPS 与 easywall-web 通信，后者以无权限身份运行；easywall-web 通过 Unix socket 以类型化 JSON 与 easywall-core 通信，后者以 root 运行，并通过 netlink 与 nftables 表 inet easywall 通信。">
</picture>

对初版 easywall 的彻底重写 —— 那版用 Python、Flask，通过 subprocess 调用 `iptables`，在 2022 年一次 CVE 之后被归档。两个根因都已消除：权限位于另一个进程，应用路径构建的是 Go 结构体，而不是命令行。[工作原理 →](https://easywall-project.org/architecture/)

## 安装

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/diagrams/install-choice-dark.svg">
  <img src="docs/assets/diagrams/install-choice-light.svg" alt="决策树：只是看看就走演示模式；Debian 或 Ubuntu 走 .deb 包；已经在跑容器就走 Docker；其余情况从源码构建。">
</picture>

```bash
# Debian / Ubuntu — amd64 and arm64
ARCH=$(dpkg --print-architecture)
wget https://github.com/jp1337/easywall/releases/latest/download/easywall_$ARCH.deb
sudo dpkg -i easywall_$ARCH.deb && sudo apt-get install -f

# Docker
git clone https://github.com/jp1337/easywall.git && cd easywall && docker compose up -d

# From source — Go 1.27+, nftables
git clone https://github.com/jp1337/easywall.git && cd easywall
make build && sudo make install
sudo systemctl enable --now easywall-core easywall-web
```

然后打开 `https://localhost:12227`。首次访问会[创建账户并暂存第一批规则](https://easywall-project.org/installation/first-run/)。

## 功能概览

| | |
|---|---|
| **端口** | TCP 与 UDP，单个端口或端口区间，每条规则可单独启用 SSH 暴力破解限速器 |
| **拦截列表与允许列表** | IPv4、IPv6 和 CIDR，先于任何端口规则评估 |
| **防护模块** | 共十二个，默认开启五个 —— 泛洪、扫描、无效地址、分片、广播/组播/任播 |
| **端口转发** | 带协议选择的 NAT 重定向 |
| **自定义规则** | 原始 nftables，应用前先做语法检查 |
| **导出 / 导入** | 整套规则集以 JSON 表示 |
| **审计日志** | 改了什么、何时改的，每行一个 JSON 对象 |
| **与 Docker 共存** | 独占 `table inet easywall`，不触碰其他任何东西 |
| **English, Deutsch, Français, 简体中文** | 可在界面中切换，登录前也可以。语言可以是部分的：缺失的键回退为英文，缺口会被报告而不是隐藏 |
| **亮色与暗色** | 跟随系统，也可手动切换；两种主题都做过对比度检查 |

## 技术栈

| | |
|---|---|
| Go 1.27, single binary | `go-chi/chi` · `html/template` |
| 通过 `google/nftables` 操作 nftables | 类型化规则走 netlink；`nft` 只用于自定义规则 |
| Argon2id | `golang.org/x/crypto`，每个密码 16 字节盐 |
| CSRF | `net/http.CrossOriginProtection`，Go 1.25 原生 |
| 设计系统 | [`DESIGN.md`](DESIGN.md) + Tailwind v4 —— 不含第三方 UI 库 |
| 字体 | Inter + JetBrains Mono，自托管，约 145 KB —— 可在隔离网络中运行 |
| CI | `govulncheck`、`gosec`、CodeQL、`-race`，以及一套面向真实内核的集成测试 |

## 获取帮助

| | |
|---|---|
| 有问题要问，或哪里不对劲 | [Discord](https://discord.gg/3zJMvChvUA) |
| 报告缺陷，或提出想要的功能 | [GitHub issues](https://github.com/jp1337/easywall/issues) |
| 安全漏洞 | [安全公告](https://github.com/jp1337/easywall/security/advisories/new) —— **不要**发到 Discord，也不要开公开 issue |

## 参与贡献

开发环境搭建、提交约定和审查清单：[CONTRIBUTING.md](CONTRIBUTING.md)。任何视觉相关的内容都先经过 [`DESIGN.md`](DESIGN.md)。

安全问题：**不要**开公开 issue —— 使用 [GitHub Security Advisories](https://github.com/jp1337/easywall/security/advisories/new)。

## 许可证

GPL-3.0 —— 见 [LICENSE](LICENSE)。
