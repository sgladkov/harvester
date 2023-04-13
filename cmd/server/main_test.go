package main

import (
	"fmt"
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
	ts := httptest.NewServer(MetricsRouter())
	defer ts.Close()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.want.method, ts.URL+test.want.request, nil)
			require.NoError(t, err)
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}(res.Body)
			assert.Equal(t, test.want.code, res.StatusCode)
			_, err = io.ReadAll(res.Body)
			require.NoError(t, err)
		})
	}
}
