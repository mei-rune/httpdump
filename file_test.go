package httpdump

import (
	"bytes"
	"io/ioutil"
	"testing"
)

const text1 = `GET /api/vx HTTP/1.1
Host: www.abc.com
Accept: application/json
Authorization: Basic xxxxx=
User-Agent: 

HTTP/1.1 200 OK
Transfer-Encoding: chunked
Content-Type: application/json;charset=utf-8
X-Xss-Protection: 1; mode=block

{
  "items" : []
}`

const responseText1 = `{
  "items" : []
}`

func TestRead1(t *testing.T) {
	_, resp, err := Read([]byte(text1))
	if err != nil {
		t.Error(err)
		return
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	if len(bs) == 0 {
		t.Error("empty")
		return
	}

	if !bytes.Equal(bs, []byte(responseText1)) {
		t.Error("want:", responseText1)
		t.Error(" got:", string(bs))
	}
}

const text2 = `POST /rest/ HTTP/1.1
Host: www.abc.com
Accept-Language: zh_CN
Content-Type: application/json;charset=UTF-8
X-Auth-Token: x-xxx

[2814]
HTTP/1.1 200 OK
Transfer-Encoding: chunked
X-Request-Id: 123

{"status_code":200,"error_code":0}`

const requestText2 = `[2814]`
const responseText2 = `{"status_code":200,"error_code":0}`

func TestRead2(t *testing.T) {
	req, resp, err := Read([]byte(text2))
	if err != nil {
		t.Error(err)
		return
	}
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Error(err)
		return
	}
	if len(bs) == 0 {
		t.Error("empty")
		return
	}

	if !bytes.Equal(bytes.TrimSpace(bs), []byte(requestText2)) {
		t.Error("want:", requestText2)
		t.Error(" got:", string(bs))
	}

	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	if len(bs) == 0 {
		t.Error("empty")
		return
	}

	if !bytes.Equal(bs, []byte(responseText2)) {
		t.Error("want:", responseText2)
		t.Error(" got:", string(bs))
	}
}
