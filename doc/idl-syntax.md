# IDL 语法说明

本文档详细说明 icomplie 支持的 IDL（接口定义语言）语法。

## 文件结构

一个 IDL 文件由以下部分组成：

```
[头部声明]
[定义部分]
```

## 头部声明

### namespace

声明命名空间，决定生成代码的包名。

```idl
namespace go <package_name>
```

示例：
```idl
namespace go order
```

### go_import

导入其他 IDL 文件中定义的类型。

```idl
go_import [alias] "<go_package_path>" "<relative_idl_path>"
```

参数说明：
- `alias`（可选）: 导入别名
- `go_package_path`: Go 包的完整路径
- `relative_idl_path`: 相对于当前文件的 IDL 文件路径

示例：
```idl
go_import "go.example.com/project/order/share" "./share.idl"
go_import "go.example.com/project/stable" "../stable/pager.idl"
```

---

## 数据类型

### 基本类型

| IDL 类型 | Go 类型 | Java 类型 | TypeScript 类型 | 说明 |
|---------|---------|-----------|-----------------|------|
| `bool` | `bool` | `Boolean` | `boolean` | 布尔值 |
| `byte` | `byte` | `Byte` | `number` | 字节 |
| `i8` | `int8` | `Byte` | `number` | 8位整数 |
| `i16` | `int16` | `Short` | `number` | 16位整数 |
| `i32` | `int32` | `Integer` | `number` | 32位整数 |
| `i64` | `int64` | `Long` | `number` | 64位整数 |
| `float` | `float32` | `Float` | `number` | 单精度浮点 |
| `double` | `float64` | `Double` | `number` | 双精度浮点 |
| `string` | `string` | `String` | `string` | 字符串 |

### 复合类型

#### list

列表类型，包含同一类型的多个元素。

```idl
list<T>
```

示例：
```idl
list<i64>                    // 整数列表
list<string>                 // 字符串列表
list<struct Traveller>       // 结构体列表
```

#### map

映射类型，键值对集合。

```idl
map<K, V>
```

键类型限制：`byte`, `i16`, `i32`, `i64`, `double`, `string`

示例：
```idl
map<string, i64>             // 字符串到整数的映射
map<i64, struct User>        // 整数到结构体的映射
```

#### struct

引用自定义结构体类型。

```idl
struct <StructName>
```

示例：
```idl
struct OrderRequest          // 当前命名空间的结构体
struct order_share.PayInfo   // 其他命名空间的结构体
```

---

## 结构体定义

### 基本语法

```idl
struct <Name> {
    [required|optional] <type> <field_name>(annotations...),
}
```

示例：
```idl
struct OrderRequest {
    required i64 userId(remark="用户ID"),
    required string productId(remark="产品ID"),
    optional i32 quantity(remark="数量"),
}
```

### 字段修饰符

| 修饰符 | 说明 |
|--------|------|
| `required` | 必填字段 |
| `optional` | 可选字段 |
| （无） | 默认为可选 |

### 结构体继承

使用 `extends` 关键字继承另一个结构体。

```idl
struct <Name> extends [POINT] struct <ParentName> {
    // 新增字段
}
```

- `POINT`: 可选关键字，表示使用指针嵌入（Go 中生成 `*ParentStruct`）

示例：
```idl
// 值嵌入
struct OrderResponse extends struct BaseResponse {
    required i64 orderId(remark="订单ID"),
}

// 指针嵌入
struct AgentOrderResp extends POINT struct CustomOrderResp {
    required string modifyTime(remark="修改时间"),
}
```

### Form 提示

使用 `[form]` 标记结构体用于表单绑定（影响 Go 的 binding tag）。

```idl
struct QueryParams [form] {
    required string keyword(remark="搜索关键词"),
}
```

---

## 字段注解

字段注解用于添加元数据，影响代码生成。

### 语法

```idl
<type> <field_name>(key1="value1", key2="value2")
```

### 常用注解

| 注解 | 说明 | 示例 |
|------|------|------|
| `remark` | 字段说明，生成注释和 Swagger 文档 | `remark="用户ID"` |
| `binding` | Go 验证规则（gin binding tag） | `binding="required,min=1"` |
| `json` | 自定义 JSON 字段名 | `json="user_id"` |

