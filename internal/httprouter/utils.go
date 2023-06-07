package httprouter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sgladkov/harvester/internal/logger"
	"go.uber.org/zap"
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

func HashFromData(data any, key []byte) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	bytesToHash, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	logger.Log.Info("Marshalled", zap.String("data", string(bytesToHash)))
	h := hmac.New(sha256.New, key)
	h.Write(bytesToHash)
	dst := h.Sum(nil)
	return hex.EncodeToString(dst), nil
}
