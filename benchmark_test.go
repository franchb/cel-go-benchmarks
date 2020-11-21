package cel_go_benchmarks_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/franchb/cel-go-benchmarks/internal/iterator"
	benchmarkv1 "github.com/franchb/cel-go-benchmarks/proto/benchmark/v1"
)

const (
	defaultSamplesCount = 100
	testdataFilename    = "testdata/testdata_100.ndjson"
)

var iter *iterator.Iterator

func TestMain(m *testing.M) {
	fmt.Println("cel-go benchmark")

	if cache, err := loadCache(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	} else {
		iter = iterator.New(cache)
	}
	m.Run()
}

func BenchmarkCelGo(b *testing.B) {

	e, err := cel.NewEnv(
		cel.ClearMacros(),
		cel.Container("benchmarkv1"),

		cel.Types(&benchmarkv1.Message{}),

		cel.Declarations(
			decls.NewVar("benchmarkv1.Message",
				decls.NewObjectType("benchmarkv1.Message")),
		),
	)
	if err != nil {
		b.Fatalf("environment creation error: %s\n", err)
	}

	// Parse and check the expression.
	p, issues := e.Parse(`Message.Id > 710`)
	if issues != nil && issues.Err() != nil {
		b.Fatal(issues.Err())
	}

	prg, err := e.Program(p, cel.EvalOptions(cel.OptOptimize))
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
					_, _, err := prg.Eval(map[string]interface{}{
						"Message": iter.Next(),
					})
					if err != nil {
						b.Fatalf("runtime error: %s\n", err)
					}
				}
			})
		})
	}

}

func loadCache() ([]*benchmarkv1.Message, error) {
	file, err := os.Open(testdataFilename)
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
