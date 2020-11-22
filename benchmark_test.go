package cel_go_benchmarks_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/antonmedv/expr"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/interpreter"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/franchb/cel-go-benchmarks/internal/iterator"
	benchmarkv1 "github.com/franchb/cel-go-benchmarks/proto/benchmark/v1"
)

const (
	testdataFilename = "testdata/testdata_100.ndjson"
)

var (
	iter     *iterator.Iterator
	celGoEnv *cel.Env
	exprEnv  expr.Option
)

func BenchmarkCelGo(b *testing.B) { benchmarkCelGo(b, `Message.meta2 == "March"`) }
func BenchmarkExpr(b *testing.B)  { benchmarkExpr(b, `Message.Meta2 == "March"`) }

func benchmarkCelGo(b *testing.B, expression string) {
	// Parse and check the expression.
	p, issues := celGoEnv.Compile(expression)
	if issues != nil && issues.Err() != nil {
		b.Fatal(issues.Err())
	}

	prg, err := celGoEnv.Program(p, cel.EvalOptions(cel.OptOptimize))
	if err != nil {
		b.Fatalf("program creation error: %s\n", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 1; i <= 4; i *= 2 {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, _, err := prg.Eval(&Request{Message: iter.Next()})
					if err != nil {
						b.Fatalf("runtime error: %s\n", err)
					}
				}
			})
		})
	}
}

func benchmarkExpr(b *testing.B, expression string) {
	program, err := expr.Compile(expression, exprEnv)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 1; i <= 4; i *= 2 {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := expr.Run(program, &Request{Message: iter.Next()})
					if err != nil {
						b.Fatalf("runtime error: %s\n", err)
					}
				}
			})
		})
	}
}

func TestMain(m *testing.M) {
	fmt.Println("cel-go benchmark")

	if cache, err := loadCache(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	} else {
		iter = iterator.New(cache)
	}

	ce, err := prepareCelGoEnv()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	celGoEnv = ce
	exprEnv = prepareExprEnv()

	m.Run()
}

func prepareCelGoEnv() (*cel.Env, error) {
	e, err := cel.NewEnv(
		cel.ClearMacros(),
		cel.Container("benchmarkv1"),

		cel.Types(&benchmarkv1.Message{}),

		cel.Declarations(
			decls.NewVar("Message",
				decls.NewObjectType("benchmark.v1.Message")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("environment creation error: %w", err)
	}
	return e, nil
}

type Request struct {
	interpreter.Activation
	Message *benchmarkv1.Message
}

func (r *Request) ResolveName(name string) (interface{}, bool) {
	if name == "Message" {
		return r.Message, true
	}
	return nil, false
}

func prepareExprEnv() expr.Option {
	return expr.Env(&Request{})
}

func loadCache() ([]*benchmarkv1.Message, error) {
	filename := os.Getenv(`BENCHMARK_DATA`)
	if filename == "" {
		filename = testdataFilename
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create target file: %w", err)
	}
	defer file.Close()
	cache, err := load(file)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Loaded %d records\n", len(cache))
	return cache, nil
}

func load(r io.Reader) ([]*benchmarkv1.Message, error) {
	dec := json.NewDecoder(r)
	cache := make([]*benchmarkv1.Message, 0)
	for dec.More() {
		var b json.RawMessage
		if err := dec.Decode(&b); err != nil {
			return nil, err
		}
		var m benchmarkv1.Message
		err := protojson.UnmarshalOptions{
			AllowPartial:   false,
			DiscardUnknown: false,
		}.Unmarshal(b, &m)
		if err != nil {
			return nil, err
		}
		cache = append(cache, &m)
	}
	return cache, nil
}
