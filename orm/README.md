# ORM 之 基準比較

## 心得

- 重構很累 
- 很多關聯性還需待釐清
- **構造結果集** 部分需要多看幾次

從結果來看，明顯地 unsafe 效率好不少，效能約是反射的 **3.6** 倍

```shell
goos: darwin
goarch: arm64
pkg: geektime-go/orm/internal/valuer
BenchmarkSetColumns/reflect-10             10000              1537 ns/op             232 B/op         12 allocs/op
BenchmarkSetColumns/unsafe-10              10000               426.1 ns/op           152 B/op          4 allocs/op
PASS
ok      geektime-go/orm/internal/valuer 0.139s

```