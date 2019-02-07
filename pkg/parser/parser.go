package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// commonLogFormat is a regular expression that captures all fields in a line formatted according to the Common LogFile
// format. (see: https://www.w3.org/Daemon/User/Config/Logging.html#common-logfile-format)
var commonLogFormat = regexp.MustCompile(`^(?P<remote_host>\S+) (?P<rfc931>\S+) (?P<user>\S+) \[(?P<datetime>[^\]]+)\] "(?P<request_method>[A-Z]+) (?P<request_uri>[^\s"]+) HTTP/(?P<http_version>[0-9.]+)" (?P<status_code>\d{3}) (?P<response_size>[\d-]+)`)

const (
	// datetimeLayout is a time.Parse() compatible layout string formatted according to strftime format: `%d/%b/%Y:%H:%M:%S %z`
	datetimeLayout = "02/Jan/2006:15:04:05 -0700"

	// `-` in a field indicates missing data
	missingValue = "-"
)

// Request represents an HTTP request
type Request struct {
	RemoteHost        string
	RemoteLogname     string
	AuthUser          string
	Timestamp         *time.Time
	Method            string
	URI               string
	HTTPVersion       string
	StatusCode        int
	ResponseSizeBytes int
}

// Section returns the website section, which is defined as the first path before the
// second '/' in the URI.
func (r *Request) Section() string {
	return "/" + strings.Split(r.URI, "/")[1]
}

// Parser is a log line parser
type Parser struct {
	lineFormat *regexp.Regexp
}

// NewParser initializes and returns a new Parser configured for Common LogFile format by default.
func NewParser() *Parser {
	return &Parser{lineFormat: commonLogFormat}
}

// fieldsByName maps all of the submatches to the regex subexpression names.
func fieldsByName(matches, fields []string) map[string]string {
	m := make(map[string]string, len(matches))
	for i, field := range fields {
		m[field] = matches[i]
	}
	return m
}

// ParseLine parsers a line log line and returns any parsing errors
func (p *Parser) ParseLine(line string) (*Request, error) {
	var err error
	matches := p.lineFormat.FindStringSubmatch(line)

	if len(matches) == 0 {
		return nil, fmt.Errorf("failed to parse log line")
	}

	// NOTE: The first element in the set of matches is the input string and the
	// first element of SubexpNames() is always empty string (see: https://golang.org/pkg/regexp/#Regexp.SubexpNames)
	fields := fieldsByName(matches[1:], p.lineFormat.SubexpNames()[1:])

	r := new(Request)
	r.RemoteHost = fields["remote_host"]
	r.RemoteLogname = fields["rfc931"]
	r.AuthUser = fields["user"]
	r.Method = fields["request_method"]
	r.URI = fields["request_uri"]
	r.HTTPVersion = fields["http_version"]

	// TODO: don't set missing values to a default value, this will effect averages
	status := 0
	if fields["status_code"] != missingValue {
		status, err = strconv.Atoi(fields["status_code"])
		if err != nil {
			return nil, fmt.Errorf("failed to convert HTTP status code to int")
		}
	}
	r.StatusCode = status

	size := 0
	if fields["response_size"] != missingValue {
		size, err = strconv.Atoi(fields["response_size"])
		if err != nil {
			return nil, fmt.Errorf("failed to convert ResponseSize to int")
		}
	}
	r.ResponseSizeBytes = size

	ts, err := time.Parse(datetimeLayout, fields["datetime"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse request timestamp: %s", err.Error())
	}
	r.Timestamp = &ts

	return r, nil
}
