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

	s3Client := s3.NewFromConfig(cfg)

	keysCh := IterateBucket(s3Client, bucketName, prefix)
	logLinesCh := AsyncFileLoader(s3Client, bucketName, keysCh, *parallel)
	errCh := AsyncWriter(output, logLinesCh)

	err, haveMore := <-errCh
	if !haveMore {
		log.Printf("All completed")
	} else {
		log.Print("ERROR ", err)
	}
}
