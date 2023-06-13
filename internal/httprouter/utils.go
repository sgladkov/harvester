package httprouter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func HashFromData(data []byte, key []byte) string {
	if len(key) == 0 {
		return ""
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	dst := h.Sum(nil)
	return hex.EncodeToString(dst)
}
