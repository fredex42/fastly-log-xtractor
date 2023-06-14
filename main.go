package main

import (
	"context"
	"flag"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
)

func main() {
	bucketName := flag.String("bucket", "", "bucket to scan")
	prefix := flag.String("prefix", "", "optional prefix to load")
	output := flag.String("out", "fastly-logs.csv", "file to write outputs into")
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	keysCh := IterateBucket(s3Client, bucketName, prefix)
	logLinesCh := AsyncFileLoader(s3Client, bucketName, keysCh)
	errCh := AsyncWriter(output, logLinesCh)

	err, haveMore := <-errCh
	if !haveMore {
		log.Printf("All completed")
	} else {
		log.Print("ERROR ", err)
	}
}
