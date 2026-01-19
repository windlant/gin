---

# User API

## `GET /users`
获取所有用户。

**响应示例：**
```json
[
  {"id":1,"name":"Alice","email":"alice@example.com"},
  {"id":2,"name":"Bob","email":"bob@example.com"}
]
```

---

## `GET /users/:id`
根据 ID 获取单个用户。

**路径参数：**  
- `id` (int)

**成功响应（200）：**
```json
{"id":1,"name":"Alice","email":"alice@example.com"}
```

**失败响应（404）：**
```json
{"error":"User not found"}
```

---

## `POST /users`
批量创建用户。

**请求体（JSON 数组）：**
```json
[
  {"name":"Bob","email":"bob@example.com"},
  {"name":"Charlie","email":"charlie@example.com"}
]
```

**响应（201）：**
```json
{
  "created": [
    {"id":2,"name":"Bob","email":"bob@example.com"},
    {"id":3,"name":"Charlie","email":"charlie@example.com"}
  ],
  "errors": []
}
```

---

## `PUT /users`
批量更新用户。

**请求体（JSON 数组，每项必须含 `id`）：**
```json
[
  {"id":2,"name":"Robert","email":"robert@example.com"}
]
```

**响应（200）：**
```json
{
  "updated": [
    {"id":2,"name":"Robert","email":"robert@example.com"}
  ],
  "not_found": [999]
}
```

---

## `DELETE /users`
批量删除用户。

**请求体（JSON 对象）：**
```json
{"ids": [1, 2, 999]}
```

**响应（200）：**
```json
{
  "deleted": [1, 2],
  "not_found": [999]
}
```