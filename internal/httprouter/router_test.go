package httprouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sgladkov/harvester/internal/models"
	storage2 "github.com/sgladkov/harvester/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	ts := httptest.NewServer(MetricsRouter(storage2.NewMemStorage("", false), nil))
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
	ts := httptest.NewServer(MetricsRouter(storage2.NewMemStorage("", false), nil))
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
	ts := httptest.NewServer(MetricsRouter(storage2.NewMemStorage("", false), nil))
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

func TestCounterJSON(t *testing.T) {
	type want struct {
		method      string
		name        string
		status      int
		json        bool
		value       int64
		returnValue int64
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "update counter",
			want: want{
				method:      "/update/",
				name:        `test`,
				status:      http.StatusOK,
				json:        true,
				value:       1,
				returnValue: 1,
			},
		},
		{
			name: "second update counter",
			want: want{
				method:      "/update/",
				name:        `test`,
				status:      http.StatusOK,
				json:        true,
				value:       1,
				returnValue: 2,
			},
		},
		{
			name: "get counter",
			want: want{
				method:      "/value/",
				name:        "test",
				status:      http.StatusOK,
				json:        true,
				returnValue: 2,
			},
		},
		{
			name: "get unknown counter",
			want: want{
				method: "/value/",
				name:   "unknown",
				status: http.StatusNotFound,
				json:   false,
			},
		},
	}
	ts := httptest.NewServer(MetricsRouter(storage2.NewMemStorage("", false), nil))
	defer ts.Close()
	m := models.Metrics{}
	m.MType = "counter"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m.ID = test.want.name
			m.Delta = &test.want.value
			body, _ := json.Marshal(m)
			req, err := http.NewRequest(http.MethodPost, ts.URL+test.want.method, bytes.NewReader(body))
			req.Header["Content-Type"] = []string{"application/json"}
			require.NoError(t, err)
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			reply, _ := io.ReadAll(res.Body)
			err = json.Unmarshal(reply, &m)
			if test.want.json {
				require.NoError(t, err)
				require.Equal(t, test.want.value, *m.Delta)
			}
			defer func() {
				err := res.Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}()
			assert.Equal(t, test.want.status, res.StatusCode)
			_, err = io.ReadAll(res.Body)
			require.NoError(t, err)
		})
	}
}

func TestGougeJSON(t *testing.T) {
	type want struct {
		method      string
		name        string
		status      int
		json        bool
		value       float64
		returnValue float64
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "update gauge",
			want: want{
				method:      "/update/",
				name:        `test`,
				status:      http.StatusOK,
				json:        true,
				value:       1,
				returnValue: 1,
			},
		},
		{
			name: "second update gauge",
			want: want{
				method:      "/update/",
				name:        `test`,
				status:      http.StatusOK,
				json:        true,
				value:       1,
				returnValue: 1,
			},
		},
		{
			name: "get counter",
			want: want{
				method:      "/value/",
				name:        "test",
				status:      http.StatusOK,
				json:        true,
				returnValue: 1,
			},
		},
		{
			name: "get unknown gauge",
			want: want{
				method: "/value/",
				name:   "unknown",
				status: http.StatusNotFound,
				json:   false,
			},
		},
	}
	ts := httptest.NewServer(MetricsRouter(storage2.NewMemStorage("", false), nil))
	defer ts.Close()
	m := models.Metrics{}
	m.MType = "gauge"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m.ID = test.want.name
			m.Value = &test.want.value
			body, _ := json.Marshal(m)
			req, err := http.NewRequest(http.MethodPost, ts.URL+test.want.method, bytes.NewReader(body))
			req.Header["Content-Type"] = []string{"application/json"}
			require.NoError(t, err)
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			reply, _ := io.ReadAll(res.Body)
			fmt.Printf("Reply: %s\n", string(reply))
			err = json.Unmarshal(reply, &m)
			if test.want.json {
				require.NoError(t, err)
				require.Equal(t, test.want.value, *m.Value)
			}
			defer func() {
				err := res.Body.Close()
				if err != nil {
					fmt.Println(err)
				}
			}()
			assert.Equal(t, test.want.status, res.StatusCode)
			_, err = io.ReadAll(res.Body)
			require.NoError(t, err)
		})
	}
}
