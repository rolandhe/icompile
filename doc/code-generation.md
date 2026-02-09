# 代码生成说明

本文档详细说明 icomplie 为各目标语言生成的代码结构和使用方法。

---

## Go 代码生成

### 生成命令

```bash
# 生成服务端代码
./icomplie -i example/order/order.idl -o ./out -pp "myproject/api" -lang go -target server

# 生成客户端代码
./icomplie -i example/order/order.idl -o ./out -pp "myproject/api" -lang go -target client

# 生成全部代码
./icomplie -i example/order/order.idl -o ./out -pp "myproject/api" -lang go -target all

# 仅生成结构体
./icomplie -i example/order/order.idl -o ./out -pp "myproject/api" -lang go -onlyStruct
```

### 生成的目录结构

```
<output_dir>/<namespace>/
├── structs/
│   └── <namespace>_structs.go    # 结构体定义
├── <namespace>_controller.go     # 服务接口 + Gin 路由绑定
├── <namespace>_controller_impl.go # 接口实现（仅首次生成）
├── <service>_client.go           # HTTP 客户端（-target client/all）
└── <namespace>_swagger.json      # Swagger 文档
```

### 结构体文件

生成的结构体包含 JSON tag 和验证 tag：

```go
package structs

type OrderRequest struct {
    UserId    int64  `json:"userId" binding:"required"`    // 用户ID
    ProductId string `json:"productId" binding:"required"` // 产品ID
    Quantity  int32  `json:"quantity,omitempty"`           // 数量
}

type OrderResponse struct {
    *common.BaseResponse                                   // 继承（指针嵌入）
    OrderId    int64  `json:"orderId" binding:"required"`  // 订单ID
    Status     string `json:"status" binding:"required"`   // 订单状态
}
```

### 服务端控制器

生成 Gin 框架的控制器代码：

```go
package order

// 服务接口
type OrderService interface {
    CreateOrder(bc *commons.BaseContext, req *structs.OrderRequest) *commons.Result[*structs.OrderResponse]
    GetOrder(bc *commons.BaseContext, req *structs.GetOrderRequest) *commons.Result[*structs.OrderResponse]
}

// 路由绑定函数
func BindOrderServiceController(e *gin.Engine) {
    g := e.Group("/api/v1")
    svc := OrderServiceControllerStd(&orderServiceControllerImpl{})

    requests.POST(g, &requests.RequestDesc[structs.OrderRequest, *commons.Result[*structs.OrderResponse]]{
        RelativePath: "order/create",
        BizCoreFunc:  svc.CreateOrder,
        LogLevel:     requests.LOG_LEVEL_ALL,
    })
    // ...
}
```

### 客户端代码

生成类型安全的 HTTP 客户端：

```go
package order

type OrderServiceClient struct {
    client *http.Client
    host   string
}

func NewOrderServiceClient(host string) *OrderServiceClient {
    return &OrderServiceClient{
        client: &http.Client{},
        host:   host,
    }
}

func (c *OrderServiceClient) CreateOrder(ctx context.Context, req *structs.OrderRequest) (*structs.OrderResponse, error) {
    // HTTP POST 实现
}

func (c *OrderServiceClient) GetOrder(ctx context.Context, orderId int64) (*structs.OrderResponse, error) {
    // HTTP GET 实现
}
```

### Swagger 文档

自动生成 OpenAPI 2.0 格式的 JSON 文档，包含：
- 所有结构体的 Schema 定义
- 所有 API 端点的路径和参数
- 请求/响应类型定义

---

## Java 代码生成

### 生成命令

```bash
# 生成服务端代码
./icomplie -i example/order/order.idl -o ./out -pp "com.example.order" -lang java -target server

# 生成客户端代码
./icomplie -i example/order/order.idl -o ./out -pp "com.example.order" -lang java -target client

# 生成全部代码
./icomplie -i example/order/order.idl -o ./out -pp "com.example.order" -lang java -target all
```

### 生成的目录结构

```
<output_dir>/
├── pojo/
│   └── <StructName>.java         # POJO 类
├── controller/
│   └── <ServiceName>Controller.java  # Spring MVC 控制器
└── client/
    ├── HttpClient.java           # HTTP 客户端接口
    ├── ApacheHttpClient.java     # Apache HttpComponents 实现
    └── <ServiceName>Client.java  # 服务客户端
```

