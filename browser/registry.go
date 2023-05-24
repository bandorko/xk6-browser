package browser

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/grafana/xk6-browser/env"
)

// pidRegistry keeps track of the launched browser process IDs.
type pidRegistry struct {
	mu  sync.RWMutex
	ids []int
}

// registerPid registers the launched browser process ID.
func (r *pidRegistry) registerPid(pid int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ids = append(r.ids, pid)
}

// Pids returns the launched browser process IDs.
func (r *pidRegistry) Pids() []int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pids := make([]int, len(r.ids))
	copy(pids, r.ids)

	return pids
}

// remoteRegistry contains the details of the remote web browsers.
// At the moment it's the WS URLs.
type remoteRegistry struct {
	isRemote bool
	wsURLs   []string
}

// newRemoteRegistry will create a new RemoteRegistry. This will
// parse the K6_BROWSER_WS_URL env var to retrieve the defined
// list of WS URLs.
//
// K6_BROWSER_WS_URL can be defined as a single WS URL or a
// comma separated list of URLs.
func newRemoteRegistry(envLookup env.LookupFunc) *remoteRegistry {
	r := &remoteRegistry{}

	wsURL, isRemote := envLookup("K6_BROWSER_WS_URL")
	if !isRemote {
		return r
	}

	if !strings.ContainsRune(wsURL, ',') {
		r.isRemote = true
		r.wsURLs = []string{wsURL}
		return r
	}

	// If last parts element is a void string,
	// because WS URL contained an ending comma,
	// remove it
	parts := strings.Split(wsURL, ",")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	r.isRemote = true
	r.wsURLs = parts

	return r
}

// newRemoteRegistry will create a new RemoteRegistry. This will
// parse the K6_BROWSER_WS_URL env var to retrieve the defined
// list of WS URLs from the scenarios object.
//
// TODO: Name will change at the end of the PR.
func newRemoteRegistryFromScenarios(envLookup env.LookupFunc) (*remoteRegistry, error) {
	r := &remoteRegistry{}

	scenariosJSON, isRemote := envLookup("K6_BROWSER_WS_URL")
	if !isRemote {
		return r, nil
	}

	var scenarios []struct {
		ID       string `json:"id"`
		Browsers []struct {
			Handle string `json:"handle"`
		} `json:"browsers"`
	}

	err := json.Unmarshal([]byte(scenariosJSON), &scenarios)
	if err != nil {
		return nil, fmt.Errorf("parsing K6_BROWSER_WS_URL: %w", err)
	}

	for _, s := range scenarios {
		for _, b := range s.Browsers {
			r.wsURLs = append(r.wsURLs, b.Handle)
			r.isRemote = true
		}
	}

	return r, nil
}

// isRemoteBrowser returns a WS URL and true when a WS URL is defined,
// otherwise it returns an empty string and false. If more than one
// WS URL was registered in newRemoteRegistry, a randomly chosen URL from
// the list in a round-robin fashion is selected and returned.
func (r *remoteRegistry) isRemoteBrowser() (string, bool) {
	if !r.isRemote {
		return "", false
	}

	// Choose a random WS URL from the provided list
	i, _ := rand.Int(rand.Reader, big.NewInt(int64(len(r.wsURLs))))
	wsURL := r.wsURLs[i.Int64()]

	return wsURL, true
}