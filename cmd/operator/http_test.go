package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTP(t *testing.T) {
	li, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	go runHTTP(li)

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