### 示例

```idl
struct User {
    required i64 id(remark="用户ID"),
    required string email(binding="required,email", remark="邮箱地址"),
    required i16 ageGroup(binding="required,oneof=1 2 3", remark="年龄组"),
    optional string nickname(remark="昵称"),
}
```

---

## 服务定义

### 基本语法

```idl
service <ServiceName> URL="<base_path>" {
    // 方法定义
}
```

示例：
```idl
service OrderService URL="/api/v1" {
    // 方法定义
}
```

### 方法定义

#### POST 方法

```idl
("<description>")
POST URL="<path>" <return_type> <method_name>(<params>),
```

参数类型：
- `struct <StructName> <param_name>` - 结构体参数（请求体）
- `list<struct <StructName>> <param_name>` - 结构体列表参数

#### GET 方法

```idl
("<description>")
GET URL="<path>" <return_type> <method_name>(<params>),
```

参数类型：
- `struct <StructName> <param_name>` - 结构体参数（查询参数）
- 简单参数列表：`<type> <name>, <type> <name>, ...`

#### PUT 方法

```idl
("<description>")
PUT URL="<path>" <return_type> <method_name>(<params>),
```

### 返回类型

| 类型 | 说明 |
|------|------|
| `void` | 无返回值 |
| `<base_type>` | 基本类型 |
| `struct <Name>` | 结构体 |
| `list<struct <Name>>` | 结构体列表 |
| `struct <Name> PAGEABLE` | 分页结构体（生成分页响应） |

### 方法修饰符

| 修饰符 | 说明 |
|--------|------|
| `not_login` | 不需要登录验证 |
| `PAGEABLE` | 分页返回（与返回类型配合使用） |

### 完整示例

```idl
service OrderService URL="/api/v1" {
    ("创建订单")
    POST URL="order/create" i64 createOrder(struct OrderRequest req),

    ("查询订单列表")
    POST URL="order/list" struct OrderResponse PAGEABLE listOrders(struct QueryRequest req),

    ("获取订单详情")
    GET URL="order/detail" struct OrderResponse getOrder(required i64 orderId(remark="订单ID")),

    ("删除订单")
    POST URL="order/delete" void deleteOrder(struct IdRequest req),

    ("更新订单")
    PUT URL="order/update" struct OrderResponse updateOrder(struct OrderRequest req),

    ("公开接口-不需要登录")
    GET URL="order/public" struct OrderResponse publicApi(required string code) not_login,
}
```

---

## 注释

IDL 支持三种注释格式：

```idl
// 单行注释

# 单行注释（Shell 风格）

/*
 * 多行注释
 */
```

---

## 完整示例

```idl
namespace go order

go_import "go.example.com/project/common" "../common/common.idl"

// 订单请求结构体
struct OrderRequest {
    required i64 userId(remark="用户ID"),
    required string productId(remark="产品ID"),
    optional i32 quantity(remark="数量", binding="omitempty,min=1"),
    optional map<string, string> extra(remark="扩展信息"),
}

// 订单响应结构体
struct OrderResponse extends POINT struct common.BaseResponse {
    required i64 orderId(remark="订单ID"),
    required string status(remark="订单状态"),
    required string createTime(remark="创建时间"),
    list<struct OrderItem> items(remark="订单项"),
}

// 订单项
struct OrderItem {
    required string productId(remark="产品ID"),
    required i32 quantity(remark="数量"),
    required i64 price(remark="单价，单位分"),
}

// 订单服务
service OrderService URL="/api/v1" {
    ("创建订单")
    POST URL="order/create" struct OrderResponse createOrder(struct OrderRequest req),

    ("查询订单")
    GET URL="order/get" struct OrderResponse getOrder(required i64 orderId(remark="订单ID")),

    ("订单列表")
    POST URL="order/list" struct OrderResponse PAGEABLE listOrders(struct common.PageRequest req),

    ("删除订单")
    POST URL="order/delete" void deleteOrder(struct common.IdRequest req),
}
```