### POJO 类

生成带 Lombok 注解的 Java 类：

```java
package com.example.order.pojo;

import lombok.Data;

/**
 * 订单请求
 */
@Data
public class OrderRequest {
    /** 用户ID */
    private Long userId;

    /** 产品ID */
    private String productId;

    /** 数量 */
    private Integer quantity;
}
```

继承关系：

```java
@Data
public class OrderResponse extends BaseResponse {
    /** 订单ID */
    private Long orderId;

    /** 订单状态 */
    private String status;
}
```

### Spring MVC 控制器

```java
package com.example.order.controller;

import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1")
public class OrderServiceController {

    /**
     * 创建订单
     */
    @PostMapping("order/create")
    public Result<OrderResponse> createOrder(@RequestBody OrderRequest req) {
        // TODO: 实现业务逻辑
        return null;
    }

    /**
     * 获取订单
     */
    @GetMapping("order/get")
    public Result<OrderResponse> getOrder(@RequestParam("orderId") Long orderId) {
        // TODO: 实现业务逻辑
        return null;
    }
}
```

### HTTP 客户端

基于 Apache HttpComponents 的客户端实现：

```java
// 接口定义
public interface HttpClient {
    <T> T doPost(String url, Object body, Class<T> responseType) throws Exception;
    <T> T doGet(String url, Class<T> responseType) throws Exception;
    String buildUrl(String basePath, Map<String, Object> params);
}

// 使用示例
HttpClient httpClient = new ApacheHttpClient("https://api.example.com");
OrderServiceClient orderClient = new OrderServiceClient(httpClient);

OrderResponse response = orderClient.createOrder(request);
```

### Maven 依赖

```xml
<dependencies>
    <!-- Lombok -->
    <dependency>
        <groupId>org.projectlombok</groupId>
        <artifactId>lombok</artifactId>
        <version>1.18.30</version>
        <scope>provided</scope>
    </dependency>

    <!-- Spring Web (服务端) -->
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-web</artifactId>
    </dependency>

    <!-- Apache HttpComponents (客户端) -->
    <dependency>
        <groupId>org.apache.httpcomponents.client5</groupId>
        <artifactId>httpclient5</artifactId>
        <version>5.2.1</version>
    </dependency>

    <!-- Jackson -->
    <dependency>
        <groupId>com.fasterxml.jackson.core</groupId>
        <artifactId>jackson-databind</artifactId>
        <version>2.15.2</version>
    </dependency>
</dependencies>
```

---

## TypeScript 代码生成

TypeScript 仅支持客户端代码生成，提供两种平台：浏览器端（axios）和微信小程序。

### 生成命令

```bash
# 生成浏览器端客户端（基于 axios）
./icomplie -i example/order/order.idl -o ./out -lang typescript -platform browser

# 生成微信小程序客户端
./icomplie -i example/order/order.idl -o ./out -lang typescript -platform miniapp

# 简写形式
./icomplie -i example/order/order.idl -o ./out -lang ts
```

### 生成的目录结构

```
<output_dir>/<namespace>/
├── types.ts              # 类型定义
├── httpClient.ts         # HTTP 客户端接口
├── axiosClient.ts        # Axios 实现（浏览器端）
├── wxClient.ts           # wx.request 实现（小程序）
└── <service>Client.ts    # 服务客户端
```

### 类型定义

```typescript
// types.ts

export interface OrderRequest {
  userId: number;
  productId: string;
  quantity?: number;
}

export interface OrderResponse extends BaseResponse {
  orderId: number;
  status: string;
  createTime: string;
  items?: OrderItem[];
}

export interface OrderItem {
  productId: string;
  quantity: number;
  price: number;
}
```

### HTTP 客户端接口

```typescript
// httpClient.ts

export interface HttpClient {
  doPost<T>(url: string, body?: any, headers?: Record<string, string>): Promise<T>;
  doGet<T>(url: string, headers?: Record<string, string>): Promise<T>;
  doPut<T>(url: string, body?: any, headers?: Record<string, string>): Promise<T>;
  buildUrl(basePath: string, params: Record<string, string | number>): string;
}
```

### 浏览器端实现（Axios）

