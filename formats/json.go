package formats

import (
	"context"
	"fmt"
	"os"

	"github.com/olivere/elastic/v7"
)

type JSON struct {
	Outfile *os.File
}

func (j JSON) Run(ctx context.Context, hits <-chan *elastic.SearchHit) error {
	for hit := range hits {
		_, err := fmt.Fprintln(j.Outfile, string(hit.Source))
		if err != nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return nil
}
