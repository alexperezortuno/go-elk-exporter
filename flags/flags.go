package flags

const (
	FormatCSV  = "csv"
	FormatJSON = "json"
	FormatRAW  = "raw"
)

type Flags struct {
	ElasticURL       string `cli:"connect" cliAlt:"c" usage:"ElasticSearch URL"`
	ElasticUser      string `cli:"user" usage:"ElasticSearch Username"`
	ElasticPass      string `cli:"pass" usage:"ElasticSearch Password"`
	ElasticVerifySSL bool   `cli:"verifySSL" usage:"Verify SSL certificate"`
	Index            string `cli:"index" cliAlt:"i" usage:"ElasticSearch Index (or Index Prefix)"`
	RAWQuery         string `cli:"rawquery" cliAlt:"r" usage:"ElasticSearch raw query string"`
	Query            string `cli:"query" cliAlt:"q" usage:"Lucene query same that is used in Kibana search input"`
	OutFormat        string `cli:"outformat" cliAlt:"f" usage:"Format of the output data. [json|csv]"`
	Filename         string `cli:"filename" cliAlt:"fn" usage:"File name output"`
	Outpath          string `cli:"outpath" cliAlt:"o" usage:"Path to output file"`
	StartDate        string `cli:"start" cliAlt:"s" usage:"Start date for included documents"`
	EndDate          string `cli:"end" cliAlt:"e" usage:"End date for included documents"`
	ScrollSize       int    `cli:"size" usage:"Number of documents that will be returned per shard"`
	Timefield        string `cli:"timefield" usage:"Field name to use for start and end date query"`
	Timezone         string `cli:"timezone" usage:"Timezone to use for parse dates response"`
	Fieldlist        string `cli:"fields" usage:"Fields to include in export as comma separated list"`
	Convertdate      bool   `cli:"convertdate" usage:"Convert UTC to date format with timezone in response"`
	Dateformat       string `cli:"dateformat" usage:"Date format to use for convert dates default 2006-01-12 15:03:05"`
	Trace            bool   `cli:"trace" usage:"Enable debug output"`
	Debug            bool   `cli:"logconsole" usage:"Enable debug output"`
	LogLevel         string `cli:"loglevel" usage:"Log level"`
	LogOutput        string `cli:"logoutput" usage:"Log output path"`
	LogFile          bool   `cli:"logfile" usage:"Log to file"`
	LogFormat        string `cli:"logformat" usage:"Log format [text|json]"`
	LogResult        bool   `cli:"logresult" usage:"Log elasticsearch query result"`
	Fields           []string
}
