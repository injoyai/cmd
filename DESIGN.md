---
version: alpha
name: "i 配置"
description: "CLI 配套的紧凑桌面配置表单"
colors:
  background: "#f5f6f8"
  surface: "#ffffff"
  text: "#172033"
  muted: "#5c687b"
  border: "#dce2e8"
  primary: "#0c6b62"
  primary-hover: "#09574f"
  primary-soft: "#eaf5f2"
  danger: "#b42318"
typography:
  sans:
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif'
  mono:
    fontFamily: '"SFMono-Regular", Consolas, monospace'
rounded:
  DEFAULT: "6px"
  panel: "8px"
spacing:
  page-max: "1120px"
  section-gap: "18px"
components:
  field:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.DEFAULT}"
  field-hint:
    textColor: "{colors.muted}"
  field-error:
    textColor: "{colors.danger}"
  control-border:
    backgroundColor: "{colors.border}"
  button:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.surface}"
  button-hover:
    backgroundColor: "{colors.primary-hover}"
  field-focus:
    backgroundColor: "{colors.primary-soft}"
  section:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.panel}"
---

# i 配置界面

## Overview

范围为 `global/global.html` 的单页配置表单。面向直接操作 CLI 的中文用户，以桌面使用为主，并支持窄窗口。采用系统设置面板的密度：白色分组表单、左侧分组导航、固定顶部保存操作。配置点路径以小号等宽字显示，方便与 CLI 对照。无外部字体、装饰图、统计卡片或营销横幅。

## Colors

运行时 CSS 自定义属性是规范来源，本文件记录相同的已实施值；不新增主题适配或生成链。`background/surface/text/muted/border` 对应 `--bg/--surface/--text/--muted/--border`；`primary/primary-hover/primary-soft/danger` 对应 `--accent/--accent-hover/--accent-soft/--danger`。控件只消费这些变量。青绿色用于主操作、焦点和保存成功，红色同时配合文字说明错误。

## Typography

`--font` 使用上述系统中文无衬线字体，正文 14px、行高 1.6，页面标题 19px，分组标题 15px。`--mono` 用于配置点路径和资源地址，不把正文全部设为代码字体。保留长地址换行能力。

## Layout

顶部工具栏最小高度 72px。桌面页面最大宽度 1120px，左右留白 28px，导航宽 144px。每组表单间距 18px；标签与输入区使用 138px / 自适应两列。760px 以下字段改为上下排列，540px 以下导航变为水平滚动。页面自然纵向滚动，标题只出现一次，不裁剪表单；顶部操作和分组锚点避让固定工具栏。

## Elevation & Depth

用白色面板与 1px 边界表达层级；配置页不使用阴影、模糊或渐变。状态文字预留高度，保存按钮宽度固定。

## Shapes

控件圆角使用 `--radius` 6px，分组面板 8px。开关保留原生 checkbox 的键盘行为；不实现自定义选择器或弹窗。

## Components

表单及字段由 `loadConfig` / `renderField` 统一生成，配置值通过 Go 的 `saveGUIConfig` 写入。无共享前端组件库，不复制其它页面的输入或反馈实现。

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
|---|---|---|---|---|
| Form | renderField / collectConfig | 本文件与配置元数据 | 文本、密钥、开关、地址列表、快捷映射 | `node --test global/config_ui_test.js`、Go 保存测试 |
| Scrollbar | 页面全局 CSS | global/global.html | 纵向页面、窄屏横向导航 | CSS 与静态检查 |
| Toast | showStatus 状态区域 | 本文件 | 普通、成功、错误 | 保存成功、失败、重试脚本测试 |

保存期间禁用编辑和重复提交，失败时保留输入并允许重试；状态通过 live region 宣告。快捷名称为空或重复时显示关联错误并聚焦输入。密钥默认隐藏，显隐按钮可用键盘操作。快捷删除只修改草稿，保存后生效；离开有未保存修改的窗口使用浏览器卸载提醒。禁用动画偏好下关闭控件过渡。输入不强制协议或 URL 格式。

## Do's and Don'ts

保留未展示的自定义配置，不把整段 MinIO 配置覆盖成表单字段。资源列表保留逗号和等号。不要把测速配置重新加入界面，也不要为了单页配置引入框架、路由、网络字体或额外生产依赖。
