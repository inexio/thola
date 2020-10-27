package network

import (
	"github.com/inexio/thola/core/tholaerr"
	"sync"
)

/*
requestCache stores all sent requests and its results.
*/
type requestCache struct {
	sync.RWMutex

	cache map[string]cachedRequestResult
}

func newRequestCache() requestCache {
	return requestCache{
		cache: make(map[string]cachedRequestResult),
	}
}

func (r *requestCache) add(identifier string, result interface{}, err error) {
	r.Lock()
	defer r.Unlock()
	(r.cache)[identifier] = cachedRequestResult{res: result, err: err}
}

// getSuccessfulRequests returns all successful request that do not contain an error.
func (r *requestCache) getSuccessfulRequests() map[string]cachedRequestResult {
	m := make(map[string]cachedRequestResult)
	r.RLock()
	defer r.RUnlock()
	for k, v := range r.cache {
		if !v.returnedError() {
			m[k] = v
		}
	}
	return m
}

// get returns the cache entry for an identifier, an error if its not in the cache
func (r *requestCache) get(identifier string) (cachedRequestResult, error) {
	r.RLock()
	defer r.RUnlock()
	if v, ok := (r.cache)[identifier]; ok {
		return v, nil
	}
	return cachedRequestResult{}, tholaerr.NewNotFoundError("no cache entry for this value")
}

type cachedRequestResult struct {
	err error
	res interface{}
}

func (d *cachedRequestResult) returnedError() bool {
	return d.err != nil
}
