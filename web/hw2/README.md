# 可路由的 middleware 设计

## 目標

- 允許用戶在特定路由註冊 middleware
- middleware結果為所有 route 匹配到的middleware的集合
- 越具體路由越後調度
  - 調度順序： ms3、ms2、ms1
```go
Use("GET", "/a/b", ms1)
Use("GET", "/a/*", ms2)
Use("GET", "/a", ms3)
```