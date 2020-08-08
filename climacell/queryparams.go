package climacell

import (
	"fmt"
	"strings"
)

type QueryParams struct {
	flags  map[string]string
	fields []string
}

func (q QueryParams) String() string {
	flags := ""
	fields := ""
	if len(q.fields) > 0 {
		fields = fmt.Sprintf("fields=%s", strings.Join(q.fields, `%2C`))
	}
	for key, value := range q.flags {
		flags = fmt.Sprintf("%s&%s=%s", flags, key, value)
	}
	return fmt.Sprintf("%s&%s", flags, fields)
}

func buildURL(endpoint string, queryParams *QueryParams) string {
	return fmt.Sprintf("%s%s?apikey=%s%s", apiURL, endpoint, apiKey, queryParams)
}
