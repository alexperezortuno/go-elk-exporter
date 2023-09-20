package export

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"gitlab.com/sixbell/proyectos/brasil/claro/vas/elastic-query-export/logger"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gitlab.com/sixbell/proyectos/brasil/claro/vas/elastic-query-export/flags"
	"gitlab.com/sixbell/proyectos/brasil/claro/vas/elastic-query-export/formats"
	logWrapper "gitlab.com/sixbell/proyectos/brasil/claro/vas/elastic-query-export/logger"
)

const workers = 8

// Formatter defines how an output formatter has to look like
type Formatter interface {
	Run(context.Context, <-chan *elastic.SearchHit) error
}

type RawQuery struct {
	Query *Query `json:"query,omitempty"`
}

type Query struct {
	Bool *Bool `json:"bool,omitempty"`
}

type Bool struct {
	Filter *Filter `json:"filter,omitempty"`
	Must   *Must   `json:"must,omitempty"`
}

type Filter struct {
	Range *Range `json:"range,omitempty"`
}

type Range struct {
	Timestamp *Timestamp `json:"@timestamp,omitempty"`
}

type Timestamp struct {
	From         *string `json:"from,omitempty"`
	IncludeLower *bool   `json:"include_lower,omitempty"`
	IncludeUpper *bool   `json:"include_upper,omitempty"`
	To           *string `json:"to,omitempty"`
}

type Must struct {
	QueryString *QueryString `json:"query_string,omitempty"`
}

type QueryString struct {
	Query *string `json:"query,omitempty"`
}

