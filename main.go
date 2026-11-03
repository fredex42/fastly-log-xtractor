package main

import (
	"context"
	"flag"
	"log"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	bucketName := flag.String("bucket", "", "bucket to scan")
	prefix := flag.String("prefix", "", "optional prefix to load")
	output := flag.String("out", "fastly-logs.csv", "file to write outputs into")
	parallel := flag.Int("loaders", 1, "number of parallel loaders to use")
	format := flag.String("format", "csv", "output format (csv or sqlite)")
	sqliteCommitEvery := flag.Int("sqlite-commit-every", 10000, "number of rows to commit per transaction when using sqlite")
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	if *parallel < 1 {
		log.Fatalf("Number of parallel loaders must be at least 1")
	}
	if *parallel >= runtime.NumCPU()*2 {
		log.Fatalf("Number of parallel loaders must be no more than %d (core count * 2)", runtime.NumCPU()*2)
	}

	if *sqliteCommitEvery < 1 {
		log.Fatalf("sqlite-commit-every must not be less than 1")
	}

	s3Client := s3.NewFromConfig(cfg)

	keysCh := IterateBucket(s3Client, bucketName, prefix)
	logLinesCh := AsyncFileLoader(s3Client, bucketName, keysCh, *parallel)
	var errCh chan error

	if *format == "csv" {
		errCh = AsyncWriter(output, logLinesCh)
	} else if *format == "sqlite" {
		errCh = AsyncSQLiteWriter(output, logLinesCh, *sqliteCommitEvery)
	} else {
		log.Fatalf("Unknown format %s.  Use either 'csv' or 'sqlite'", *format)
	}

	err, haveMore := <-errCh
	if !haveMore {
		log.Printf("All completed")
	} else {
		log.Print("ERROR ", err)
	}
}
