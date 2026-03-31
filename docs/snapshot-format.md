# ko-browser 快照格式规范

> 版本：1.0  
> 本文档定义了 ko-browser 输出的 AX Tree 快照文本格式。  
> 该格式是 ko-browser 的核心差异化设计，面向 LLM/AI Agent 场景做了极致的 token 优化。

---

## 1. 格式总览

ko-browser 的快照是对浏览器 Accessibility Tree（无障碍树）的结构化文本表示。每个可交互或有意义的页面元素被分配一个**递增序号**，LLM 可以直接使用该序号来执行操作。

### 完整示例

```
Page: "百度一下，你就知道"

1: link "新闻"
2: link "hao123"
3: link "地图"
4: textbox "搜索" focused
5: button "百度一下"
6: list
  7: listitem
    8: link "关于百度"
  9: listitem
    10: link "About Baidu"
11: heading "热搜榜"
12: link "植树造林 总书记强调要"利民""
13: img "百度Logo" value="https://www.baidu.com/img/logo.png"
```

---

## 2. 行格式

每个元素占一行，格式为：

```
<indent><id>: <role> "<name>" [value="<value>"] [<state1> <state2> ...]
```

### 各部分说明

| 部分 | 说明 | 是否必须 |
|------|------|---------|
| `<indent>` | 2 空格 × 层级深度，表示树的层次关系 | 是（根层级无缩进） |
| `<id>` | 递增整数序号（从 1 开始），唯一标识该元素 | 是 |
| `:` | 冒号分隔符，紧跟序号后面 | 是 |
| `<role>` | 元素的 ARIA 角色，如 `link`、`button`、`textbox` | 是 |
| `"<name>"` | 元素的可访问名称，用双引号包裹 | 否（无名称时省略） |
| `value="<value>"` | 元素的值（仅在与名称不同时显示） | 否 |
| `<states>` | 元素状态，空格分隔，如 `focused`、`disabled`、`checked` | 否 |

### 示例分解

```
  4: textbox "搜索" focused
  │  │       │       │
  │  │       │       └── 状态：当前聚焦
  │  │       └────────── 名称：搜索
  │  └────────────────── 角色：文本输入框
  └───────────────────── 序号 4，可用 b.Click(4) 操作
```

---

## 3. 根节点格式

页面的根文档节点使用特殊格式：

```
Page: "页面标题"
```

- 根节点**不分配序号**（不可交互）
- 后面紧跟一个空行，再开始列出子元素
- 如果页面无标题，则不输出此行

---

## 4. 序号规则

| 规则 | 说明 |
|------|------|
| **从 1 开始** | 第一个非根元素编号为 1 |
| **递增不跳号** | 严格按深度优先遍历顺序 +1 |
| **根节点不编号** | `Page:` 行不占用序号 |
| **每次快照重新编号** | 序号是临时的，下次 `Snapshot()` 会重新分配 |
| **引用方式** | CLI: `ko-browser click 5`；Library: `b.Click(5)` |

### 序号不持久化

序号在每次快照时重新生成。页面 DOM 变化后，同一个元素的序号可能不同。因此：

- ✅ 获取快照 → 立即使用序号操作
- ❌ 缓存序号跨多次快照使用

---

## 5. 缩进规则

使用 **2 个空格** 表示一级层次：

```
1: navigation "主导航"
  2: link "首页"
  3: link "产品"
    4: link "产品A"
    5: link "产品B"
  6: link "关于"
```

- 根层级（顶层元素）：无缩进
- 每深一层：+2 空格
- 缩进仅用于视觉展示层级关系，不影响序号分配

---

## 6. 名称与值

### 6.1 名称（Name）

- 用双引号包裹：`"搜索"`
- 超过 **80 字符**时截断并加 `...`：`"这是一段很长的文本内容会被截断到八十个字符以内..."`
- 无名称时省略（仅显示 `id: role`）

### 6.2 值（Value）

- 仅当值**不等于名称**时才显示
- 格式：`value="内容"`
- 超过 **50 字符**时截断
- 典型场景：表单输入框的当前输入值

```
4: textbox "用户名" value="admin"
5: textbox "密码"
```

---

## 7. 状态（States）

状态紧跟在名称/值之后，空格分隔：

| 状态 | 含义 | 场景 |
|------|------|------|
| `focused` | 当前聚焦 | 输入框、按钮 |
| `disabled` | 不可用 | 灰色按钮 |
| `checked` | 已勾选 | 复选框、单选框 |
| `expanded` | 已展开 | 下拉菜单、手风琴 |
| `collapsed` | 已折叠 | 折叠面板 |
| `selected` | 已选中 | 标签页、列表项 |
| `required` | 必填 | 表单字段 |
| `readonly` | 只读 | 不可编辑的输入框 |
| `multiline` | 多行 | textarea |

```
3: textbox "邮箱" required focused
7: checkbox "记住我" checked
9: button "提交" disabled
```

---

## 8. 常见角色（Role）

以下是快照中常见的 ARIA 角色：

### 交互角色（可操作）

| 角色 | 说明 | 典型操作 |
|------|------|---------|
| `link` | 超链接 | Click |
| `button` | 按钮 | Click |
| `textbox` | 文本输入框 | Type / Fill |
| `checkbox` | 复选框 | Check / Uncheck |
| `radio` | 单选框 | Click |
| `combobox` | 下拉选择框 | Select |
| `slider` | 滑块 | 暂不支持 |
| `tab` | 标签页 | Click |
| `menuitem` | 菜单项 | Click |
| `searchbox` | 搜索框 | Type / Fill |
| `spinbutton` | 数字步进器 | Type |
| `switch` | 开关 | Click |

### 结构角色（提供上下文）

