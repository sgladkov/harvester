package main

import (
	"github.com/sgladkov/harvester/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhook(t *testing.T) {
	type want struct {
		method  string
		request string
		code    int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "update gauge",
			want: want{
				method:  http.MethodPost,
				request: `/update/gauge/test/1`,
				code:    http.StatusOK,
			},
		},
		{
			name: "update counter",
			want: want{
				method:  http.MethodPost,
				request: `/update/counter/test/1`,
				code:    http.StatusOK,
			},
		},
		{
			name: "wrong counter",
			want: want{
				method:  http.MethodPost,
				request: `/update/counter/test/1.1`,
				code:    http.StatusBadRequest,
			},
		},
		{
			name: "wrong method",
			want: want{
				method:  http.MethodGet,
				request: `/update/counter/test/1`,
				code:    http.StatusMethodNotAllowed,
			},
		},
	}
	storage = internal.NewMemStorage()
	handler := http.StripPrefix("/update/", http.HandlerFunc(webhook))
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.want.method, test.want.request, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			require.NoError(t, err)
		})
	}
}
