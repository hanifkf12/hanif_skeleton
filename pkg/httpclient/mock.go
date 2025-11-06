package httpclient

import (
	"context"
	"sync"
)

// MockClient is a mock implementation of HTTPClient for testing
type MockClient struct {
	mu              sync.RWMutex
	getResponses    map[string]*Response
	postResponses   map[string]*Response
	putResponses    map[string]*Response
	patchResponses  map[string]*Response
	deleteResponses map[string]*Response
	errors          map[string]error
}

// NewMockClient creates a new mock HTTP client
func NewMockClient() *MockClient {
	return &MockClient{
		getResponses:    make(map[string]*Response),
		postResponses:   make(map[string]*Response),
		putResponses:    make(map[string]*Response),
		patchResponses:  make(map[string]*Response),
		deleteResponses: make(map[string]*Response),
		errors:          make(map[string]error),
	}
}

// OnGet sets up mock response for GET request
func (m *MockClient) OnGet(url string, response *Response, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.errors[url] = err
	} else {
		m.getResponses[url] = response
	}
}

// OnPost sets up mock response for POST request
func (m *MockClient) OnPost(url string, response *Response, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.errors[url] = err
	} else {
		m.postResponses[url] = response
	}
}

// OnPut sets up mock response for PUT request
func (m *MockClient) OnPut(url string, response *Response, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.errors[url] = err
	} else {
		m.putResponses[url] = response
	}
}

// OnPatch sets up mock response for PATCH request
func (m *MockClient) OnPatch(url string, response *Response, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.errors[url] = err
	} else {
		m.patchResponses[url] = response
	}
}

// OnDelete sets up mock response for DELETE request
func (m *MockClient) OnDelete(url string, response *Response, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.errors[url] = err
	} else {
		m.deleteResponses[url] = response
	}
}

// Get returns mocked GET response
func (m *MockClient) Get(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[url]; ok {
		return nil, err
	}

	if resp, ok := m.getResponses[url]; ok {
		return resp, nil
	}

	return &Response{StatusCode: 200, Body: []byte("{}")}, nil
}

// Post returns mocked POST response
func (m *MockClient) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[url]; ok {
		return nil, err
	}

	if resp, ok := m.postResponses[url]; ok {
		return resp, nil
	}

	return &Response{StatusCode: 200, Body: []byte("{}")}, nil
}

// Put returns mocked PUT response
func (m *MockClient) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[url]; ok {
		return nil, err
	}

	if resp, ok := m.putResponses[url]; ok {
		return resp, nil
	}

	return &Response{StatusCode: 200, Body: []byte("{}")}, nil
}

// Patch returns mocked PATCH response
func (m *MockClient) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[url]; ok {
		return nil, err
	}

	if resp, ok := m.patchResponses[url]; ok {
		return resp, nil
	}

	return &Response{StatusCode: 200, Body: []byte("{}")}, nil
}

// Delete returns mocked DELETE response
func (m *MockClient) Delete(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[url]; ok {
		return nil, err
	}

	if resp, ok := m.deleteResponses[url]; ok {
		return resp, nil
	}

	return &Response{StatusCode: 200, Body: []byte("{}")}, nil
}

// Do returns mocked response based on method and URL
func (m *MockClient) Do(ctx context.Context, req *Request) (*Response, error) {
	switch req.Method {
	case "GET":
		return m.Get(ctx, req.URL, req.Headers)
	case "POST":
		return m.Post(ctx, req.URL, req.Body, req.Headers)
	case "PUT":
		return m.Put(ctx, req.URL, req.Body, req.Headers)
	case "PATCH":
		return m.Patch(ctx, req.URL, req.Body, req.Headers)
	case "DELETE":
		return m.Delete(ctx, req.URL, req.Headers)
	default:
		return &Response{StatusCode: 200, Body: []byte("{}")}, nil
	}
}
