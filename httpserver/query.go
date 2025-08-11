package httpserver

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// ReplacementFields contains the fields that need to be replaced in the query string.
type ReplacementFields struct {
	// Fields that apply to the entire row.
	Keys []string

	// Fields related to the filter key `q`, which are then included in the SQL query.
	Query []string
}

type Param struct {
	Q       map[string]string `json:"q"`
	Sort    string            `json:"sort"`
	Offset  int               `json:"offset"`
	Limit   int               `json:"limit"`
	Mutator map[string]string `json:"mutator"`
}

func NewParam() *Param {
	return &Param{
		Q:       make(map[string]string),
		Mutator: make(map[string]string),
	}
}

// ParseQueryString parses the query string, does unescaping.
func (s *Server) ParseQueryString(c echo.Context, mapData *ReplacementFields) (*Param, error) { //nolint: cyclop, lll // nothing to simplify
	rawQuery := c.QueryString()

	if mapData != nil && mapData.Keys != nil {
		r := strings.NewReplacer(mapData.Keys...)
		rawQuery = r.Replace(rawQuery)
	}

	param := NewParam()

	for _, condition := range strings.Split(rawQuery, "&") {
		couple := strings.Split(condition, "=")

		if len(couple) <= 1 {
			continue
		}

		key := couple[0]
		value := couple[1]

		var err error

		switch key {
		case "q":
			param.Q, err = parseQuery(value, mapData)
		case "m":
			param.Mutator, err = parseMutators(value)
		case "sort":
			param.Sort, err = url.QueryUnescape(value)
		case "offset":
			param.Offset, err = parseInt(value)
		case "limit":
			param.Limit, err = parseInt(value)
		}

		if err != nil {
			return nil, err
		}
	}

	return param, nil
}

func parseQuery(value string, mapData *ReplacementFields) (map[string]string, error) {
	result := make(map[string]string)

	if mapData != nil && mapData.Query != nil {
		r := strings.NewReplacer(mapData.Query...)
		value = r.Replace(value)
	}

	queryString, err := url.QueryUnescape(value)
	if err != nil {
		return result, err
	}

	qsplit := strings.Split(queryString, ",")
	for _, qq := range qsplit {
		qpar := strings.Split(qq, ":")
		if len(qpar) == 2 { //nolint:mnd // Not a magic, need to validate couples
			result[qpar[0]] = qpar[1]
		}
	}

	return result, nil
}

func parseMutators(value string) (map[string]string, error) {
	result := make(map[string]string)

	muts, err := url.QueryUnescape(value)
	if err != nil {
		return result, err
	}

	mutators := strings.Split(muts, ",")

	for _, m := range mutators {
		if m == "" {
			continue
		}

		result[m] = m
	}

	return result, nil
}

func parseInt(value string) (int, error) {
	offset, err2 := url.QueryUnescape(value)
	if err2 != nil {
		return 0, err2
	}

	return strconv.Atoi(offset)
}
