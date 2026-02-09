# 使用示例

本文档提供 icomplie 的完整使用示例。

---

## 示例 IDL 文件

### 基础结构体定义

```idl
// common.idl
namespace go common

struct IdRequest {
    required i64 id(remark="ID"),
}

struct PageRequest {
    required i32 page(remark="页码", binding="required,min=1"),
    required i32 pageSize(remark="每页数量", binding="required,min=1,max=100"),
}

struct BaseResponse {
    required i64 id(remark="ID"),
    required string createTime(remark="创建时间"),
    required string updateTime(remark="更新时间"),
}
```

### 业务服务定义

```idl
// user.idl
namespace go user

go_import "myproject/common" "../common/common.idl"

struct User {
    i64 id(remark="用户ID"),
    required string username(remark="用户名", binding="required,min=3,max=50"),
    required string email(remark="邮箱", binding="required,email"),
    string phone(remark="手机号"),
    required i16 status(remark="状态: 1-正常 2-禁用", binding="required,oneof=1 2"),
}

struct UserResponse extends POINT struct common.BaseResponse {
    required string username(remark="用户名"),
    required string email(remark="邮箱"),
    string phone(remark="手机号"),
    required i16 status(remark="状态"),
}

struct UserListRequest extends POINT struct common.PageRequest {
    string keyword(remark="搜索关键词"),
    i16 status(remark="状态筛选"),
}

struct CreateUserRequest {
    required string username(remark="用户名", binding="required,min=3,max=50"),
    required string email(remark="邮箱", binding="required,email"),
    string phone(remark="手机号"),
    required string password(remark="密码", binding="required,min=6"),
}

struct UpdateUserRequest {
    required i64 id(remark="用户ID", binding="required"),
    string username(remark="用户名", binding="omitempty,min=3,max=50"),
    string email(remark="邮箱", binding="omitempty,email"),
    string phone(remark="手机号"),
}

service UserService URL="/api/v1" {
    ("创建用户")
    POST URL="user/create" struct UserResponse createUser(struct CreateUserRequest req),

    ("更新用户")
    PUT URL="user/update" struct UserResponse updateUser(struct UpdateUserRequest req),

    ("获取用户详情")
    GET URL="user/detail" struct UserResponse getUser(required i64 id(remark="用户ID")),

    ("用户列表")
    POST URL="user/list" struct UserResponse PAGEABLE listUsers(struct UserListRequest req),

    ("删除用户")
    POST URL="user/delete" void deleteUser(struct common.IdRequest req),
}
```

---

## Go 代码生成示例

### 生成命令

```bash
# 生成服务端和客户端
./icomplie -i user.idl -o ./generated -pp "myproject/api" -lang go -target all
```

### 生成的文件

```
generated/user/
├── structs/
│   └── user_structs.go
├── user_controller.go
├── user_controller_impl.go
├── user_service_client.go
└── user_swagger.json
```

### 使用生成的服务端代码

```go
package main

import (
    "github.com/gin-gonic/gin"
    "myproject/api/user"
)

func main() {
    e := gin.Default()

    // 绑定用户服务路由
    user.BindUserServiceController(e)

    e.Run(":8080")
}
```

实现业务逻辑（编辑 `user_controller_impl.go`）：

```go
package user

import (
    "myproject/api/user/structs"
    "myproject/commons"
)

type userServiceControllerImpl struct{}

func (s *userServiceControllerImpl) CreateUser(bc *commons.BaseContext, req *structs.CreateUserRequest) *commons.Result[*structs.UserResponse] {
    // 实现创建用户逻辑
    user := &structs.UserResponse{
        BaseResponse: &commons.BaseResponse{
            Id:         1,
            CreateTime: "2024-01-01 00:00:00",
            UpdateTime: "2024-01-01 00:00:00",
        },
        Username: req.Username,
        Email:    req.Email,
        Phone:    req.Phone,
        Status:   1,
    }
    return commons.Success(user)
}

func (s *userServiceControllerImpl) GetUser(bc *commons.BaseContext, req *structs.GetUserRequest) *commons.Result[*structs.UserResponse] {
    // 实现获取用户逻辑
    // ...
}

// 其他方法实现...
```

### 使用生成的客户端代码

