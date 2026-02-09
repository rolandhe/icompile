# 架构设计

## 整体架构

icomplie 采用经典的编译器架构，分为前端（解析）和后端（代码生成）两个主要阶段。

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   IDL 文件   │ ──▶ │   词法分析   │ ──▶ │   语法分析   │ ──▶ │  语义验证   │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                    │
                                                                    ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  输出文件    │ ◀── │  模板渲染   │ ◀── │  代码生成   │ ◀── │ Definition  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

## 编译流程

1. **词法分析 (Lexer)**: ANTLR4 生成的词法分析器将 IDL 文件转换为 Token 流
2. **语法分析 (Parser)**: ANTLR4 生成的语法分析器将 Token 流转换为 AST
3. **语义转换 (Transfer)**: 将 AST 转换为内部数据结构 `Definition`
4. **语义验证 (Semantic)**: 检查类型引用、重复定义、循环继承等
5. **代码生成 (CodeGen)**: 根据目标语言生成代码
6. **模板渲染 (Template)**: 使用 Go 模板引擎渲染最终代码

## 模块说明

### internal/parser

ANTLR4 生成的解析器代码，由 `antlr4/Service.g4` 语法文件生成。

| 文件 | 说明 |
|------|------|
| `service_lexer.go` | 词法分析器 |
| `service_parser.go` | 语法分析器 |
| `service_listener.go` | 监听器接口 |
| `service_base_listener.go` | 监听器基类 |

重新生成解析器：
```bash
cd antlr4
./antlr4.sh
```

### internal/transfer

将 ANTLR4 的 AST 转换为内部数据结构。

| 文件 | 说明 |
|------|------|
| `definition.go` | 核心数据结构定义 |
| `process.go` | AST 遍历和转换逻辑 |
| `parse_idl.go` | IDL 解析入口 |

核心数据结构：
```go
type Definition struct {
    Namespace       string
    GoStructImports map[string][]*ImportStructDefine
    Structs         []*StructDefine
    Services        []*ServiceDefine
}

type StructDefine struct {
    Name       string
    Fields     []*Field
    Extends    string
    Point      bool       // 指针嵌入（POINT 关键字）
    HasFormTag bool       // 生成 form tag（[form] 标记）
}

type ServiceDefine struct {
    Name    string
    RootUrl string
    Methods []*Method      // 统一方法列表，保持 IDL 中的定义顺序
    Posts   []*PostMethod
    Gets    []*GetMethod
    Puts    []*PutMethod
}

// Method 是统一的方法表示，保持 IDL 中的定义顺序
type Method struct {
    HTTPMethod HTTPMethod // GET, POST, PUT
    BaseMethod
    PostParams *PostParam
    GetParams  *GetParam
}
```

### pkg/semantic

语义验证模块，检查 IDL 定义的正确性。

验证项目：
- 重复的结构体定义
- 重复的服务定义
- 重复的方法名
- 未定义的类型引用
- 循环继承
- URL 路径冲突

### pkg/codegen

代码生成器模块，采用模板化设计。

Java 和 TypeScript 生成器实现了统一的 `Generator` 接口（`pkg/codegen/generator.go`）。Go 生成器目前使用独立的 `Main()` 函数入口（`pkg/codegen/golang/generator.go`），尚未迁移到 `Generator` 接口。

```
pkg/codegen/
├── generator.go          # Generator 接口定义（Java/TypeScript 实现）
├── golang/               # Go 代码生成器（通过 Main() 调用）
│   ├── generator.go      # 入口（Main 函数）
│   ├── generator_server.go
│   ├── generator_client.go
│   ├── generator_structs.go
│   ├── swagger.go        # Swagger/OpenAPI 2.0 生成
│   ├── client/           # 客户端渲染逻辑
│   └── template/         # 模板文件（go:embed）
├── java/                 # Java 代码生成器（实现 Generator 接口）
│   ├── generator.go
│   ├── pojo.go
│   ├── controller.go
│   ├── client/
│   └── template/
└── typescript/           # TypeScript 代码生成器（实现 Generator 接口）
    ├── generator.go
    ├── client/
    └── template/
```

### 模板系统

每个语言的代码生成器都使用独立的模板系统，模板通过 `go:embed` 嵌入到二进制文件中：

```
<lang>/template/
├── loader.go      # 模板加载器
├── embedded.go    # 嵌入式模板（go:embed）
├── registry.go    # 模板注册表
├── funcs.go       # 模板函数
└── templates/     # 模板文件目录
```

模板使用 Go 的 `text/template` 引擎。可以通过修改 `templates/` 目录下的模板文件来自定义生成的代码格式（需重新编译）。

## 扩展指南

### 添加新语言支持

1. 创建新的生成器目录：`pkg/codegen/<language>/`
2. 实现 `Generator` 接口：
   ```go
   type Generator interface {
       Name() string
       Generate(ctx *Context, def *transfer.Definition) (*Result, error)
   }
   ```
3. 创建模板系统：`template/` 目录
4. 创建模板文件：`templates/` 目录
5. 在 `cmd/root.go` 中添加语言分支调用

注意：当前 Go 生成器未遵循此接口模式，Java 和 TypeScript 生成器已实现。

### 模板变量示例：
```go
type ServiceClientRender struct {
    ServiceName string
    Methods     []*MethodRender
}

type MethodRender struct {
    MethodName      string
    HTTPMethod      string
    FullPath        string
    ParamsSignature string
    ReturnType      string
}
```

## 依赖关系

```
cmd/root.go
    ├── internal/transfer      # IDL 解析
    ├── pkg/semantic           # 语义验证
    ├── pkg/codegen/golang     # Go 生成器
    ├── pkg/codegen/java       # Java 生成器
    └── pkg/codegen/typescript # TypeScript 生成器

pkg/codegen/<lang>/
    ├── internal/transfer      # Definition 数据结构
    └── template/              # 模板系统

pkg/semantic/
    ├── internal/transfer      # Definition 数据结构
    └── pkg/errors             # 错误类型
```
