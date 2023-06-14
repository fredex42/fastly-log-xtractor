package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
)

func IterateBucket(client s3.ListObjectsV2APIClient, bucketName *string, prefix *string) chan string {
	outputCh := make(chan string, 100)

	go func() {
		var token *string = nil
		for {
			req := &s3.ListObjectsV2Input{
				Bucket:            bucketName,
				ContinuationToken: token,
				Prefix:            prefix,
				MaxKeys:           100,
			}
			response, err := client.ListObjectsV2(context.Background(), req)
			if err != nil {
				log.Printf("ERROR Could not list %s: %s", *bucketName, err)
				close(outputCh)
				return
			}

			for _, obj := range response.Contents {
				outputCh <- *obj.Key
			}

			token = response.NextContinuationToken
			if !response.IsTruncated {
				//we are all done
				log.Printf("INFO Found all items in %s", *bucketName)
				close(outputCh)
				return
			}
		}
	}()

	return outputCh
}
