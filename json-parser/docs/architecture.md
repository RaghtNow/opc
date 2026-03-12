# JSON Parser Tool - 架构文档

## 1. 项目概述

- **项目名称**: JSON Parser Tool
- **项目类型**: Web 工具应用
- **技术栈**: Vue 3 + TypeScript + Vite
- **版本**: 1.0.0

## 2. 功能列表

### 2.1 核心功能
- ✅ JSON 解析 - 将输入的 JSON 字符串解析为 JavaScript 对象
- ✅ 格式化输出 - 美化 JSON，支持 2/4 空格缩进
- ✅ 压缩输出 - 移除所有空白，输出紧凑 JSON
- ✅ 错误验证 - 提供详细的错误信息，包括错误位置（行/列）

### 2.2 UI 功能
- ✅ 输入区 - 支持粘贴/输入 JSON 字符串
- ✅ 输出区 - 显示格式化/压缩后的结果
- ✅ 复制功能 - 一键复制结果到剪贴板
- ✅ 示例数据 - 加载示例 JSON
- ✅ 清空功能 - 清除所有内容
- ✅ 验证状态 - 显示 JSON 类型、键数量、字符数等信息

## 3. 项目结构

```
json-parser/
├── index.html              # 入口 HTML
├── package.json            # 项目依赖
├── vite.config.ts          # Vite 配置
├── tsconfig.json           # TypeScript 配置
├── tsconfig.node.json      # Node TypeScript 配置
├── src/
│   ├── main.ts             # 入口文件
│   ├── App.vue             # 根组件
│   ├── vite-env.d.ts       # Vite 类型声明
│   ├── components/
│   │   └── JsonParser.vue  # 主组件
│   └── utils/
│       └── jsonUtils.ts    # JSON 工具函数
└── docs/
    └── architecture.md     # 本文档
```

## 4. 核心模块

### 4.1 jsonUtils.ts

| 函数 | 功能 | 返回值 |
|------|------|--------|
| `parseJson(input)` | 解析 JSON 字符串 | `{ success, data?, error?, errorLine?, errorColumn? }` |
| `formatJson(data, indent)` | 格式化 JSON | 格式化后的字符串 |
| `minifyJson(data)` | 压缩 JSON | 压缩后的字符串 |
| `validateJson(input)` | 验证 JSON | 详细信息对象 |
| `isValidJsonStructure(input)` | 检查是否有效 | boolean |

### 4.2 JsonParser.vue

| 功能 | 描述 |
|------|------|
| 输入区 | Textarea，支持大文本输入 |
| 模式选择 | 美化/压缩模式切换 |
| 缩进选择 | 2 空格/4 空格 |
| 解析按钮 | 执行解析和转换 |
| 验证状态 | 显示类型、大小、键数量等 |
| 错误显示 | 显示错误信息和位置 |
| 复制按钮 | 复制输出到剪贴板 |
| 示例加载 | 加载示例 JSON |
| 清空 | 重置所有状态 |

## 5. API 设计

### 5.1 JSON 工具函数

```typescript
interface ParseResult {
  success: boolean;
  data?: any;
  error?: string;
  errorLine?: number;
  errorColumn?: number;
}

interface ValidationResult {
  isValid: boolean;
  type?: string;
  size?: number;
  keys?: string[];
  keyCount?: number;
  length?: number;
  error?: string;
  errorLine?: number;
  errorColumn?: number;
}
```

## 6. 页面布局

```
┌─────────────────────────────────────────┐
│           🔧 JSON Parser Tool           │
│            解析 · 格式化 · 验证          │
├─────────────────────────────────────────┤
│  📥 输入 JSON                            │
│  ┌─────────────────────────────────────┐ │
│  │                                     │ │
│  │     JSON 输入区域 (300px min)       │ │
│  │                                     │ │
│  └─────────────────────────────────────┘ │
│  [❌ 错误信息]                            │
│  ┌──────────────┐  ┌─────────────────┐  │
│  │ ○ 美化       │  │ 🔄 解析 & 转换   │  │
│  │   缩进: 2 ▼  │  │                 │  │
│  │ ○ 压缩       │  └─────────────────┘  │
│  └──────────────┘                        │
├─────────────────────────────────────────┤
│  ✅ 有效 JSON (object, 256 字符) · 5个键  │
├─────────────────────────────────────────┤
│  📤 输出结果                   [📋复制]  │
│  ┌─────────────────────────────────────┐ │
│  │                                     │ │
│  │     输出结果区域 (只读)              │ │
│  │                                     │ │
│  └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

## 7. 开发进度

| 阶段 | 状态 | 完成时间 |
|------|------|----------|
| 项目初始化 | ✅ | 2026-03-12 |
| 依赖安装 | ✅ | 2026-03-12 |
| 核心功能开发 | ✅ | 2026-03-12 |
| UI 界面开发 | ✅ | 2026-03-12 |
| 开发服务器启动 | ✅ | 2026-03-12 |
| 架构文档编写 | ✅ | 2026-03-12 |

## 8. 待完成

- [ ] 生产构建优化
- [ ] 单元测试
- [ ] 错误边界处理
- [ ] 移动端适配

---

**文档创建时间**: 2026-03-12
**最后更新**: 2026-03-12
