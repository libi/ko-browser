# ko-browser Snapshot Format Specification

> Version: 1.0  
> This document defines the AX Tree snapshot text format output by ko-browser.  
> This format is a core differentiating design of ko-browser, with extreme token optimization for LLM/AI Agent scenarios.

---

## 1. Format Overview

The ko-browser snapshot is a structured text representation of the browser's Accessibility Tree (AX Tree). Each interactive or meaningful page element is assigned an **incremental ID**, which LLMs can directly use to perform actions.

### Full Example

```
Page: "Google"

1: link "Gmail"
2: link "Images"
3: link "Maps"
4: textbox "Search" focused
5: button "Google Search"
6: list
  7: listitem
    8: link "About Google"
  9: listitem
    10: link "Advertising"
11: heading "Trending"
12: link "Breaking news: Major tech announcement"
13: img "Google Logo" value="https://www.google.com/images/logo.png"
```

---

## 2. Line Format

Each element occupies one line, formatted as:

```
<indent><id>: <role> "<name>" [value="<value>"] [<state1> <state2> ...]
```

### Field Descriptions

| Field | Description | Required |
|-------|-------------|----------|
| `<indent>` | 2 spaces × nesting depth, representing tree hierarchy | Yes (no indent at root level) |
| `<id>` | Incremental integer ID (starting from 1), uniquely identifies the element | Yes |
| `:` | Colon separator, immediately follows the ID | Yes |
| `<role>` | Element's ARIA role, e.g., `link`, `button`, `textbox` | Yes |
| `"<name>"` | Element's accessible name, wrapped in double quotes | No (omitted when no name) |
| `value="<value>"` | Element's value (only shown when different from name) | No |
| `<states>` | Element states, space-separated, e.g., `focused`, `disabled`, `checked` | No |

### Example Breakdown

```
  4: textbox "Search" focused
  │  │       │        │
  │  │       │        └── State: currently focused
  │  │       └─────────── Name: Search
  │  └─────────────────── Role: text input box
  └────────────────────── ID 4, operable via b.Click(4)
```

---

## 3. Root Node Format

The page's root document node uses a special format:

```
Page: "Page Title"
```

- The root node is **not assigned an ID** (not interactive)
- Followed by a blank line, then child elements begin
- If the page has no title, this line is omitted

---

## 4. ID Rules

| Rule | Description |
|------|-------------|
| **Starts from 1** | The first non-root element is numbered 1 |
| **Incremental, no gaps** | Strictly follows depth-first traversal order, incrementing by 1 |
| **Root not numbered** | The `Page:` line does not consume an ID |
| **Re-numbered per snapshot** | IDs are temporary; the next `Snapshot()` call will reassign them |
| **Reference method** | CLI: `ko-browser click 5`; Library: `b.Click(5)` |

### IDs Are Not Persistent

IDs are regenerated with each snapshot. After DOM changes, the same element may receive a different ID. Therefore:

- ✅ Take snapshot → immediately use IDs to interact
- ❌ Cache IDs across multiple snapshots

---

## 5. Indentation Rules

Use **2 spaces** to represent one level of hierarchy:

```
1: navigation "Main Nav"
  2: link "Home"
  3: link "Products"
    4: link "Product A"
    5: link "Product B"
  6: link "About"
```

- Root level (top-level elements): no indentation
- Each deeper level: +2 spaces
- Indentation is for visual hierarchy only and does not affect ID assignment

---

## 6. Name and Value

### 6.1 Name

- Wrapped in double quotes: `"Search"`
- Truncated with `...` when exceeding **80 characters**: `"This is a very long text content that will be truncated to stay within eighty..."`
- Omitted when no name (only `id: role` is shown)

### 6.2 Value

- Only displayed when the value **differs from the name**
- Format: `value="content"`
- Truncated when exceeding **50 characters**
- Typical scenario: current input value of a form field

```
4: textbox "Username" value="admin"
5: textbox "Password"
```

---

## 7. States

States follow the name/value, space-separated:

| State | Meaning | Scenario |
|-------|---------|----------|
| `focused` | Currently focused | Input fields, buttons |
| `disabled` | Not available | Grayed-out buttons |
| `checked` | Checked/selected | Checkboxes, radio buttons |
| `expanded` | Expanded | Dropdowns, accordions |
| `collapsed` | Collapsed | Collapsible panels |
| `selected` | Selected | Tabs, list items |
| `required` | Required | Form fields |
| `readonly` | Read-only | Non-editable inputs |
| `multiline` | Multi-line | textarea |

```
3: textbox "Email" required focused
7: checkbox "Remember me" checked
9: button "Submit" disabled
```

---

## 8. Common Roles

Below are common ARIA roles found in snapshots:

### Interactive Roles (Actionable)

