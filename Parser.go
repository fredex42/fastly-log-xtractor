package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/guardian/fastly-log-xtractor/models"
)

func ParseLine(content string) (*models.FastlyLog, error) {
	var output models.FastlyLog
	err := json.Unmarshal([]byte(content), &output)
	if err != nil {
		return nil, err
	} else {
		return &output, nil
	}
}

func ParseFile(client *s3.Client, bucketName *string, key string, outputCh chan models.FastlyLog) error {
	req := &s3.GetObjectInput{
		Bucket: bucketName,
		Key:    &key,
	}
	log.Printf("INFO Loading s3://%s/%s", *bucketName, key)

	response, err := client.GetObject(context.Background(), req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	reader := response.Body
	if strings.HasSuffix(key, ".gz") {
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return err
		}
	}

	scanner := bufio.NewScanner(reader)
	c := 0
	for scanner.Scan() {
		c += 1
		model, err := ParseLine(scanner.Text())
		if err == nil {
			outputCh <- *model
		} else {
			log.Printf("WARNING: Could not unmarshal line %d from %s: %s", c, key, err)
		}
	}
	return scanner.Err() //this will be nil if the scanner completed successfully
}

func AsyncFileLoader(client *s3.Client, bucketName *string, keysInput chan string, parallel int) chan models.FastlyLog {
	outputCh := make(chan models.FastlyLog, 100)

	wg := sync.WaitGroup{}
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			for {
				key, haveMore := <-keysInput
				if !haveMore {
					log.Printf("INFO Loaded all content")
					wg.Done()
					return
				}

				err := ParseFile(client, bucketName, key, outputCh)
				if err != nil {
					log.Printf("ERROR Could not read in s3://%s/%s", *bucketName, key)
					wg.Done()
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(outputCh)
	}()

	return outputCh
}