```go
package main

import (
    "context"
    "fmt"
    "myproject/api/user"
    "myproject/api/user/structs"
)

func main() {
    // 创建客户端
    client := user.NewUserServiceClient("http://localhost:8080")

    // 创建用户
    resp, err := client.CreateUser(context.Background(), &structs.CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "123456",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created user: %+v\n", resp)

    // 获取用户
    user, err := client.GetUser(context.Background(), 1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("User: %+v\n", user)
}
```

---

## Java 代码生成示例

### 生成命令

```bash
# 生成服务端和客户端
./icomplie -i user.idl -o ./generated -pp "com.example.user" -lang java -target all
```

### 生成的文件

```
generated/
├── pojo/
│   ├── User.java
│   ├── UserResponse.java
│   ├── UserListRequest.java
│   ├── CreateUserRequest.java
│   └── UpdateUserRequest.java
├── controller/
│   └── UserServiceController.java
└── client/
    ├── HttpClient.java
    ├── ApacheHttpClient.java
    └── UserServiceClient.java
```

### 使用生成的服务端代码

```java
// UserServiceController.java 已生成，添加业务逻辑

package com.example.user.controller;

import com.example.user.pojo.*;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1")
public class UserServiceController {

    @PostMapping("user/create")
    public Result<UserResponse> createUser(@RequestBody CreateUserRequest req) {
        // 实现创建用户逻辑
        UserResponse response = new UserResponse();
        response.setId(1L);
        response.setUsername(req.getUsername());
        response.setEmail(req.getEmail());
        response.setStatus((short) 1);
        return Result.success(response);
    }

    @GetMapping("user/detail")
    public Result<UserResponse> getUser(@RequestParam("id") Long id) {
        // 实现获取用户逻辑
        // ...
    }

    // 其他方法...
}
```

### 使用生成的客户端代码

```java
package com.example;

import com.example.user.client.*;
import com.example.user.pojo.*;

public class Main {
    public static void main(String[] args) throws Exception {
        // 创建 HTTP 客户端
        HttpClient httpClient = new ApacheHttpClient("http://localhost:8080");

        // 创建服务客户端
        UserServiceClient userClient = new UserServiceClient(httpClient);

        // 创建用户
        CreateUserRequest createReq = new CreateUserRequest();
        createReq.setUsername("testuser");
        createReq.setEmail("test@example.com");
        createReq.setPassword("123456");

        UserResponse response = userClient.createUser(createReq);
        System.out.println("Created user: " + response.getUsername());

        // 获取用户
        UserResponse user = userClient.getUser(1L);
        System.out.println("User: " + user.getUsername());
    }
}
```

---

## TypeScript 代码生成示例

### 浏览器端（axios）

#### 生成命令

```bash
./icomplie -i user.idl -o ./generated -lang typescript -platform browser
```

#### 生成的文件

```
generated/user/
├── types.ts
├── httpClient.ts
├── axiosClient.ts
└── userServiceClient.ts
```

#### 使用示例

```typescript
// main.ts
import { AxiosHttpClient } from './generated/user/axiosClient';
import { UserServiceClient } from './generated/user/userServiceClient';
import { CreateUserRequest } from './generated/user/types';

async function main() {
  // 创建 HTTP 客户端
  const httpClient = new AxiosHttpClient('http://localhost:8080');

  // 创建服务客户端
  const userClient = new UserServiceClient(httpClient);

  // 创建用户
  const createReq: CreateUserRequest = {
    username: 'testuser',
    email: 'test@example.com',
    password: '123456',
  };

  const response = await userClient.createUser(createReq);
  console.log('Created user:', response.username);

  // 获取用户
  const user = await userClient.getUser(1);
  console.log('User:', user.username);

  // 用户列表
  const listResp = await userClient.listUsers({ page: 1, pageSize: 10 });
  console.log('Users:', listResp);
}

main().catch(console.error);
```

#### package.json

```json
{
  "name": "user-client",
  "version": "1.0.0",
  "dependencies": {
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "@types/node": "^20.0.0"
  }
}
```

### 微信小程序

#### 生成命令

```bash
./icomplie -i user.idl -o ./generated -lang typescript -platform miniapp
```

#### 生成的文件

```
generated/user/
├── types.ts
├── httpClient.ts
├── wxClient.ts
└── userServiceClient.ts
```

