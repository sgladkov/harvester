package httprouter

import (
	"net/http"
	"strings"
)

func ContainsHeaderValue(r *http.Request, header string, value string) bool {
	values := r.Header.Values(header)
	for _, v := range values {
		fields := strings.FieldsFunc(v, func(c rune) bool { return c == ' ' || c == ',' || c == ';' })
		for _, f := range fields {
			if f == value {
				return true
			}
		}
	}
	return false
}
