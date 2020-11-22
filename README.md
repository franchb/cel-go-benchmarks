# cel-go-benchmarks
Benchmark [github.com/google/cel-go](github.com/google/cel-go) for different payloads and use cases.

## Results


`testdata_100.ndjson` - 100 unique messages

```shell
$ make benchmark
go test -run=Bench -bench=. -benchmem
cel-go benchmark
Loaded 100 records
goos: linux
goarch: amd64
pkg: github.com/franchb/cel-go-benchmarks
BenchmarkCelGo/1-4       9491524               123 ns/op              64 B/op          3 allocs/op
BenchmarkCelGo/2-4       9614991               125 ns/op              64 B/op          3 allocs/op
BenchmarkCelGo/4-4       9288651               124 ns/op              64 B/op          3 allocs/op
BenchmarkExpr/1-4        6305431               186 ns/op              96 B/op          5 allocs/op
BenchmarkExpr/2-4        5647844               186 ns/op              96 B/op          5 allocs/op
BenchmarkExpr/4-4        6359542               190 ns/op              96 B/op          5 allocs/op
PASS
ok      github.com/franchb/cel-go-benchmarks    8.791s
```

10 000 unique messages (need to generate it first):

```shell
$ make generate-samples COUNT=10000
$ make benchmark SRC=./testdata/testdata_10000.ndjson
BENCHMARK_DATA=./testdata/testdata_10000.ndjson go test -run=Bench -bench=. -benchmem
cel-go benchmark
Loaded 10000 records
goos: linux
goarch: amd64
pkg: github.com/franchb/cel-go-benchmarks
BenchmarkCelGo/1-4       8410110               140 ns/op              64 B/op          3 allocs/op
BenchmarkCelGo/2-4       8466694               140 ns/op              64 B/op          3 allocs/op
BenchmarkCelGo/4-4       8552630               143 ns/op              64 B/op          3 allocs/op
BenchmarkExpr/1-4        5584927               203 ns/op              96 B/op          5 allocs/op
BenchmarkExpr/2-4        5696047               205 ns/op              96 B/op          5 allocs/op
BenchmarkExpr/4-4        5739594               205 ns/op              96 B/op          5 allocs/op
PASS
ok      github.com/franchb/cel-go-benchmarks    10.665s

```

When cel-go is used right - it's fast. That's what this benchmark I made for.