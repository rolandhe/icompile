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
| `-onlySwagger` | bool | 否 | `false` | 仅生成 Swagger 文档（仅 Go） |

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

### -onlySwagger（仅 Swagger 文档）

仅生成 Swagger/OpenAPI 文档，不生成代码。此选项仅对 Go 语言有效。

```bash
./icomplie -onlySwagger -lang go ...
```

适用场景：
- 只需要 API 文档
- 在 CI/CD 中单独生成 Swagger JSON

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
  -onlySwagger
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

未提供任何参数时，程序会输出用法说明并退出。

```bash
# 需要至少提供 -i 和 -o 参数
./icomplie -i example/order/order.idl -o ./out -pp "myproject/api"
```

### 2. Go/Java 缺少包路径

Go 和 Java 语言需要提供 `-pp` 参数来指定包路径，否则生成的代码中 import/package 路径为空。

### 3. IDL 文件解析错误

如果 IDL 文件语法不正确，程序会输出解析错误信息（包含行号和列号），参见 [IDL 语法说明](idl-syntax.md)。

### 4. 导入的 IDL 文件找不到

确保 `go_import` 中的相对路径正确，相对于当前 IDL 文件所在目录。