// Run starts the export of Elastic data
func Run(ctx context.Context, conf *flags.Flags) {
	var loggerWrapper *logWrapper.StandardLogger

	if conf.Debug {
		loggerWrapper = logger.NewLogger(conf)
	}

	if loggerWrapper != nil {
		loggerWrapper.Info("Starting export")
	}

	exportCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !conf.ElasticVerifySSL},
	}
	httpClient := &http.Client{Transport: tr}

	esOpts := make([]elastic.ClientOptionFunc, 0)
	esOpts = append(esOpts,
		elastic.SetHttpClient(httpClient),
		elastic.SetURL(conf.ElasticURL),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(60*time.Second),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
	)

	if conf.Debug {
		esOpts = append(esOpts, elastic.SetTraceLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)))
	}

	if conf.ElasticUser != "" && conf.ElasticPass != "" {
		esOpts = append(esOpts, elastic.SetBasicAuth(conf.ElasticUser, conf.ElasticPass))
	}

	client, err := elastic.NewClient(esOpts...)
	if err != nil {
		log.Fatalf("Error connecting to ElasticSearch: %s", err)
	}
	defer client.Stop()

	if conf.Fieldlist != "" {
		conf.Fields = strings.Split(conf.Fieldlist, ",")
	}

	outfile, err := os.Create(fullFilePath(conf.Outpath, conf.OutFormat, conf.Filename))
	if err != nil {
		log.Fatalf("Error creating output file - %s", err)
	}
	defer outfile.Close()

	var rangeQuery *elastic.RangeQuery

	esQuery := elastic.NewBoolQuery()

	if conf.StartDate != "" && conf.EndDate != "" {
		rangeQuery = elastic.NewRangeQuery(conf.Timefield).Gte(conf.StartDate).Lte(conf.EndDate)
	} else if conf.StartDate != "" {
		rangeQuery = elastic.NewRangeQuery(conf.Timefield).Gte(conf.StartDate)
	} else if conf.EndDate != "" {
		rangeQuery = elastic.NewRangeQuery(conf.Timefield).Lte(conf.EndDate)
	} else {
		rangeQuery = nil
	}

	if rangeQuery != nil {
		esQuery = esQuery.Filter(rangeQuery)
	}

	if conf.RAWQuery != "" {
		if loggerWrapper != nil {
			loggerWrapper.Debug(fmt.Sprintf("rawQuery: %s", conf.RAWQuery))
		}
		esQuery = esQuery.Must(elastic.NewRawStringQuery(conf.RAWQuery))
	} else if conf.Query != "" {
		if loggerWrapper != nil {
			loggerWrapper.Debug(fmt.Sprintf("query: %s", conf.Query))
		}

		newQuery, err := parseQuery(conf.Query, loggerWrapper)
		if err != nil {
			if loggerWrapper != nil {
				loggerWrapper.Debug(fmt.Sprintf("Error parsing query: %s", err))
			}
		}
		esQuery = esQuery.Must(elastic.NewQueryStringQuery(newQuery))
	} else {
		esQuery = esQuery.Must(elastic.NewMatchAllQuery())
	}

	// Count total and setup progress
	total, err := client.Count(conf.Index).Query(esQuery).Do(ctx)
	if err != nil {
		if loggerWrapper != nil {
			loggerWrapper.Error(fmt.Sprintf("Error counting ElasticSearch documents - %v", err))
		}
	}

	if conf.LogResult {
		if loggerWrapper != nil {
			loggerWrapper.Debug(fmt.Sprintf("Total documents to export: %d", total))
		}
	}

	hits := make(chan *elastic.SearchHit)

	go func() {
		defer close(hits)

		scroll := client.Scroll(conf.Index).Size(conf.ScrollSize).Query(esQuery)

		// include selected fields otherwise export all
		if conf.Fields != nil {
			fetchSource := elastic.NewFetchSourceContext(true)
			for _, field := range conf.Fields {
				fetchSource.Include(field)
			}
			scroll = scroll.FetchSourceContext(fetchSource)
		}

		if conf.LogResult {
			if loggerWrapper != nil {
				data, _ := json.MarshalIndent(scroll.Human(true), "", "  ")
				loggerWrapper.Debug(fmt.Sprintf("ElasticSearch query: %s", string(data)))
			}
		}

		for {
			results, err := scroll.Do(ctx)
			if err == io.EOF {
				return // all results retrieved
			}
			if err != nil {
				if loggerWrapper != nil {
					loggerWrapper.Error(fmt.Sprintf("Error scrolling ElasticSearch documents - %v", err))
				}
				cancel()
				return // something went wrong
			}

			// Send the hits to the hits channel
			for _, hit := range results.Hits.Hits {
				// Check if we need to terminate early
				select {
				case hits <- hit:
				case <-exportCtx.Done():
					return
				}
			}
		}
	}()

	var output Formatter
	switch conf.OutFormat {
	case flags.FormatJSON:
		output = formats.JSON{
			Outfile: outfile,
		}
	case flags.FormatRAW:
		output = formats.Raw{
			Outfile: outfile,
		}
	default:
		output = formats.CSV{
			Conf:    conf,
			Outfile: outfile,
			Workers: workers,
		}
	}

	err = output.Run(exportCtx, hits)
	if err != nil {
		if loggerWrapper != nil {
			loggerWrapper.Error(fmt.Sprintf("Error exporting ElasticSearch documents - %v", err))
		}
	}
}

func fullFilePath(path string, format string, filename string) string {
	if filename != "" {
		return path + "/" + filename + "." + format
	}
	return path + "/" + time.Now().Format("20060102150405") + "." + format
}

func parseQuery(query string, loggerWrapper *logWrapper.StandardLogger) (string, error) {
	strs := strings.Split(query, " ")
	var newQuery []string

	for i, str := range strs {
		if loggerWrapper != nil {
			loggerWrapper.Debug(fmt.Sprintf("str: %s %d", str, i))
		}
		if strings.Count(str, ":") > 1 {
			countChar := strings.Count(str, ":")
			var tempString = str
			for j := 1; j < countChar; j++ {
				tempString = replaceNth(tempString, ":", "\\:", j+1)
			}
			newQuery = append(newQuery, tempString)
			continue
		}

		newQuery = append(newQuery, str)
	}

	query = strings.Join(newQuery, " ")
	if loggerWrapper != nil {
		loggerWrapper.Debug(fmt.Sprintf("parse query: %s", query))
	}

	return query, nil
}

func replaceNth(s, old, new string, n int) string {
	i := 0
	for m := 1; m <= n; m++ {
		x := strings.Index(s[i:], old)
		if x < 0 {
			break
		}
		i += x
		if m == n {
			return s[:i] + new + s[i+len(old):]
		}
		i += len(old)
	}
	return s
}
