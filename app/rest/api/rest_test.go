package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theshamuel/gemini-proxy/app/service"
	"go.uber.org/goleak"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRest_Shutdown(t *testing.T) {
	srv := Rest{}
	done := make(chan bool)

	go func() {
		time.Sleep(200 * time.Millisecond)
		srv.Shutdown()
		close(done)
	}()

	st := time.Now()
	srv.Run(8888)
	assert.True(t, time.Since(st).Seconds() < 1, "should take about 1s")
	<-done
}

func TestRest_Run(t *testing.T) {
	srv := Rest{}
	port := generateRndPort()
	go func() {
		srv.Run(port)
	}()

	waitHTTPServer(port)

	client := http.Client{}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	srv.Shutdown()
}

func TestRest_Ping(t *testing.T) {
	ts, _, teardown := startHTTPServer()
	defer teardown()

	res, code := getRequest(t, ts.URL+"/ping", "")
	assert.Equal(t, "pong\n", res)
	assert.Equal(t, http.StatusOK, code)
}

func startHTTPServer() (ts *httptest.Server, rest *Rest, gracefulTeardown func()) {
	rest = &Rest{
		Version: "test",
		Service: &service.GeminiProxy{},
	}
	ts = httptest.NewServer(rest.routes())
	gracefulTeardown = func() {
		ts.Close()
	}
	return ts, rest, gracefulTeardown
}

func generateRndPort() (port int) {
	for i := 0; i < 10; i++ {
		port = 40000 + int(rand.Int31n(10000))
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			_ = ln.Close()
			break
		}
	}
	return port
}

func waitHTTPServer(port int) {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 1)
		conn, _ := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), time.Millisecond*10)
		if conn != nil {
			_ = conn.Close()
			break
		}
	}
}

func getRequest(t *testing.T, url string, rBody string) (data string, statusCode int) {
	var req *http.Request
	if rBody == "" {
		req, _ = http.NewRequest("GET", url, nil)
	} else {
		req, _ = http.NewRequest("GET", url, strings.NewReader(rBody))
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	return string(body), resp.StatusCode
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
