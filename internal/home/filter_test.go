package home

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testStartFilterListener() net.Listener {
	mux := http.NewServeMux()
	mux.HandleFunc("/filters/1.txt", func(w http.ResponseWriter, r *http.Request) {
		content := `||example.org^$third-party
# Inline comment example
||example.com^$third-party
0.0.0.0 example.com
`
		_, _ = w.Write([]byte(content))
	})

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	go func() { _ = http.Serve(listener, mux) }()
	return listener
}

func TestFilters(t *testing.T) {
	l := testStartFilterListener()
	defer func() { _ = l.Close() }()

	dir := prepareTestDir()
	defer func() { _ = os.RemoveAll(dir) }()
	Context = homeContext{}
	Context.workDir = dir
	Context.client = &http.Client{
		Timeout: 5 * time.Second,
	}
	Context.filters.Init()

	f := filter{
		URL: fmt.Sprintf("http://127.0.0.1:%d/filters/1.txt", l.Addr().(*net.TCPAddr).Port),
	}

	// download
	ok, err := Context.filters.update(&f)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, 3, f.RulesCount)

	// refresh
	ok, err = Context.filters.update(&f)
	assert.False(t, ok)
	assert.Nil(t, err)

	err = Context.filters.load(&f)
	assert.Nil(t, err)

	f.unload()
	_ = os.Remove(f.Path())
}
