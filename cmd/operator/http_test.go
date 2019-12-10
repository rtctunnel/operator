package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	li := startTestServer(t)
	defer li.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := http.Get(fmt.Sprintf("http://%s/pub?%s", li.Addr(), url.Values{
			"data":    {"Hello World"},
			"address": {"test"},
		}.Encode()))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := http.Get(fmt.Sprintf("http://%s/sub?%s", li.Addr(), url.Values{
			"address": {"test"},
		}.Encode()))
		assert.NoError(t, err)
		defer resp.Body.Close()

		bs, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Hello World", string(bs))
	}()

	wg.Wait()
}

func TestPost(t *testing.T) {
	li := startTestServer(t)
	defer li.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := http.PostForm(fmt.Sprintf("http://%s/pub", li.Addr()), url.Values{
			"data":    {"Hello World"},
			"address": {"test"},
		})
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := http.Get(fmt.Sprintf("http://%s/sub?%s", li.Addr(), url.Values{
			"address": {"test"},
		}.Encode()))
		assert.NoError(t, err)
		defer resp.Body.Close()

		bs, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Hello World", string(bs))
	}()

	wg.Wait()
}

func TestHTTPSizeLimit(t *testing.T) {
	li := startTestServer(t)
	defer li.Close()

	resp, err := http.Get(fmt.Sprintf("http://%s/pub?%s", li.Addr(), url.Values{
		"data":    {strings.Repeat("x", maxMessageSize+1)},
		"address": {"test"},
	}.Encode()))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHTTPTimeout(t *testing.T) {
	originalTimeout := timeout
	defer func() {
		timeout = originalTimeout
	}()
	timeout = time.Millisecond

	li := startTestServer(t)
	defer li.Close()

	for _, endpoint := range []string{"pub", "sub"} {
		resp, err := http.Get(fmt.Sprintf("http://%s/%s?%s", li.Addr(), endpoint, url.Values{
			"data":    {"x"},
			"address": {"test"},
		}.Encode()))
		assert.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusGatewayTimeout, resp.StatusCode)
	}
}

func TestHead(t *testing.T) {
	li := startTestServer(t)
	defer li.Close()

	req, err := http.NewRequest("HEAD", fmt.Sprintf("http://%s", li.Addr()), nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestHTTPCORS(t *testing.T) {
	li := startTestServer(t)
	defer li.Close()

	req, err := http.NewRequest("OPTIONS", fmt.Sprintf("http://%s", li.Addr()), nil)
	assert.NoError(t, err)
	req.Header.Add("Origin", "https://www.example.com")
	req.Header.Add("Access-Control-Request-Method", "PUT")

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "PUT", res.Header.Get("Access-Control-Allow-Methods"))
}

func startTestServer(t *testing.T) net.Listener {
	li, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	go runHTTP(li)

	return li
}
