package httputil

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
)

func StartTestServerWithResponseMap(stubResponses map[string]*http.Response) (
	serverUrl string,
	client *http.Client,
	closeFn func(),
	incomingRequestsFn func() []*http.Request,
) {
	requests := make([]*http.Request, 0)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		copiedReq, _ := CopyRequest(req)
		requests = append(requests, copiedReq)

		if resp, exists := stubResponses[req.URL.String()]; exists {
			rw.WriteHeader(resp.StatusCode)
			rw.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
			if resp.Body != nil {
				var body io.ReadCloser
				body, resp.Body, _ = DrainBody(resp.Body)
				_, _ = io.Copy(rw, body)
			}
			return
		}

		rw.WriteHeader(http.StatusNotFound)
	}))

	return server.URL,
		server.Client(),
		func() { server.Close() },
		func() []*http.Request { return requests }
}

func StartTestServerWithResponseList(stubResponses []*http.Response) (
	url string,
	client *http.Client,
	closeFn func(),
	incomingRequestsFn func() []*http.Request,
) {
	requests := make([]*http.Request, 0)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		copiedReq, _ := CopyRequest(req)
		requests = append(requests, copiedReq)

		if len(stubResponses) > 0 {
			resp := stubResponses[0]
			rw.WriteHeader(resp.StatusCode)
			rw.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
			if resp.Body != nil {
				_, _ = io.Copy(rw, resp.Body)
			}

			stubResponses = stubResponses[1:]
			return
		}

		rw.WriteHeader(http.StatusNotFound)
	}))

	return server.URL,
		server.Client(),
		func() { server.Close() },
		func() []*http.Request { return requests }
}

func AssertGetJsonRequest(
	t assert.TestingT,
	expectToken string,
	expectUrl string,
	expectHeaders map[string]string,
	actual *http.Request,
) bool {
	return AssertJsonRequestWithoutBody(t, http.MethodGet, expectToken, expectUrl, expectHeaders, actual)
}

func AssertPostJsonRequest(
	t assert.TestingT,
	expectToken string,
	expectUrl string,
	expectBody string,
	expectHeaders map[string]string,
	actual *http.Request,
) bool {
	if !AssertJsonRequestWithoutBody(t, http.MethodPost, expectToken, expectUrl, expectHeaders, actual) {
		return false
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(actual.Body); err != nil {
		return assert.Fail(t, "cannot read from actual request body")
	}
	if !assert.JSONEq(t, expectBody, buf.String()) {
		return false
	}
	return true
}

func AssertJsonRequestWithoutBody(
	t assert.TestingT,
	expectMethod string,
	expectToken string,
	expectUrl string,
	expectHeaders map[string]string,
	actual *http.Request,
) bool {
	if !assert.Equal(t, expectMethod, actual.Method) {
		return false
	}
	if expectToken != "" && !assert.Equal(t, "Bearer "+expectToken, actual.Header.Get("Authorization")) {
		return false
	}

	if !assert.Equal(t, "application/json", actual.Header.Get("Accept")) {
		return false
	}
	for header, val := range expectHeaders {
		if !assert.Equal(t, val, actual.Header.Get(header)) {
			return false
		}
	}
	if !assert.Equal(t, expectUrl, actual.URL.String()) {
		return false
	}
	return true
}

func ResponseFromFile(filePath string) *http.Response {
	content, _ := ioutil.ReadFile(filePath)
	reader := bytes.NewReader(content)

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(reader),
	}
}
