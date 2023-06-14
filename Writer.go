package main

import (
	"encoding/csv"
	"github.com/guardian/fastly-log-xtractor/models"
	"log"
	"os"
)

func AsyncWriter(filename *string, contentCh chan models.FastlyLog) chan error {
	errCh := make(chan error, 1)

	go func() {
		fp, err := os.OpenFile(*filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			errCh <- err
			return
		}
		defer fp.Close()

		writer := csv.NewWriter(fp)
		err = writer.Write(models.CSVHeaderRow())
		if err != nil {
			log.Printf("ERROR Could not write %s: %s", *filename, err)
			errCh <- err
			close(errCh)
			return
		}
		for {
			rec, haveMore := <-contentCh
			if !haveMore {
				log.Print("All records now written")
				close(errCh)
				return
			}

			err = writer.Write(rec.ToCSV())
			if err != nil {
				log.Printf("ERROR Could not write %s: %s", *filename, err)
				errCh <- err
				close(errCh)
				return
			}
		}
	}()
	return errCh
}
