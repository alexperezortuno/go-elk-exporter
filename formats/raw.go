package formats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/olivere/elastic/v7"
)

type Raw struct {
	Outfile *os.File
}

func (r Raw) Run(ctx context.Context, hits <-chan *elastic.SearchHit) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case hit := <-hits:
			data, err := json.Marshal(hit)
			if err != nil {
				log.Println(err)
				continue
			}
			_, err = fmt.Fprintln(r.Outfile, string(data))
			if err != nil {
				log.Println(err)
			}
		}
	}
}
