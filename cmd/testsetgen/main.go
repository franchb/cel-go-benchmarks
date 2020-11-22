package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/scylladb/go-set/strset"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	benchmarkv1 "github.com/franchb/cel-go-benchmarks/proto/benchmark/v1"

	"github.com/bxcodec/faker/v3"
)

const (
	defaultSamplesCount = 100
	testdataFilename    = "testdata.ndjson"
)

func main() {
	fmt.Println("cel-go benchmark test data generator")

	if err := run(); err != nil {
		glog.Error("job failed, reason: ", err)
		os.Exit(1)
	}
	glog.Info("job done")
}

func run() error {
	opts, err := parseFlags()
	if err != nil {
		return err
	}
	file, err := os.Create(opts.path)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer file.Close()
	glog.Infof("generating %d samples", opts.count)

	return generate(file, opts.count)
}

type options struct {
	path  string
	count int
}

func parseFlags() (*options, error) {
	path := flag.String("target", "", "target path")
	count := flag.Int("count", defaultSamplesCount, "samples count")
	flag.Parse()

	if *count < 1 {
		return nil, errors.New("zero samples count is not allowed")
	}

	if *path == "" {
		glog.Info("using current working directory")
		path, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("could not find target path: %w", err)
		}
		return &options{
			path:  filepath.Join(path, testdataFilename),
			count: *count,
		}, nil
	} else {
		dir, _ := filepath.Split(*path)
		_, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("could not find target path %s: %w", dir, err)
		}
		return &options{
			path:  *path,
			count: *count,
		}, nil
	}
}

func generate(w io.Writer, count int) error {
	tags := genTags()
	enc := json.NewEncoder(w)
	for i := 0; i < count; i++ {
		tagSubset := genTagSet(tags, rand.Intn(10))

		msg := fake(tagSubset)

		b, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(msg)
		if err != nil {
			return err
		}

		if err := enc.Encode(json.RawMessage(b)); err != nil {
			return err
		}
	}
	return nil
}

func fake(tags []string) *benchmarkv1.Message {
	mac, _ := net.ParseMAC(faker.MacAddress())
	return &benchmarkv1.Message{
		Id:        rand.Int63n(1_000_000),
		Name:      faker.Name(),
		Url:       faker.URL(),
		Fqdn:      faker.DomainName(),
		Ip:        net.ParseIP(faker.IPv4()),
		Mac:       mac,
		Meta1:     faker.Name(),
		Meta2:     faker.MonthName(),
		Meta3:     faker.Email(),
		Meta4:     faker.CCNumber(),
		Meta5:     faker.DayOfMonth(),
		CreatedAt: timestamppb.New(time.Unix(faker.UnixTime(), 0)),
		UpdatedAt: timestamppb.New(time.Unix(faker.UnixTime(), 0)),
		Tags:      tags,
	}
}

func genTagSet(tags []string, count int) []string {
	randomSubset := make([]string, count)
	for i := 0; i < count; i++ {
		randomSubset[i] = tags[rand.Intn(len(tags))]
	}
	return randomSubset
}

func genTags() []string {
	tagsCount := 10000
	tags := strset.New()
	for i := 0; i < tagsCount; i++ {
		tags.Add(faker.Username())
	}
	return tags.List()
}