#### 使用示例

```typescript
// pages/user/user.ts
import { WxHttpClient } from '../../generated/user/wxClient';
import { UserServiceClient } from '../../generated/user/userServiceClient';

Page({
  data: {
    user: null as any,
  },

  onLoad() {
    this.loadUser();
  },

  async loadUser() {
    const httpClient = new WxHttpClient({
      baseURL: 'https://api.example.com',
      timeout: 30000,
      header: {
        'Authorization': 'Bearer xxx',
      },
    });

    const userClient = new UserServiceClient(httpClient);

    try {
      const user = await userClient.getUser(1);
      this.setData({ user });
    } catch (error) {
      console.error('Failed to load user:', error);
      wx.showToast({ title: '加载失败', icon: 'error' });
    }
  },

  async createUser() {
    const httpClient = new WxHttpClient({
      baseURL: 'https://api.example.com',
    });

    const userClient = new UserServiceClient(httpClient);

    try {
      const response = await userClient.createUser({
        username: 'newuser',
        email: 'new@example.com',
        password: '123456',
      });

      wx.showToast({ title: '创建成功', icon: 'success' });
      console.log('Created user:', response);
    } catch (error) {
      console.error('Failed to create user:', error);
      wx.showToast({ title: '创建失败', icon: 'error' });
    }
  },
});
```

#### 小程序配置

```json
// project.config.json
{
  "setting": {
    "useCompilerPlugins": ["typescript"]
  }
}
```

安装类型定义：

```bash
npm install --save-dev miniprogram-api-typings
```

---

## 复杂类型示例

### 嵌套结构体

```idl
struct Address {
    required string country(remark="国家"),
    required string city(remark="城市"),
    required string street(remark="街道"),
    string zipCode(remark="邮编"),
}

struct Company {
    required string name(remark="公司名称"),
    required struct Address address(remark="公司地址"),
    list<struct Employee> employees(remark="员工列表"),
}

struct Employee {
    required string name(remark="姓名"),
    required string position(remark="职位"),
    struct Address homeAddress(remark="家庭地址"),
}
```

### Map 类型

```idl
struct Config {
    required map<string, string> settings(remark="配置项"),
    map<string, i64> counters(remark="计数器"),
    map<i64, struct User> userCache(remark="用户缓存"),
}
```

### 继承链

```idl
struct BaseEntity {
    i64 id(remark="ID"),
    string createTime(remark="创建时间"),
    string updateTime(remark="更新时间"),
}

struct AuditEntity extends POINT struct BaseEntity {
    i64 createBy(remark="创建人ID"),
    i64 updateBy(remark="更新人ID"),
}

struct Order extends POINT struct AuditEntity {
    required string orderNo(remark="订单号"),
    required i64 amount(remark="金额"),
    required i16 status(remark="状态"),
}
```

---

## 分页查询示例

### IDL 定义

```idl
struct PageRequest {
    required i32 page(remark="页码", binding="required,min=1"),
    required i32 pageSize(remark="每页数量", binding="required,min=1,max=100"),
}

struct OrderQueryRequest extends POINT struct PageRequest {
    string orderNo(remark="订单号"),
    i16 status(remark="状态"),
    string startDate(remark="开始日期"),
    string endDate(remark="结束日期"),
}

service OrderService URL="/api/v1" {
    ("订单列表-分页")
    POST URL="order/list" struct OrderResponse PAGEABLE listOrders(struct OrderQueryRequest req),
}
```

### 生成的分页响应

Go:
```go
type PageableResult[T any] struct {
    List     []T   `json:"list"`
    Total    int64 `json:"total"`
    Page     int32 `json:"page"`
    PageSize int32 `json:"pageSize"`
}
```

TypeScript:
```typescript
interface PageableResult<T> {
    list: T[];
    total: number;
    page: number;
    pageSize: number;
}
```

### 使用示例

```typescript
const result = await orderClient.listOrders({
    page: 1,
    pageSize: 20,
    status: 1,
    startDate: '2024-01-01',
});

console.log(`Total: ${result.total}`);
console.log(`Page ${result.page} of ${Math.ceil(result.total / result.pageSize)}`);
result.list.forEach(order => {
    console.log(`Order: ${order.orderNo}`);
});
```
