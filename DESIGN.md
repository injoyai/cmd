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
rounded:
  DEFAULT: "6px"
  panel: "8px"
spacing:
  page-max: "820px"
  section-gap: "16px"
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

范围为 `config/gui.html` 的单页配置表单。面向直接操作 CLI 的中文用户，以桌面使用为主，并支持窄窗口。采用系统设置面板的密度：白色分组表单、单列页面、固定顶部保存操作。使用原生复选框，减少控件样式与重复渲染代码。无外部字体、装饰图、统计卡片或营销横幅。

## Colors

运行时 CSS 自定义属性是规范来源，本文件记录相同的已实施值；不新增主题适配或生成链。`background/surface/text/muted/border` 对应 `--bg/--surface/--text/--muted/--border`；`primary/primary-hover/primary-soft/danger` 对应 `--accent/--accent-hover/--accent-soft/--danger`。控件只消费这些变量。青绿色用于主操作、焦点和保存成功，红色同时配合文字说明错误。

## Typography

`--font` 使用上述系统中文无衬线字体，正文 14px、行高 1.6，页面标题 19px，分组标题 15px。配置项只展示用户可读的名称，省去重复的点路径标签。保留长地址换行能力。

## Layout

顶部工具栏随内容自然增高并固定在页面顶部。页面最大宽度 820px，左右留白 24px；每组表单间距 16px，标签与输入区使用 124px / 自适应两列。580px 以下字段改为上下排列、左右留白改为 16px。页面自然纵向滚动，不设置侧栏或分组锚点，保证表单完整可达。

## Elevation & Depth

用白色面板与 1px 边界表达层级；配置页不使用阴影、模糊或渐变。状态文字预留高度，保存按钮宽度固定。

## Shapes

控件圆角使用 `--radius` 6px，分组面板 8px。开关保留原生 checkbox 的键盘行为；不实现自定义选择器或弹窗。

密钥显隐、添加和移除使用 38px 方形图标按钮，图标为内联 SVG；保留 `title` 悬停提示和 `aria-label`，显隐状态同时更新图标和提示。顶部保存操作保留文字。

## Components

表单及字段由 `loadConfig` / `renderField` 统一生成，配置值通过 Go 的 `saveGUIConfig` 写入。无共享前端组件库，不复制其它页面的输入或反馈实现。

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
|---|---|---|---|---|
| Form | renderField / collectConfig | 本文件与配置元数据 | 文本、密钥、开关、地址列表、快捷映射 | `node --test config/gui_test.js`、Go 保存测试 |
| Scrollbar | 页面全局 CSS | config/gui.html | 纵向页面 | CSS 与静态检查 |
| Toast | showStatus 状态区域 | 本文件 | 普通、成功、错误 | 保存成功、失败、重试脚本测试 |

保存期间禁用编辑和重复提交，失败时保留输入并允许重试；状态通过 live region 宣告。快捷名称为空或重复时显示关联错误并聚焦输入。密钥默认隐藏，显隐按钮可用键盘操作。快捷删除只修改草稿，保存后生效；关闭窗口直接退出，不弹出未保存修改提示。不使用控件动画。输入不强制协议或 URL 格式。

## Do's and Don'ts

保留未展示的自定义配置，不把整段 MinIO 配置覆盖成表单字段。资源列表保留逗号和等号。不要把测速配置重新加入界面，也不要为了单页配置引入框架、路由、网络字体或额外生产依赖。

页面由 lorca 编码到 Chrome 启动参数；`TestConfigGUIWindowsCommandLineBudget` 为其他 Chrome 参数预留 8192 字符，防止 Windows 命令行超长。精简时保留可读源码，不使用运行时压缩器或额外服务。
