package request

import (
	"time"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"io"
	"net"
	"crypto/tls"
	"net/http"
	"errors"
)

type Request interface {
	Post(url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain)
	Get(url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain)
}

type chain struct {
	StartTime int64         `json:"start_time"`
	EndTime   int64         `json:"end_time"`
	Url       string        `json:"url"`
	WorkerId  string        `json:"worker_id"`
	Time      time.Duration `json:"time"`
}

type request struct{}

func NewRequest() (r *request) {
	return &request{}
}

func (r *request) Send(method string, url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {

	var data string

	if strings.ToUpper(method) == "GET" {
		if params != nil {
			url += "?" + params.Encode()
		}
	} else {
		if len(params.Get("")) > 0 {
			data = params.Get("")
		} else {
			data = params.Encode()
		}
	}

	return send(method, url, strings.NewReader(data), headers, proxyUrl)
}

func (r *request) Post(url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {
	var data string
	if len(params.Get("")) > 0 {
		data = params.Get("")
	} else {
		data = params.Encode()
	}
	return send("POST", url, strings.NewReader(data), headers, proxyUrl)
}

func (r *request) Patch(url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {
	var data string
	if len(params.Get("")) > 0 {
		data = params.Get("")
	} else {
		data = params.Encode()
	}
	return send("PATCH", url, strings.NewReader(data), headers, proxyUrl)
}

func (r *request) Get(url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {
	if params != nil {
		url += "?" + params.Encode()
	}
	return send("GET", url, nil, headers, proxyUrl)
}

func (r *request) Put(url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {
	var data string
	if len(params.Get("")) > 0 {
		data = params.Get("")
	} else {
		data = params.Encode()
	}
	return send("PUT", url, strings.NewReader(data), headers, proxyUrl)
}

func send(method string, httpUrl string, params io.Reader, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {

	startTime := time.Now().UnixNano()

	var ch chain

	ch.Url = httpUrl
	ch.StartTime = startTime

	var res *http.Response
	var proxy func(r *http.Request) (*url.URL, error)

	if proxyUrl != "" {
		proxy = func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyUrl)
		}
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(1 * int64(time.Second)),
		KeepAlive: time.Duration(1 * int64(time.Second)),
	}

	var isHttps bool
	if strings.Index(httpUrl, "https") != -1 {
		isHttps = true
	}

	transport := &http.Transport{
		Proxy: proxy, DialContext: dialer.DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: isHttps,
		},
	}
	client := &http.Client{
		Transport: transport,
	}
	request, err := http.NewRequest(strings.ToUpper(method), httpUrl, params)

	if err != nil {
		ch.Url = httpUrl
		return
	}

	if headers != nil {
		for k, v := range headers {
			request.Header.Set(k, v)
		}
	} else {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	request.Header.Set("Client-Time", time.Unix(0, time.Now().UnixNano()).Format("2006-01-02 15:04:05.999999"))
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.162 Safari/537.36")

	res, err = client.Do(request)
	if err != nil {
		return
	}

	defer res.Body.Close()

	endTime := time.Now().UnixNano()

	ch.EndTime = endTime
	ch.Time = time.Duration(endTime - startTime)

	if res.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("http status code %d", res.StatusCode))
	}

	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}

	return
}

