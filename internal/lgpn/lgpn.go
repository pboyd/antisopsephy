package lgpn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	// For sqlite3 support in database/sql

	_ "github.com/mattn/go-sqlite3"
)

var defaultClient = newClient("http://clas-lgpn2.classics.ox.ac.uk", "")

type nameRow struct {
	Name      string `json:"name"`
	NotBefore string `json:"notBefore"`
	NotAfter  string `json:"notAfter"`
}

// Names returns a channel that will receive every name in the Lexicon of Greek
// Personal Names (LGPN). If the names cannot be retrieved an error is
// returned.
//
// The returned channel will be closed after the last name has been read or
// when the passed context is closed.
func Names(ctx context.Context) (<-chan string, error) {
	return defaultClient.Names(ctx)
}

// client handles fetching and caching of names from the LGPN.
type client struct {
	lgpnBase *url.URL
	cacheDir string
}

// newClient returns a client which will connect to the LGPN using lgpnBase for
// the URL and cache results in cacheDir.
//
// lgpnBase should contain the scheme and hostname of the URL. For example,
// "http://clas-lgpn2.classics.ox.ac.uk" or "http://localhost:8080". If
// lgpnBase is an invalid URL newClient will panic.
//
// If cacheDir is an empty string a suitable directory will be selected based
// of the system platform. If the cacheDir does not exist an attempt will be
// made to create it. If there is any problem with the cacheDir newClient will
// panic.
func newClient(lgpnBase, cacheDir string) *client {
	u, err := url.Parse(lgpnBase)
	if err != nil {
		panic(fmt.Sprintf("invalid url %q: %v", lgpnBase, err))
	}

	if cacheDir == "" {
		cacheDir, err = os.UserCacheDir()
		if err != nil {
			panic(fmt.Sprintf("cannot find user cache dir: %v", err))
		}
		cacheDir = filepath.Join(cacheDir, "antisopsephy")
	}

	err = os.MkdirAll(cacheDir, 0777)
	if err != nil {
		panic(fmt.Sprintf("unable to create %s: %v", cacheDir, err))
	}

	return &client{
		lgpnBase: u,
		cacheDir: cacheDir,
	}
}

// Names is the internal implementation of the package-level Names function.
// See that function for the documentation.
func (c *client) Names(ctx context.Context) (<-chan string, error) {
	cache, err := newCache(ctx, c.cacheDir)
	if err != nil {
		return nil, err
	}
	defer cache.Close()

	count, err := cache.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to count names: %w", err)
	}
	if count == 0 {
		fmt.Fprint(os.Stderr, "Downloading name list and building cache (this may take a minute)...")

		rows, err := c.downloadNames(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to download names: %w", err)
		}

		err = cache.Populate(ctx, rows)
		if err != nil {
			return nil, fmt.Errorf("unable to populate name cache: %w", err)
		}
		fmt.Fprintln(os.Stderr, " done.")
	}

	return cache.Names(ctx)
}

// downloadNames retrieves the JSON file of all name from LGPN (or the host in
// the lgpnBase URL), parses it and returns each entry by way of the channel.
// The mechanics of the channel work exactly like Names().
func (c *client) downloadNames(ctx context.Context) (<-chan nameRow, error) {
	u := c.lgpnBase.JoinPath("/cgi-bin/lgpn_search.cgi")
	u.RawQuery = "qtype=names"

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("internal error: an invalid request was generated: %w", err)
	}

	req.Header.Set("User-Agent", "github.com/pboyd/antisopsephy")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make request: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("LGPN returned unexpected status: %s", res.Status)
	}

	names := make(chan nameRow)
	go func() {
		defer res.Body.Close()
		defer close(names)

		dec := json.NewDecoder(res.Body)

		first, err := dec.Token()
		if err != nil {
			fmt.Fprintf(os.Stderr, "json error: %v\n", err)
			return
		}
		if first != json.Delim('[') {
			fmt.Fprintf(os.Stderr, "json error: unexpected token %q\n", first)
			return
		}

		for dec.More() {
			var row nameRow
			err = dec.Decode(&row)
			if err != nil {
				// The JSON we get back has an invalid comma after the last item, just ignore that error.
				if strings.Contains(err.Error(), "invalid character ']'") {
					continue
				}

				fmt.Fprintf(os.Stderr, "json error: %v\n", err)
				return
			}

			names <- row
		}
	}()

	return names, nil
}
