package raop

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

// based on: http://hassansin.github.io/Unit-Testing-http-client-in-Go

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestPlay(t *testing.T) {
	dc := newDacpClient("1.1.1.1", 333, "testID", "testActiveRemote")
	url := ""
	header := ""
	dc.httpClient = NewTestClient(func(req *http.Request) *http.Response {
		// capture passed in URL
		url = req.URL.String()
		header = req.Header.Get("Active-Remote")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
	dc.Play()
	expectedURL := "http://1.1.1.1:333/ctrl-int/1/play"
	expectedRemote := "testActiveRemote"
	if url != expectedURL {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedURL, url)
	}
	if header != expectedRemote {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedRemote, header)
	}
}

func TestPause(t *testing.T) {
	dc := newDacpClient("1.1.1.1", 333, "testID", "testActiveRemote")
	url := ""
	header := ""
	dc.httpClient = NewTestClient(func(req *http.Request) *http.Response {
		// capture passed in URL
		url = req.URL.String()
		header = req.Header.Get("Active-Remote")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
	dc.Pause()
	expectedURL := "http://1.1.1.1:333/ctrl-int/1/pause"
	expectedRemote := "testActiveRemote"
	if url != expectedURL {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedURL, url)
	}
	if header != expectedRemote {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedRemote, header)
	}
}

func TestPlayPause(t *testing.T) {
	dc := newDacpClient("1.1.1.1", 333, "testID", "testActiveRemote")
	url := ""
	header := ""
	dc.httpClient = NewTestClient(func(req *http.Request) *http.Response {
		// capture passed in URL
		url = req.URL.String()
		header = req.Header.Get("Active-Remote")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
	dc.PlayPause()
	expectedURL := "http://1.1.1.1:333/ctrl-int/1/playpause"
	expectedRemote := "testActiveRemote"
	if url != expectedURL {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedURL, url)
	}
	if header != expectedRemote {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedRemote, header)
	}
}

func TestStop(t *testing.T) {
	dc := newDacpClient("1.1.1.1", 333, "testID", "testActiveRemote")
	url := ""
	header := ""
	dc.httpClient = NewTestClient(func(req *http.Request) *http.Response {
		// capture passed in URL
		url = req.URL.String()
		header = req.Header.Get("Active-Remote")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
	dc.Stop()
	expectedURL := "http://1.1.1.1:333/ctrl-int/1/stop"
	expectedRemote := "testActiveRemote"
	if url != expectedURL {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedURL, url)
	}
	if header != expectedRemote {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedRemote, header)
	}
}

func TestNext(t *testing.T) {
	dc := newDacpClient("1.1.1.1", 333, "testID", "testActiveRemote")
	url := ""
	header := ""
	dc.httpClient = NewTestClient(func(req *http.Request) *http.Response {
		// capture passed in URL
		url = req.URL.String()
		header = req.Header.Get("Active-Remote")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
	dc.Next()
	expectedURL := "http://1.1.1.1:333/ctrl-int/1/nextitem"
	expectedRemote := "testActiveRemote"
	if url != expectedURL {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedURL, url)
	}
	if header != expectedRemote {
		t.Errorf("Expected: %s\r\n Received: %s\r\n", expectedRemote, header)
	}
}
