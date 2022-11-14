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

## Benchmark

- 不確定是否標準
- 靜態路由
```go
// 靜態路由
goos: darwin
goarch: arm64
pkg: geektime-go/web
Benchmark_findRoute_Static
Benchmark_findRoute_Static/method_not_found
Benchmark_findRoute_Static/method_not_found-10         	1000000000	         0.0000013 ns/op
Benchmark_findRoute_Static/path_not_found
Benchmark_findRoute_Static/path_not_found-10           	1000000000	         0.0000024 ns/op
Benchmark_findRoute_Static/root
Benchmark_findRoute_Static/root-10                     	1000000000	         0.0000036 ns/op
Benchmark_findRoute_Static/user_home
Benchmark_findRoute_Static/user_home-10                	1000000000	         0.0000066 ns/op
Benchmark_findRoute_Static/order_detail
Benchmark_findRoute_Static/order_detail-10             	1000000000	         0.0000027 ns/op
PASS
```

- 通配符路由
```go
// 通配符路由
goos: darwin
goarch: arm64
pkg: geektime-go/web
Benchmark_findRoute_Any
Benchmark_findRoute_Any/star_match
Benchmark_findRoute_Any/star_match-10         	1000000000	         0.0000044 ns/op
Benchmark_findRoute_Any/star_in_middle
Benchmark_findRoute_Any/star_in_middle-10     	1000000000	         0.0000040 ns/op
PASS
```

- 參數路由
```go
// 參數路由
goos: darwin
goarch: arm64
pkg: geektime-go/web
Benchmark_findRoute_Param
Benchmark_findRoute_Param/:id
Benchmark_findRoute_Param/:id-10         	1000000000	         0.0000059 ns/op
Benchmark_findRoute_Param/:id*
Benchmark_findRoute_Param/:id*-10        	1000000000	         0.0000051 ns/op
Benchmark_findRoute_Param/:id*#01
Benchmark_findRoute_Param/:id*#01-10     	1000000000	         0.0000066 ns/op
PASS
```

- 正則路由
```go
// 正則路由
goos: darwin
goarch: arm64
pkg: geektime-go/web
Benchmark_findRoute_RegExpr
Benchmark_findRoute_RegExpr/:id(.*)
Benchmark_findRoute_RegExpr/:id(.*)-10         	1000000000	         0.0000130 ns/op
Benchmark_findRoute_RegExpr/:id([0-9]+)
Benchmark_findRoute_RegExpr/:id([0-9]+)-10     	1000000000	         0.0000070 ns/op
PASS
```

### 再跑benchmark一次
- 在相同設備、相同次數、相同路徑長度下
- 判斷單純的`靜態路由`速度最快
- 而參數路由需要解析取值`paraString`
- 正則路由除了取值外，還需要進行正則判斷，故花費較長時間比參數路由多了`1.06`倍

- 花費時間基準比較： 正則路由 > 參數路由 > 通配符路由 > 靜態路由
```go
// go test -run none -bench=. -benchmem
goos: darwin
goarch: arm64
pkg: geektime-go/web
Benchmark_findRoute_Static/method_not_found-10          1000000000               0.0000010 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Static/path_not_found-10            1000000000               0.0000042 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Static/root-10                      1000000000               0.0000055 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Static/user_home-10                 1000000000               0.0000020 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Static/order_detail-10              1000000000               0.0000050 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Any/star_match-10                   1000000000               0.0000028 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Any/star_in_middle-10               1000000000               0.0000023 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Param/:id-10                        1000000000               0.0000063 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Param/:id*-10                       1000000000               0.0000025 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_Param/:id*#01-10                    1000000000               0.0000034 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_RegExpr/:id(.*)-10                  1000000000               0.0000058 ns/op               0 B/op          0 allocs/op
Benchmark_findRoute_RegExpr/:id([0-9]+)-10              1000000000               0.0000073 ns/op               0 B/op          0 allocs/op
PASS
ok      geektime-go/web 0.281s
```
- 透過`pprof`查看profile
  - `runtime.kevent`很像是`darwin`的file I/O開銷
  - 我們可以從`cpu`觀察到，時間花在`runtime.pthread_cond_wait`線程切換
  - 從`memory`觀察到`正則匹配`，在內存內花費`528.17kB`、` 512.31kB`進行匹配操作
```shell
# 產生 cpu、memory profile
go test -bench=.  \
-benchmem -memprofile=mem.pprof \
-cpuprofile=cpu.pprof \

# 使用 pprof 查看數據
go tool pprof cpu.pprof (mem.pprof)

# ==> cpu
Type: cpu
Showing nodes accounting for 30ms, 100% of 30ms total
Showing top 10 nodes out of 21
      flat  flat%   sum%        cum   cum%
      10ms 33.33% 33.33%       10ms 33.33%  runtime.kevent
      10ms 33.33% 66.67%       10ms 33.33%  runtime.pthread_cond_wait
      10ms 33.33%   100%       10ms 33.33%  runtime.pthread_kill
         0     0%   100%       10ms 33.33%  runtime.findRunnable
         0     0%   100%       10ms 33.33%  runtime.mPark (inline)
         0     0%   100%       10ms 33.33%  runtime.mcall
         0     0%   100%       10ms 33.33%  runtime.netpoll
         0     0%   100%       10ms 33.33%  runtime.notesleep
         0     0%   100%       10ms 33.33%  runtime.park_m
         0     0%   100%       10ms 33.33%  runtime.preemptM

# ==> memory
Type: alloc_space
Showing nodes accounting for 6620.42kB, 100% of 6620.42kB total
Showing top 10 nodes out of 48
      flat  flat%   sum%        cum   cum%
 2050.25kB 30.97% 30.97%  2050.25kB 30.97%  runtime.allocm
 1024.41kB 15.47% 46.44%  1024.41kB 15.47%  runtime.malg
  902.59kB 13.63% 60.08%   902.59kB 13.63%  compress/flate.NewWriter
  578.66kB  8.74% 68.82%   578.66kB  8.74%  runtime/pprof.StartCPUProfile
  528.17kB  7.98% 76.79%   528.17kB  7.98%  regexp.(*bitState).reset
  512.31kB  7.74% 84.53%   512.31kB  7.74%  regexp/syntax.(*compiler).inst (inline)
  512.02kB  7.73% 92.27%   512.02kB  7.73%  crypto/internal/nistec.init
  512.02kB  7.73%   100%   512.02kB  7.73%  runtime.gcBgMarkWorker
         0     0%   100%   902.59kB 13.63%  compress/gzip.(*Writer).Write
         0     0%   100%   512.31kB  7.74%  github.com/davecgh/go-spew/spew.init

```