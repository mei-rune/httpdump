package httpdump

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
)

var http11 = []byte("HTTP/")

func ReadFile(filename string) (*http.Request, *http.Response, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	return Read(bs)
}

func Read(bs []byte) (*http.Request, *http.Response, error) {
	idx1 := bytes.Index(bs, http11)
	if idx1 < 0 {
		return nil, nil, errors.New("Request HTTP/ isnot found")
	}
	idx2 := bytes.Index(bs[idx1+len(http11):], http11)
	if idx2 < 0 {
		return nil, nil, errors.New("Response HTTP/ isnot found")
	}

	bs1 := bs[:idx1+len(http11)+idx2]
	bs2 := bs[idx1+len(http11)+idx2:]

	reqbuf := bufio.NewReader(bytes.NewReader(bs1))
	req, err := http.ReadRequest(reqbuf)
	if err != nil {
		return nil, nil, errors.New("read request error:" + err.Error())
	}

	if req.Method == "GET" || req.Method == "POST" {
		if req.ContentLength <= 0 {
			req.Body = ioutil.NopCloser(reqbuf)
		}
	}

	var respbuf = bufio.NewReader(bytes.NewReader(bs2))
	resp, err := http.ReadResponse(respbuf, req)
	if err != nil {
		return nil, nil, errors.New("read response error:" + err.Error())
	}
	if resp.ContentLength <= 0 {
		resp.Body = ioutil.NopCloser(respbuf)
	}
	return req, resp, nil
}

type ReqestResponse struct {
	Request       *http.Request
	Response      *http.Response
	RequestBytes  []byte
	ResponseBytes []byte
}

func ReadData(dir string) (map[string][]ReqestResponse, error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var data = map[string][]ReqestResponse{}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		
		req, resp, err := ReadFile(filepath.Join(dir, fi.Name()))
		if err != nil {
			return nil, err
		}

		old := data[req.URL.Path]
		var requestBytes []byte
		if req.Method == "POST" || req.Method == "PUT" {
			requestBytes, _ = ioutil.ReadAll(req.Body)
		}
		responseBytes, _ := ioutil.ReadAll(resp.Body)
		data[req.URL.Path] = append(old, ReqestResponse{
			Request:       req,
			Response:      resp,
			RequestBytes:  requestBytes,
			ResponseBytes: responseBytes,
		})
	}
	return data, nil
}

func ReadDir(dir string) (http.Handler, error) {
	data, err := ReadData(dir)
	if err != nil {
		return nil, err
	}
	return ToHandler(data), nil
}

func ToHandler(data map[string][]ReqestResponse) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseList := data[r.URL.Path]
		if len(responseList) == 0 {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return
		}

		values := r.URL.Query()
		for idx := range responseList {
			resp := responseList[idx]

			if queryEqual(values, resp.Request.URL.Query()) {
				if r.Method == "POST" || r.Method == "PUT" {
					requestBytes, _ := ioutil.ReadAll(r.Body)
					if !bodyEqual(requestBytes, resp.RequestBytes) {
						http.Error(w, "request body isnot same for "+r.Method, http.StatusBadRequest)
						return
					}
				}

				for _, name := range []string {
					"Content-Type",
					"Server",
					"X-Content-Type-Options",
					"X-Frame-Options",
					"X-Xss-Protection",
				}{
					ss, ok := resp.Response.Header[name]
					if ok {
						w.Header()[name] = ss
					}
				}

				w.WriteHeader(resp.Response.StatusCode)
				io.Copy(w, bytes.NewReader(resp.ResponseBytes))
				return
			}
		}

		fmt.Println("=====")
		fmt.Println(values)
		fmt.Println("-----")
		for idx := range responseList {
			resp := responseList[idx]
			fmt.Println(resp.Request.URL.Query())
		}
		http.Error(w, "404 url query isnot match", http.StatusNotFound)
	})
}

func queryEqual(a, b url.Values) bool {
	for key, values := range a {
		bvalues, ok := b[key]
		if !ok {
			if ignoreValue(key, values) {
				continue
			}
			return false
		}

		if !equalValues(values, bvalues) {
			return false
		}
	}

	return true
}

func equalValues(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for idx := range a {
		if a[idx] != b[idx] {
			return false
		}
	}
	return true
}

func ignoreValue(key string, values []string) bool {
	return false
}

func bodyEqual(a, b []byte) bool {
	a = bytes.TrimSpace(a)
	b = bytes.TrimSpace(b)
	if bytes.Equal(a, b) {
		return true
	}
	var aj, bj interface{}

	ad := json.NewDecoder(bytes.NewReader(a))
	ad.UseNumber()
	err := ad.Decode(&aj)
	if err != nil {
		fmt.Println(err)
		return false
	}

	bd := json.NewDecoder(bytes.NewReader(b))
	bd.UseNumber()
	err = bd.Decode(&bj)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return reflect.DeepEqual(aj, bj)
}