```typescript
// axiosClient.ts

import axios, { AxiosInstance } from 'axios';
import { HttpClient } from './httpClient';

export class AxiosHttpClient implements HttpClient {
  private instance: AxiosInstance;

  constructor(baseURL: string, timeout: number = 30000) {
    this.instance = axios.create({ baseURL, timeout });
  }

  async doPost<T>(url: string, body?: any, headers?: Record<string, string>): Promise<T> {
    const response = await this.instance.post<T>(url, body, { headers });
    return response.data;
  }

  async doGet<T>(url: string, headers?: Record<string, string>): Promise<T> {
    const response = await this.instance.get<T>(url, { headers });
    return response.data;
  }

  // ...
}
```

### 微信小程序实现

```typescript
// wxClient.ts

import { HttpClient } from './httpClient';

export interface WxRequestConfig {
  baseURL: string;
  timeout?: number;
  header?: Record<string, string>;
}

export class WxHttpClient implements HttpClient {
  private config: WxRequestConfig;

  constructor(config: WxRequestConfig) {
    this.config = config;
  }

  async doPost<T>(url: string, body?: any, headers?: Record<string, string>): Promise<T> {
    return new Promise((resolve, reject) => {
      wx.request({
        url: this.config.baseURL + url,
        method: 'POST',
        data: body,
        header: { 'Content-Type': 'application/json', ...this.config.header, ...headers },
        success: (res) => resolve(res.data as T),
        fail: (err) => reject(err),
      });
    });
  }

  // ...
}
```

### 服务客户端

```typescript
// orderServiceClient.ts

import { HttpClient } from './httpClient';
import { OrderRequest, OrderResponse } from './types';

export class OrderServiceClient {
  private client: HttpClient;

  constructor(client: HttpClient) {
    this.client = client;
  }

  async createOrder(req: OrderRequest): Promise<OrderResponse> {
    return this.client.doPost<OrderResponse>('/api/v1/order/create', req);
  }

  async getOrder(orderId: number): Promise<OrderResponse> {
    const url = this.client.buildUrl('/api/v1/order/get', { orderId });
    return this.client.doGet<OrderResponse>(url);
  }
}
```

### 使用示例

**浏览器端：**

```typescript
import { AxiosHttpClient } from './axiosClient';
import { OrderServiceClient } from './orderServiceClient';

const httpClient = new AxiosHttpClient('https://api.example.com');
const orderClient = new OrderServiceClient(httpClient);

// 调用 API
const response = await orderClient.createOrder({
  userId: 12345,
  productId: 'PROD001',
  quantity: 2,
});
```

**微信小程序：**

```typescript
import { WxHttpClient } from './wxClient';
import { OrderServiceClient } from './orderServiceClient';

const httpClient = new WxHttpClient({ baseURL: 'https://api.example.com' });
const orderClient = new OrderServiceClient(httpClient);

// 调用 API
const response = await orderClient.createOrder({
  userId: 12345,
  productId: 'PROD001',
  quantity: 2,
});
```

### npm 依赖

**浏览器端 package.json：**

```json
{
  "dependencies": {
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}
```

**微信小程序：**

```bash
npm install --save-dev miniprogram-api-typings
```

---

## 模板自定义

icomplie 支持自定义代码生成模板。

### 使用自定义模板

```bash
./icomplie -i input.idl -o ./out -pp "myproject" -tpl ./my-templates
```

### 模板目录结构

```
my-templates/
├── go/
│   ├── server/
│   │   ├── struct.tmpl
│   │   ├── interface.tmpl
│   │   └── ...
│   └── client/
│       └── service_client.tmpl
├── java/
│   ├── server/
│   │   ├── pojo.tmpl
│   │   └── controller.tmpl
│   └── client/
│       └── service_client.tmpl
└── typescript/
    ├── browser/
    │   └── ...
    └── miniapp/
        └── ...
```

### 模板语法

模板使用 Go 的 `text/template` 语法。可用的模板函数包括：

| 函数 | 说明 | 示例 |
|------|------|------|
| `formatVariable` | 格式化变量名 | `{{formatVariable .Name}}` |
| `capitalize` | 首字母大写 | `{{capitalize .Name}}` |
| `lower` | 转小写 | `{{lower .Name}}` |
| `upper` | 转大写 | `{{upper .Name}}` |
| `join` | 连接字符串 | `{{join .Items ","}}` |