| Role | Description | Typical Action |
|------|-------------|----------------|
| `link` | Hyperlink | Click |
| `button` | Button | Click |
| `textbox` | Text input field | Type / Fill |
| `checkbox` | Checkbox | Check / Uncheck |
| `radio` | Radio button | Click |
| `combobox` | Dropdown select | Select |
| `slider` | Slider | Not yet supported |
| `tab` | Tab | Click |
| `menuitem` | Menu item | Click |
| `searchbox` | Search box | Type / Fill |
| `spinbutton` | Number stepper | Type |
| `switch` | Toggle switch | Click |

### Structural Roles (Provide Context)

| Role | Description |
|------|-------------|
| `heading` | Heading (h1-h6) |
| `navigation` | Navigation region |
| `list` | List |
| `listitem` | List item |
| `table` | Table |
| `row` | Table row |
| `cell` | Table cell |
| `img` | Image |
| `banner` | Page header |
| `main` | Main content area |
| `complementary` | Sidebar |
| `contentinfo` | Page footer |
| `dialog` | Dialog box |
| `alert` | Alert message |
| `status` | Status information |

---

## 9. Design Decisions: Why `id:` Instead of Other Formats

### 9.1 Format Comparison with Other Tools

| Tool | Format | Example | Reference Method |
|------|--------|---------|------------------|
| **agent-browser** | `@eN role "name"` | `@e5 button "Submit"` | `click @e5` |
| **ko-browser v0 (old)** | `[N] role "name"` | `[5] button "Submit"` | `click 5` |
| **ko-browser v1 (current)** | `N: role "name"` | `5: button "Submit"` | `click 5` |

### 9.2 Why the `id:` Format Was Chosen

**1. Highest Token Efficiency**

For LLM token counting, every character matters:

| Format | Text | Extra Characters |
|--------|------|-----------------|
| `@e5` | `@e5 button "Submit"` | 3 (`@`, `e`, space) |
| `[5]` | `[5] button "Submit"` | 3 (`[`, `]`, space) |
| `5:` | `5: button "Submit"` | 1 (`:`) |

In a typical page snapshot (100+ elements), the `id:` format saves approximately 200 characters (~50-100 tokens) compared to the `[id]` format.

**2. Most Concise References**

When users (or LLMs) reference elements, they simply use the number:

```bash
ko-browser click 5      # ← most concise
ko-browser click @e5     # ← agent-browser style
ko-browser click [5]     # ← requires bracket escaping
```

Equally concise in the Library API:

```go
b.Click(5)               // ← just a number
```

**3. LLM-Friendly**

- Numbers are the most efficiently encoded content across all LLM tokenizers
- The colon `:` is a common separator that doesn't cause unexpected tokenizer splits
- No special prefix (`@e`) or wrapping characters (`[]`), reducing LLM error probability
- LLMs are less likely to make mistakes generating `click 5` than `click @e5` or `click [5]`

**4. Unambiguous Reading**

The colon `:` is a natural key-value separator. `5: button "Submit"` has clear semantics:
- `5` is the ID
- `button` is the role
- `"Submit"` is the name

---

## 10. Relationship with the Snapshot() API

```go
snap, _ := b.Snapshot()

// snap.Text is the formatted text defined in this document
fmt.Println(snap.Text)
// Output:
// Page: "Google"
//
// 1: link "Gmail"
// 2: link "Images"
// 3: textbox "Search" focused
// 4: button "Google Search"

// snap.IDMap is the id → BackendDOMNodeID mapping
// When you call b.Click(3), it internally looks up the CDP node ID via IDMap

// snap.Nodes is the complete filtered tree structure, for users who need custom processing
```

---

## 11. Full Format BNF

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

## 12. Examples

### 12.1 Simple Search Page

```
Page: "Google"

1: combobox "Search" focused
2: button "Google Search"
3: button "I'm Feeling Lucky"
4: link "Gmail"
5: link "Images"
```

### 12.2 Page with Form

```
Page: "Sign in to GitHub"

1: link "GitHub"
2: heading "Sign in to GitHub"
3: textbox "Username or email" required
4: textbox "Password" required
5: link "Forgot password?"
6: button "Sign in"
7: link "Create an account"
```

### 12.3 Complex List Page

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

### 12.4 LLM Interaction Example

```
LLM receives snapshot:
  Page: "Google"
  1: combobox "Search" focused
  2: button "Google Search"
  3: button "I'm Feeling Lucky"

LLM decides: type in search box and search
LLM output:
  fill 1 "Go language tutorial"
  click 2

→ ko-browser executes:
  b.Fill(1, "Go language tutorial")
  b.Click(2)
```

---

## 13. Changelog

| Version | Date | Change |
|---------|------|--------|
| 1.0 | 2026-03-31 | Initial version: established `id: role "name" states` format |
| — | (prior) | Old format `[id] role "name" states`, deprecated due to `[` `]` wasting tokens |
