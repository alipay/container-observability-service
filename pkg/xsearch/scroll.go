package xsearch

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/olivere/elastic"
	"golang.org/x/sync/errgroup"

	"k8s.io/klog/v2"
)

// Scroll use scroll service to query multiple documents from es.
func Scroll(client *elastic.Client, index string, query elastic.Query, f func(json.RawMessage) error, workers int) error {
	pageSize := 1000
	totalGot := 0
	hits := make(chan json.RawMessage)
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		defer close(hits)
		// Initialize scroller. Just don't call Do yet.
		var scroll *elastic.ScrollService
		if strings.Index(index, ",") > 0 {
			// search multiple index
			indices := strings.Split(index, ",")
			scroll = client.Scroll(indices...).
				Size(pageSize)
		} else {
			scroll = client.Scroll(index).
				Size(pageSize)
		}
		if query != nil {
			scroll = scroll.Query(query)
		}

		for {
			results, err := scroll.Do(ctx)
			if err == io.EOF {
				// all results retrieved
				return nil
			}
			if err != nil {
				klog.Errorf("[index: %s]scroll.Do failed: %s", index, err.Error())
				return err
			}

			// Send the hits to the hits channel
			totalGot += len(results.Hits.Hits)
			for _, hit := range results.Hits.Hits {
				select {
				case hits <- *hit.Source:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	})

	// 2nd goroutine receives hits and deserializes them.
	for i := 0; i < workers; i++ {
		g.Go(func() error {
			for hit := range hits {

				// Deserialize and process
				err := f(hit)
				if err != nil {
					klog.Errorf("failed to process hit result: %s", err.Error())
				}
				// Terminate early?
				select {
				default:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	// Check whether any goroutines failed.
	err := g.Wait()
	if err != nil {
		klog.Errorf("failed to wait all goroutines finish when scroll elasticsearch result: %s", err.Error())
	}
	return err
}
