package cacheFlex

import (
	"distributed_cache/cacheFlex/consistentHash"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_cacheFlex/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	self        string     // this peer's base URL, e.g. "https://example.net:8000"
	basePath    string     // "_cacheFelx" in our case
	mu          sync.Mutex // guards peers and httpGetters
	peers       *consistentHash.Map
	httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

/*
	v ...interface{} is a variadic parameter that allows you to pass a variable
*/

// Log info: server name + HTTP Method + URL path
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())

	/*
		In Go, if you do not explicitly write a "return" statement at the end
		of a function, the function will still return when it reaches the end
		of its execution.
	*/
}

type httpGetter struct {
	baseURL string
}

// implements interface PeerPicker
func (h *httpGetter) Get(group string, key string) ([]byte, error) {

	// use QueryEscape to escape the string to ensure it's safe to include in a URL
	u := fmt.Sprintf(
		"%v%v%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	// schedule the closing of the response body ('res.body') to occur
	// when the current funtion ('Get') returns. This ensures that
	// the response body is closed property to prevent resource leaks.
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	// read and response body using 'ioutil.ReadAll' and stores the content
	// in the 'bytes' variable. Any error encountered during the reading operation
	// is also captured
	bytes, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, err
}

/*
a compile-time check to make sure that the 'httpGetter' type implements the 'PeerGetter' interface
(*httpGetter): this part represent the type assertion or type convertion. It's specifying that we are

	creating a value of type (*httpGetter), which is a pointer to the httpGetter type

(nil):         this part represent the value being assigned to the newly created type. In this case,

	it's 'nil', which means the pointer to 'httpGetter' is set to a nil (empty) value
*/
var _ PeerGetter = (*httpGetter)(nil)

// implements interface PeerPicker
// updates the pool's list of peers
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistentHash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))

	// create a HTTP client 'httpGetter' for every node
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// picks a peer according to the key
// uses the consistent hash algorithm to select a specific peer node based on a provided key an
// returns the corresponding HTTP cilent for communication with that node
func (p *HTTPPool) PickPeer(key string) (peer PeerGetter, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
