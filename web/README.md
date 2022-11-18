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
- 在相同設備、相同次數、相同路徑長度下

- 花費時間基準比較： 正則路由 > 參數路由 > 通配符路由 > 靜態路由
- 通過分析比較：
  - 消耗時間： 參數路由 > 正則路由 > 靜態路由 > 通配符路由
  - 內存使用： 參數路由 > 正則路由 > 靜態路由 > 通配符路由
  - 通配符路由最快，參數路由最慢

```go
// go test -run none -bench=. -benchmem
goos: darwin
goarch: arm64
pkg: geektime-go/web
Benchmark_findRoute_Static-10            6265520               174.6 ns/op           112 B/op          5 allocs/op
Benchmark_findRoute_Any-10               7515556               159.2 ns/op           112 B/op          4 allocs/op
Benchmark_findRoute_Param-10             2002467               599.7 ns/op          1232 B/op         15 allocs/op
Benchmark_findRoute_RegExpr-10           2969036               410.0 ns/op           832 B/op         10 allocs/op
PASS
ok      geektime-go/web 6.686s
```
- 透過`pprof`查看profile
  - `runtime.kevent`很像是`darwin`的file I/O開銷
  - 我們可以從`cpu`觀察到，時間花在`runtime.pthread_cond_wait`線程切換
  - 從`memory`觀察到`正則匹配`、`參數路由`，在內存內花費明顯高一些
```shell
# 產生 cpu、memory profile
go test -bench=.  \
-benchmem -memprofile=mem.pprof \
-cpuprofile=cpu.pprof \

# 使用 pprof 查看數據
go tool pprof cpu.pprof (mem.pprof)

# ==> cpu
Type: cpu
Showing nodes accounting for 5090ms, 78.67% of 6470ms total
Dropped 71 nodes (cum <= 32.35ms)
Showing top 10 nodes out of 90
      flat  flat%   sum%        cum   cum%
    2160ms 33.38% 33.38%     2160ms 33.38%  runtime.kevent
     570ms  8.81% 42.19%      570ms  8.81%  runtime.madvise
     510ms  7.88% 50.08%     1160ms 17.93%  runtime.mallocgc
     500ms  7.73% 57.81%      500ms  7.73%  runtime.pthread_cond_wait
     480ms  7.42% 65.22%      480ms  7.42%  runtime.pthread_kill
     270ms  4.17% 69.40%      300ms  4.64%  runtime.heapBitsSetType
     190ms  2.94% 72.33%      190ms  2.94%  runtime.usleep
     150ms  2.32% 74.65%      160ms  2.47%  runtime.nextFreeFast (inline)
     130ms  2.01% 76.66%      130ms  2.01%  countbytebody
     130ms  2.01% 78.67%      130ms  2.01%  indexbytebody

# ==> memory
Type: alloc_space
Showing nodes accounting for 8327.90MB, 99.89% of 8337.23MB total
Dropped 57 nodes (cum <= 41.69MB)
      flat  flat%   sum%        cum   cum%
 6280.33MB 75.33% 75.33%  8327.90MB 99.89%  geektime-go/web.(*Router).findRoute
 2047.57MB 24.56% 99.89%  2047.57MB 24.56%  strings.genSplit
         0     0% 99.89%   891.03MB 10.69%  geektime-go/web.Benchmark_findRoute_Any
         0     0% 99.89%  3483.71MB 41.78%  geektime-go/web.Benchmark_findRoute_Param
         0     0% 99.89%  3173.14MB 38.06%  geektime-go/web.Benchmark_findRoute_RegExpr
         0     0% 99.89%   780.52MB  9.36%  geektime-go/web.Benchmark_findRoute_Static
         0     0% 99.89%  2047.57MB 24.56%  strings.Split (inline)
         0     0% 99.89%  8328.40MB 99.89%  testing.(*B).launch
         0     0% 99.89%  8328.91MB 99.90%  testing.(*B).runN

```