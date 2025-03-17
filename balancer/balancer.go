package balancer

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var CipherSuites = []uint16{
	// TLS 1.0 - 1.2 cipher suites
	tls.TLS_RSA_WITH_RC4_128_SHA,
	tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	// TLS 1.3 cipher suites
	tls.TLS_AES_128_GCM_SHA256,
	tls.TLS_AES_256_GCM_SHA384,
	tls.TLS_CHACHA20_POLY1305_SHA256,
}

const (
	NoAvailableServersNotify = "no available servers"
	ChangedUrlNotify         = "url changed"
)

var (
	ErrNoAvailableServers = errors.New(NoAvailableServersNotify)
)

type Client struct {
	mu         sync.RWMutex
	urls       []string
	currentIdx int
	listeners  []chan string
	httpClient *http.Client
}

func NewClient(urls []string) (*Client, error) {
	if len(urls) == 0 {
		return nil, errors.New("no urls provided")
	}
	for i, u := range urls {
		if u == "" {
			return nil, errors.New("empty balancer url")
		}
		if !strings.HasSuffix(u, "/") {
			urls[i] += "/"
		}
	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{
		CipherSuites:       CipherSuites,
		InsecureSkipVerify: true,
	}

	c := &Client{
		urls:       urls,
		currentIdx: -1,
		listeners:  []chan string{},
		httpClient: &http.Client{
			Transport: customTransport,
			Timeout:   10 * time.Second,
		},
	}

	if err := c.findWorkingServer(); err != nil {
		return nil, err
	}

	go c.monitorServers()

	return c, nil
}

func (c *Client) findWorkingServer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	resultCh := make(chan int, 1)
	var once sync.Once

	c.mu.RLock()
	urls := c.urls
	c.mu.RUnlock()

	for i := range urls {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if c.checkServerAvailability(ctx, idx) {
				once.Do(func() {
					resultCh <- idx
					cancel()
				})
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	idx, ok := <-resultCh
	if !ok {
		return ErrNoAvailableServers
	}

	c.mu.Lock()
	c.currentIdx = idx
	c.mu.Unlock()
	return nil
}

func (c *Client) checkServerAvailability(ctx context.Context, index int) bool {
	c.mu.RLock()
	if index < 0 || index >= len(c.urls) {
		c.mu.RUnlock()
		return false
	}
	urlToCheck := c.urls[index]
	c.mu.RUnlock()

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, urlToCheck, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return true
}

func (c *Client) monitorServers() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.RLock()
		idx := c.currentIdx
		c.mu.RUnlock()

		if idx < 0 || idx >= len(c.urls) {
			if err := c.findWorkingServer(); err != nil {
				c.notifyFailure()
			} else {
				c.notifyUrlChanged()
			}
			continue
		}

		if !c.checkServerAvailability(context.Background(), idx) {
			oldIdx := idx
			if !c.switchToNextServer() {
				c.notifyFailure()
				continue
			}
			if c.currentIdx != oldIdx {
				c.notifyUrlChanged()
			}
		}
	}
}

func (c *Client) switchToNextServer() bool {
	err := c.findWorkingServer()
	if err != nil {
		c.mu.Lock()
		c.currentIdx = -1
		c.mu.Unlock()
		return false
	}
	return true
}

func (c *Client) SendRequest(originalReq *http.Request, skip bool) (*http.Response, error) {
	var bodyBytes []byte
	if originalReq.Body != nil {
		b, err := io.ReadAll(originalReq.Body)
		if err != nil {
			return nil, err
		}
		_ = originalReq.Body.Close()
		bodyBytes = b
	}
	return c.doRequestWithRetry(originalReq, bodyBytes, skip)
}

func (c *Client) doRequestWithRetry(originalReq *http.Request, bodyBytes []byte, skip bool) (*http.Response, error) {
	c.mu.RLock()
	idx := c.currentIdx
	c.mu.RUnlock()

	if idx < 0 || idx >= len(c.urls) {
		return nil, ErrNoAvailableServers
	}

	reqCopy := cloneRequest(originalReq, c.urls[idx], bodyBytes)

	resp, err := c.httpClient.Do(reqCopy)
	if err != nil {
		if !skip && c.switchToNextServer() {
			return c.doRequestWithRetry(originalReq, bodyBytes, true)
		}
		return nil, err
	}
	return resp, nil
}

func cloneRequest(originalReq *http.Request, baseURL string, bodyBytes []byte) *http.Request {
	srvURL, _ := url.Parse(strings.TrimRight(baseURL, "/"))

	reqCopy := originalReq.Clone(originalReq.Context())

	newURL := *reqCopy.URL
	newURL.Scheme = srvURL.Scheme
	newURL.Host = srvURL.Host

	if srvURL.Path != "" || newURL.Path != "" {
		newURL.Path = strings.TrimRight(srvURL.Path, "/") + "/" + strings.TrimLeft(newURL.Path, "/")
	}

	reqCopy.URL = &newURL

	if len(bodyBytes) > 0 {
		reqCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		reqCopy.Body = nil
	}
	return reqCopy
}

func (c *Client) GetCurrentURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.currentIdx >= 0 && c.currentIdx < len(c.urls) {
		return c.urls[c.currentIdx]
	}
	return ""
}

func (c *Client) SwitchToNextServer() bool {
	return c.switchToNextServer()
}

func (c *Client) RegisterListener(ch chan string) {
	c.listeners = append(c.listeners, ch)
}

func (c *Client) notifyUrlChanged() {
	for _, ch := range c.listeners {
		ch <- ChangedUrlNotify
	}
}

func (c *Client) notifyFailure() {
	for _, ch := range c.listeners {
		ch <- NoAvailableServersNotify
	}
}
