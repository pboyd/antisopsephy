package lgpn

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNames(t *testing.T) {
	assert := assert.New(t)
	server := testServer(t, "testfiles/names.json")
	client := newClient(server.URL, t.TempDir())

	// Read the names from the test server.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	nameCh, err := client.Names(ctx)
	var names []string
	if assert.NoError(err) {
		names = drainNameChannel(ctx, nameCh)
		assert.Greater(len(names), 100)
	}

	// Close the server, to verify that the next read comes from cache.
	server.Close()

	nameCh, err = client.Names(ctx)
	if assert.NoError(err) {
		cachedNames := drainNameChannel(ctx, nameCh)
		assert.NoError(err)
		assert.Equal(names, cachedNames)
	}
}

func testServer(t *testing.T, namesFile string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/lgpn_search.cgi":
			if r.URL.Query().Get("qtype") == "names" {
				fh, err := os.Open(namesFile)
				if err != nil {
					t.Fatalf("unable to open %s: %v", namesFile, err)
				}
				defer fh.Close()
				w.Header().Set("Content-Type", "application/json")
				io.Copy(w, fh)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	return httptest.NewServer(handler)
}

func drainNameChannel(ctx context.Context, ch <-chan string) []string {
	names := []string{}
	done := ctx.Done()
	for {
		select {
		case <-done:
			return nil
		case name, ok := <-ch:
			if !ok {
				return names
			}
			names = append(names, name)
		}
	}
}
