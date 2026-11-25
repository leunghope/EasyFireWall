---
layout: default
title: 贡献指南
description: 三条命令完成环境搭建、评审会检查的规则，以及设计系统所在的位置。
---

# 贡献指南

完整指南：[CONTRIBUTING.md](https://github.com/jp1337/easywall/blob/main/CONTRIBUTING.md)。本页是精简版。

```bash
git clone https://github.com/jp1337/easywall.git && cd easywall
go mod download
make test lint
```

| | |
|---|---|
| 工具链 | 以 `go.mod` 里 `toolchain` 那一行为准 —— 绝不在其他任何地方写死版本号 |
| 提交 | [Conventional Commits](https://www.conventionalcommits.org/)：`feat`、`fix`、`security`、`docs`、`chore`、`refactor`、`test` |
| 覆盖率 | 不得低于 80% |
| 任何视觉相关的东西 | 先经过 [`DESIGN.md`](https://github.com/jp1337/easywall/blob/main/DESIGN.md) |

## 重建已提交的构建产物

有三份构建产物已提交到仓库，因此发版不需要 Node 工具链。

```bash
npm run build:css         # web/static/style.css      — the application
npm run build:docs-css    # docs/assets/css/style.css — this site
npm run build:diagrams    # docs/assets/diagrams/*.svg — two per source
```

CI 会重建这两个样式表，只要有差异就失败。图表渲染无法逐字节复现 —— mermaid 在两次运行之间会让曲线控制点抖动。不要靠重建并比对来验证图表；`npm run check:diagrams` 改为校验摘要，如果某个 `.mmd` 源文件改了却没有重新渲染，它就会失败。`npx @google/design.md lint DESIGN.md` 校验设计系统；有些警告是预期之内的，原因写在文件内部。

## 评审检查什么

| | |
|---|---|
| **颜色即状态** | `state-ok`、`state-warn` 与 `state-crit` 就是整套调色板 —— 生效、未确认、已回滚。数量不是状态，而一个没有颜色的界面就是没什么可报告的界面 |
| **焦点、选中态与主操作不带色相** | 改用填充、边缘与字重。仅靠背景从不标记当前项：`select-fill` 相对其所在表面的对比度只有 1.08:1 |
| **控件与容器** | 控件的描边是 `control-edge`（3:1，WCAG 1.4.11）；容器的描边是 `rule` |
| **双主题、三种宽度** | 浅色与深色，1600 / 900 / 390 px |
| **句首大写，Inter** | 带字距的大写等宽 `label` 角色只保留在侧边栏分隔线里 |
| **表格会重排** | 宽度低于 720px 时行变成卡片，前提是每个 `<td>` 都带 `data-label` |
| **一件事只有一个标题** | 页面标题与同名卡片并存，就是重复标题的 bug |
| **每个字符串都翻译** | {% raw %}`{{T "key"}}`{% endraw %} 要写进 `locales/en.json` 与 `locales/de.json` **两者**，属性也包括在内。只要求这两个 —— 后续可能再加语言，在它补齐之前会渲染为 English 原文 |
| **截图跟随界面变化** | 你改过的页面要重拍 `docs/assets/img/screens/*`，双主题，在同一个 pull request 里 |

**把改动渲染出来看。** 这个界面里每一个值得抓住的缺陷，在样式表里都看不出来，在截图里却一目了然：

- 被截断的端口号
- 一个已经不存在的 class
- 整个页面落在白色背景上

## 新增一门语言

1. 复制 `locales/en.json` 为 `locales/<lang>.json`
2. 翻译其中的值；所有 `id` 保持不动
3. 把 `language_name` 译成该语言**自己**的名字 —— 是 `Deutsch` 而不是 `German`；是 `Français` 而不是 `French`；是 `简体中文` 而不是 `Chinese`
4. 在 `locales/status.json` 中加入 `"<lang>": { "reviewed": false }`
5. PR 标题为 `feat(i18n): add <language> translation`

要让一门语言出现，不需要做其他任何事 —— 切换器是根据 `locales/*.json` 里的内容构建的。

**不完整的翻译同样欢迎。** 只有 `en` 与 `de` 被要求逐条对齐。另一种语言里你没有译的键会渲染成 English 原文，缺口会被报告而不是隐藏 —— `go test ./internal/web/ -run TestLocaleCoverage` 会打印百分比。与英文逐字节相同的值不算已翻译，所以这个数字无法靠粘贴抬高。

`reviewed: false` 表示还没有会说这门语言的人从头到尾读过它，切换器也会这样标注。等你读完之后把标志改成 `true`，本身就是一个单独的 pull request，其价值不亚于翻译本身。

有三种行内形式必须在你的译文里保留下来：

| 消息里的写法 | 渲染为 |
|---|---|
| `` `443` `` | `<code>443</code>` |
| `*before*` | `<em>before</em>` —— 承载语义的强调 |
| `{}` | 一个链接，**按你的语序** |

## 文档

信息最大化，文字最小化：能用图就不用段落，能用表格就不用一串句子，能用截图就不用文字描述界面。

本站是写给运行 easywall 的人看的。写给*维护*它的人看的笔记 —— CI 流水线、打包陷阱、守卫测试 —— 放在仓库的 `docs-tech/` 里，有意不在此发布。

## 安全问题

**不要**作为公开 issue 提出，也**不要**发在 Discord —— 请用 [GitHub Security Advisories](https://github.com/jp1337/easywall/security/advisories/new)。欢迎研究，研究没有报酬；负责任的报告会得到署名致谢。

## 由 AI 代理撰写

欢迎，但要说明：在 issue 或 pull request 描述开头，用两行写上 `> [!WARNING]` 与 `> AI-generated`。这类帖子会被标记 `ai-generated` 以便评审，而不是被拒绝。

## 提问而不是报告缺陷

任何属于提问而不是缺陷的内容，都请放到 [Discord]({{ site.discord }}) —— 包括「这是预期行为吗？」，这类问题常常是缺陷的起点。