# 命令行参考

本文档详细说明 icomplie 的命令行参数和使用方法。

## 基本用法

```bash
./icomplie [options]
```

## 参数说明

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `-i` | string | 是 | - | 输入 IDL 文件路径 |
| `-o` | string | 是 | - | 输出目录路径 |
| `-pp` | string | 条件 | - | 包路径前缀（Go/Java 必填） |
| `-lang` | string | 否 | `go` | 目标语言 |
| `-target` | string | 否 | `server` | 生成目标 |
| `-platform` | string | 否 | `browser` | 平台（仅 TypeScript） |
| `-onlyStruct` | bool | 否 | `false` | 仅生成结构体定义 |
| `-tpl` | string | 否 | - | 自定义模板目录 |

---

## 参数详解

### -i（输入文件）

指定输入的 IDL 文件路径。支持相对路径和绝对路径。

```bash
./icomplie -i ./example/order/order.idl ...
./icomplie -i /path/to/service.idl ...
```

### -o（输出目录）

指定生成代码的输出目录。如果目录不存在会自动创建。

```bash
./icomplie -o ./generated ...
./icomplie -o /path/to/output ...
```

### -pp（包路径）

指定生成代码的包路径前缀。

- **Go**: 用于生成 import 语句中的包路径
- **Java**: 用于生成 package 声明

```bash
# Go
./icomplie -pp "github.com/myproject/api" -lang go ...

# Java
./icomplie -pp "com.example.api" -lang java ...
```

**注意**: TypeScript 不需要此参数。

### -lang（目标语言）

指定生成代码的目标语言。

| 值 | 说明 |
|-----|------|
| `go` | Go 语言（默认） |
| `java` | Java 语言 |
| `typescript` | TypeScript |
| `ts` | TypeScript（简写） |

```bash
./icomplie -lang go ...
./icomplie -lang java ...
./icomplie -lang typescript ...
./icomplie -lang ts ...
```

### -target（生成目标）

指定生成服务端代码、客户端代码或全部。

| 值 | 说明 |
|-----|------|
| `server` | 仅生成服务端代码（默认） |
| `client` | 仅生成客户端代码 |
| `all` | 生成服务端和客户端代码 |

```bash
./icomplie -target server ...
./icomplie -target client ...
./icomplie -target all ...
```

**各语言支持情况：**

| 语言 | server | client |
|------|--------|--------|
| Go | ✅ | ✅ |
| Java | ✅ | ✅ |
| TypeScript | ❌ | ✅ |

### -platform（平台）

仅用于 TypeScript，指定目标平台。

| 值 | 说明 |
|-----|------|
| `browser` | 浏览器端，使用 axios（默认） |
| `miniapp` | 微信小程序，使用 wx.request |

```bash
./icomplie -lang typescript -platform browser ...
./icomplie -lang typescript -platform miniapp ...
```

### -onlyStruct（仅结构体）

仅生成结构体/类型定义，不生成服务代码。

```bash
./icomplie -onlyStruct ...
```

适用场景：
- 只需要数据模型定义
- 在多个项目间共享类型定义
- 不需要 HTTP 服务代码

### -tpl（自定义模板）

指定自定义模板目录，覆盖默认的嵌入式模板。

```bash
./icomplie -tpl ./my-templates ...
```

模板目录结构参见 [代码生成说明](code-generation.md#模板自定义)。

---

## 使用示例

### Go 服务端

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -pp "myproject/api" \
  -lang go \
  -target server
```

### Go 客户端

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -pp "myproject/api" \
  -lang go \
  -target client
```

### Go 全部代码

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -pp "myproject/api" \
  -lang go \
  -target all
```

### Go 仅结构体

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -pp "myproject/api" \
  -lang go \
  -onlyStruct
```

### Java 服务端

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -pp "com.example.order" \
  -lang java \
  -target server
```

### Java 客户端

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -pp "com.example.order" \
  -lang java \
  -target client
```

### Java 全部代码

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -pp "com.example.order" \
  -lang java \
  -target all
```

### TypeScript 浏览器端

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -lang typescript \
  -platform browser
```

### TypeScript 微信小程序

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -lang typescript \
  -platform miniapp
```

### 使用自定义模板

```bash
./icomplie \
  -i example/order/order.idl \
  -o ./out \
  -pp "myproject/api" \
  -lang go \
  -tpl ./custom-templates
```

---

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 参数错误或执行失败 |

---

## 常见问题

### 1. 缺少必填参数

```
Error: -i (input file) is required
Error: -o (output directory) is required
```

解决：确保提供了 `-i` 和 `-o` 参数。

### 2. Go/Java 缺少包路径

```
Error: -pp (package path) is required for Go/Java
```

解决：Go 和 Java 需要提供 `-pp` 参数。

### 3. IDL 文件解析错误

```
Error: failed to parse IDL file: ...
```

解决：检查 IDL 文件语法是否正确，参见 [IDL 语法说明](idl-syntax.md)。

### 4. 导入的 IDL 文件找不到

```
Error: cannot find imported file: ./share.idl
```

解决：确保 `go_import` 中的相对路径正确，相对于当前 IDL 文件。

### 5. TypeScript 不支持服务端

```
Error: TypeScript only supports client generation
```

解决：TypeScript 仅支持 `-target client`，不支持 `server`。
