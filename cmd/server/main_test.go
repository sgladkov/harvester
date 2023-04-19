package main

import (
	"fmt"
	storage2 "github.com/sgladkov/harvester/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsRouter(t *testing.T) {
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
		{
			name: "wrong metric",
			want: want{
				method:  http.MethodPost,
				request: `/update/unknown/test/1`,
				code:    http.StatusBadRequest,
			},
		},
	}
	storage = storage2.NewMemStorage()
	ts := httptest.NewServer(MetricsRouter())
	defer ts.Close()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.want.method, ts.URL+test.want.request, nil)
			require.NoError(t, err)
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func() {
				err := res.Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}()
			assert.Equal(t, test.want.code, res.StatusCode)
			_, err = io.ReadAll(res.Body)
			require.NoError(t, err)
		})
	}
}

func TestCounter(t *testing.T) {
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
			name: "update counter",
			want: want{
				method:  http.MethodPost,
				request: `/update/counter/test/123`,
				code:    http.StatusOK,
			},
		},
		{
			name: "get counter",
			want: want{
				method:  http.MethodGet,
				request: `/value/counter/test`,
				code:    http.StatusOK,
			},
		},
		{
			name: "get unknown counter",
			want: want{
				method:  http.MethodGet,
				request: `/value/counter/unknown`,
				code:    http.StatusNotFound,
			},
		},
	}
	storage = storage2.NewMemStorage()
	ts := httptest.NewServer(MetricsRouter())
	defer ts.Close()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.want.method, ts.URL+test.want.request, nil)
			require.NoError(t, err)
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func() {
				err := res.Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}()
			assert.Equal(t, test.want.code, res.StatusCode)
			_, err = io.ReadAll(res.Body)
			require.NoError(t, err)
		})
	}
}

func TestGauge(t *testing.T) {
	type want struct {
		method  string
		request string
		code    int
		body    string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "update gauge",
			want: want{
				method:  http.MethodPost,
				request: `/update/gauge/test/123.65`,
				code:    http.StatusOK,
				body:    "",
			},
		},
		{
			name: "get gauge",
			want: want{
				method:  http.MethodGet,
				request: `/value/gauge/test`,
				code:    http.StatusOK,
				body:    "123.65",
			},
		},
		{
			name: "get unknown gauge",
			want: want{
				method:  http.MethodGet,
				request: `/value/gauge/unknown`,
				code:    http.StatusNotFound,
				body:    "",
			},
		},
	}
	storage = storage2.NewMemStorage()
	ts := httptest.NewServer(MetricsRouter())
	defer ts.Close()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.want.method, ts.URL+test.want.request, nil)
			require.NoError(t, err)
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func() {
				err := res.Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}()
			assert.Equal(t, test.want.code, res.StatusCode)
			body, err := io.ReadAll(res.Body)
			if len(test.want.body) > 0 {
				assert.Equal(t, test.want.body, string(body))
			}
			require.NoError(t, err)
		})
	}
}
