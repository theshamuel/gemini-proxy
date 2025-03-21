package cmd

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"io"
	"math/rand"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServerApp(t *testing.T) {
	app, ctx, cancel := buildListCmdOpts(t, func(s ServerCmd) ServerCmd {
		return s
	})

	go func() { _ = app.run(ctx) }()
	waitHTTPServer(app.Port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", app.Port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong\n", string(body))

	cancel()
	app.Wait()
}

func createAppFromCmd(t *testing.T, cmd ServerCmd) (*application, context.Context, context.CancelFunc) {
	app := cmd.bootstrapApp()

	ctx, cancel := context.WithCancel(context.Background())
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return app, ctx, cancel
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

func buildListCmdOpts(t *testing.T, fn func(o ServerCmd) ServerCmd) (*application, context.Context, context.CancelFunc) {
	cmd := ServerCmd{}
	p := flags.NewParser(&cmd, flags.Default)
	_, err := p.ParseArgs([]string{"--port=4356"})
	require.NoError(t, err)
	cmd = fn(cmd)
	return createAppFromCmd(t, cmd)
}

func TestMain(m *testing.M) {
	//Unknown reasons
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("net/http.(*Server).Shutdown"))
}
