package request

import (
	"net/http"
	"net/url"
	"net"
	"time"
	"crypto/tls"
	"strings"
	"io"
	"github.com/satori/go.uuid"
	"errors"
	"io/ioutil"
	"fmt"
)

type chain struct {
	StartTime int64         `json:"start_time"`
	EndTime   int64         `json:"end_time"`
	Url       string        `json:"url"`
	WorkerId  string        `json:"worker_id"`
	Time      time.Duration `json:"time"`
}

func Post(url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {
	return send("POST", url, strings.NewReader(params.Encode()), headers, proxyUrl)
}

func Get(url string, params url.Values, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {
	if params != nil {
		url += "?" + params.Encode()
	}
	return send("GET", url, nil, headers, proxyUrl)
}

func send(method string, httpUrl string, params io.Reader, headers map[string]string, proxyUrl string) (body []byte, err error, trans chain) {

	startTime := time.Now().UnixNano()

	var ch chain

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

	uid, _ := uuid.NewV4()

	workerId := uid.String()

	ch.WorkerId = workerId

	request.Header.Set("Worker-Id", workerId)
	request.Header.Set("Token-Id", "X")
	request.Header.Set("Device-Id", "x.yirendai.com")
	request.Header.Set("Client-System", "X")
	request.Header.Set("Client-Time", time.Unix(0, time.Now().UnixNano()).Format("2006-01-02 15:04:05.999999"))

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