| 角色 | 说明 |
|------|------|
| `heading` | 标题（h1-h6） |
| `navigation` | 导航区域 |
| `list` | 列表 |
| `listitem` | 列表项 |
| `table` | 表格 |
| `row` | 表格行 |
| `cell` | 表格单元格 |
| `img` | 图片 |
| `banner` | 页头 |
| `main` | 主内容区 |
| `complementary` | 侧边栏 |
| `contentinfo` | 页脚 |
| `dialog` | 对话框 |
| `alert` | 警告提示 |
| `status` | 状态信息 |

---

## 9. 设计决策：为什么用 `id:` 而不是其他格式

### 9.1 与其他工具的格式对比

| 工具 | 格式 | 示例 | 引用方式 |
|------|------|------|---------|
| **agent-browser** | `@eN role "name"` | `@e5 button "Submit"` | `click @e5` |
| **ko-browser v0 (旧)** | `[N] role "name"` | `[5] button "Submit"` | `click 5` |
| **ko-browser v1 (当前)** | `N: role "name"` | `5: button "Submit"` | `click 5` |

### 9.2 为什么选择 `id:` 格式

**1. Token 效率最高**

对于 LLM token 计算，每个字符都很重要：

| 格式 | 文本 | 额外字符数 |
|------|------|-----------|
| `@e5` | `@e5 button "Submit"` | 3 (`@`, `e`, 空格) |
| `[5]` | `[5] button "Submit"` | 3 (`[`, `]`, 空格) |
| `5:` | `5: button "Submit"` | 1 (`:`) |

在一个典型页面快照中（100+ 元素），`id:` 格式比 `[id]` 格式节省约 200 个字符（~50-100 token）。

**2. 引用最简洁**

用户（或 LLM）引用元素时，直接使用数字即可：

```bash
ko-browser click 5      # ← 最简洁
ko-browser click @e5     # ← agent-browser 风格
ko-browser click [5]     # ← 需要转义方括号
```

在 Library API 中同样简洁：

```go
b.Click(5)               // ← 数字即可
```

**3. 对 LLM 友好**

- 数字是所有 LLM tokenizer 最高效编码的内容
- 冒号 `:` 是常见分隔符，不会造成 tokenizer 意外切分
- 无需特殊前缀（`@e`）或包裹字符（`[]`），减少 LLM 出错概率
- LLM 生成 `click 5` 比生成 `click @e5` 或 `click [5]` 更不容易出错

**4. 阅读不模糊**

冒号 `:` 是天然的键值分隔符，`5: button "Submit"` 的语义清晰：
- `5` 是编号
- `button` 是角色
- `"Submit"` 是名称

---

## 10. 与 Snapshot() API 的关系

```go
snap, _ := b.Snapshot()

// snap.Text 就是本文档定义的格式化文本
fmt.Println(snap.Text)
// 输出:
// Page: "百度一下，你就知道"
//
// 1: link "新闻"
// 2: link "hao123"
// 3: textbox "搜索" focused
// 4: button "百度一下"

// snap.IDMap 是 id → BackendDOMNodeID 的映射
// 当你调用 b.Click(3) 时，内部通过 IDMap 查找到 CDP 节点 ID

// snap.Nodes 是完整的过滤后树结构，供需要自定义处理的用户使用
```

---

## 11. 完整格式 BNF

```bnf
<snapshot>    ::= [<page-line> "\n\n"] <element-lines>
<page-line>   ::= 'Page: "' <title> '"'
<element-lines> ::= (<element-line> "\n")*
<element-line>  ::= <indent> <id> ": " <role> [" " <quoted-name>] [" " <value>] [" " <states>]
<indent>      ::= ("  ")*
<id>          ::= [1-9][0-9]*
<role>        ::= [a-z]+
<quoted-name> ::= '"' <text-80> '"'
<value>       ::= 'value="' <text-50> '"'
<states>      ::= <state> (" " <state>)*
<state>       ::= "focused" | "disabled" | "checked" | "expanded" | "collapsed"
                 | "selected" | "required" | "readonly" | "multiline"
<text-80>     ::= .{1,80} | .{1,77} "..."
<text-50>     ::= .{1,50} | .{1,47} "..."
```

---

## 12. 示例对比

### 12.1 简单搜索页面

```
Page: "Google"

1: combobox "搜索" focused
2: button "Google 搜索"
3: button "手气不错"
4: link "Gmail"
5: link "图片"
```

### 12.2 带表单的页面

```
Page: "登录 - GitHub"

1: link "GitHub"
2: heading "登录到 GitHub"
3: textbox "用户名或邮箱" required
4: textbox "密码" required
5: link "忘记密码？"
6: button "登录"
7: link "创建账户"
```

### 12.3 复杂列表页面

```
Page: "Hacker News"

1: link "Hacker News"
2: link "new"
3: link "past"
4: link "comments"
5: table
  6: row
    7: link "Show HN: A tool I built"
    8: link "example.com"
    9: link "42 points"
    10: link "username"
    11: link "15 comments"
  12: row
    13: link "Why Rust is awesome"
    14: link "blog.example.com"
    15: link "128 points"
    16: link "another_user"
    17: link "67 comments"
18: link "More"
```

### 12.4 LLM 交互示例

```
LLM 收到快照:
  Page: "Google"
  1: combobox "搜索" focused
  2: button "Google 搜索"
  3: button "手气不错"

LLM 决策: 在搜索框输入并搜索
LLM 输出:
  fill 1 "Go 语言教程"
  click 2

→ ko-browser 执行:
  b.Fill(1, "Go 语言教程")
  b.Click(2)
```

---

## 13. 变更历史

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2026-03-31 | 初始版本：确定 `id: role "name" states` 格式 |
| — | （之前） | 旧格式 `[id] role "name" states`，因 `[` `]` 浪费 token 而废弃 |
