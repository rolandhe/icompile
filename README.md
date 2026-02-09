# icomplie

icomplie 是一个 IDL（接口定义语言）编译器，可以从 `.idl` 文件生成多种语言的代码。

## 特性

- **多语言支持**: Go、Java、TypeScript
- **服务端代码生成**: 生成 HTTP 服务控制器和接口
- **客户端代码生成**: 生成类型安全的 HTTP 客户端
- **多平台支持**: TypeScript 支持浏览器端（axios）和微信小程序（wx.request）
- **Swagger 文档**: 自动生成 OpenAPI 2.0 文档（Go）

## 快速开始

### 安装

```bash
go build -o icomplie .
```

### 基本用法

```bash
# 生成 Go 服务端代码
./icomplie -i example/order/order.idl -o ./out -pp "myproject/api" -lang go -target server

# 生成 Go 客户端代码
./icomplie -i example/order/order.idl -o ./out -pp "myproject/api" -lang go -target client

# 生成 Java 代码
./icomplie -i example/order/order.idl -o ./out -pp "com.example" -lang java -target all

# 生成 TypeScript 浏览器客户端
./icomplie -i example/order/order.idl -o ./out -lang typescript -platform browser

# 生成 TypeScript 微信小程序客户端
./icomplie -i example/order/order.idl -o ./out -lang typescript -platform miniapp
```

## 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-i` | 输入 IDL 文件路径 | 必填 |
| `-o` | 输出目录路径 | 必填 |
| `-pp` | 包路径（Go/Java 需要） | - |
| `-lang` | 目标语言: `go`, `java`, `typescript`/`ts` | `go` |
| `-target` | 生成目标: `server`, `client`, `all` | `server` |
| `-platform` | 平台（仅 TypeScript）: `browser`, `miniapp` | `browser` |
| `-onlyStruct` | 仅生成结构体定义 | `false` |

## 支持的语言

| 语言 | 服务端 | 客户端 | 说明 |
|------|--------|--------|------|
| Go | ✅ | ✅ | Gin 框架，标准 HTTP 客户端 |
| Java | ✅ | ✅ | Spring MVC，Apache HttpComponents |
| TypeScript | - | ✅ | axios（浏览器）/ wx.request（小程序） |

## 文档

- [架构设计](doc/architecture.md) - 系统架构和模块说明
- [IDL 语法](doc/idl-syntax.md) - IDL 文件语法说明
- [代码生成](doc/code-generation.md) - 各语言代码生成详情
- [命令行参考](doc/cli-reference.md) - 完整的命令行参数说明
- [使用示例](doc/examples.md) - 完整的使用示例

## IDL 示例

```idl
namespace go order

struct OrderRequest {
    required i64 userId(remark="用户ID"),
    required string productId(remark="产品ID"),
    optional i32 quantity(remark="数量"),
}

struct OrderResponse {
    required i64 orderId(remark="订单ID"),
    required string status(remark="订单状态"),
}

service OrderService URL="/api/v1" {
    ("创建订单")
    POST URL="order/create" struct OrderResponse createOrder(struct OrderRequest req),

    ("查询订单")
    GET URL="order/get" struct OrderResponse getOrder(required i64 orderId),
}
```

## 目录结构

```
icomplie/
├── antlr4/              # ANTLR4 语法文件
├── cmd/                 # 命令行入口
├── common/              # 公共工具函数
├── doc/                 # 文档
├── example/             # 示例 IDL 文件
├── internal/
│   ├── parser/          # ANTLR4 生成的解析器
│   └── transfer/        # IDL 解析和转换
├── pkg/
│   ├── codegen/         # 代码生成器
│   │   ├── golang/      # Go 代码生成
│   │   ├── java/        # Java 代码生成
│   │   └── typescript/  # TypeScript 代码生成
│   ├── errors/          # 错误定义
│   ├── semantic/        # 语义验证
│   └── types/           # 类型注册表
├── scripts/             # 构建脚本
├── main.go              # 程序入口
└── README.md
```

## 许可证

MIT License
