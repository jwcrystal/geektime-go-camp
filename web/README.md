# 實現一棵路由樹

- 實現通配符路由
- 實現正則匹配
- 寫測試

## 構建思路

- 通配路由末尾多路由
  - 簡單思考為多路由皆會映射到`通配符`字段
  - /a/b => /a/*
  - /a/b/c => /a/*
  - /a/b/c/d => /a/*
```go
// findRoute
for _, seg := range segs {
    child, paramChild, found := root.childOf(seg)
    if !found {
        // 檢查是否為通配末尾，支援多段路由
        // 可以用 type區分 ，或是 通配後字段是否結束 來區分
        // /order/*
        // /order/detail/123 (x)
        // /order/detail/123/456 (x)
        // /order/detail/123/456/789 (x)
        // 要找最後為通配的字段，所以用root，child會採用當前字段
        if root.nodeType == nodeTypeAny {
            matchInfo.node = root
            return matchInfo, true
        }
        return nil, false
    }
```

- 正則路由匹配
  - 正則格式： `:<param>(<regExp>)`
  - e.g. :id(^[0-9]+$)
```go
	// path: :<param>(<regExp>)
	// remove ":"
	path = path[1:]
	segs := strings.Split(path, "(")
	//segs := strings.SplitN(path, "(",2)
	// :reg(xx) (o) => 2 segs, maybe
	// :reg(xx (x) => 2 segs, maybe
	// :reg(x(x (x) => 3 segs
	// :regxx (x)
	if len(segs) == 2 {
		// last element is ")"
		expr := segs[1]
		if expr[len(expr)-1] == ')' {
			return segs[0], expr[:len(expr)-1], true
		}
	}
	return path, "", false
```