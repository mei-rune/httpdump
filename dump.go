package httpdump

import (
	"io"
	"net/http"
	"net/http/httputil"
)

type DebugProvider interface {
	New() (io.WriteCloser, error)

	NewBody(request *http.Request) (io.WriteCloser, error)
}

var dumpTo DebugProvider

func SetDebugProvider(ddp DebugProvider) {
	dumpTo = ddp
}

func Do(client *http.Client, request *http.Request) (*http.Response, error) {

	var w io.WriteCloser
	var needClose = true
	if dumpTo != nil {
		w, _ = dumpTo.New()
		if w != nil {
			defer func() {
				if needClose {
					w.Close()
				}
			}()

			bs, err := httputil.DumpRequest(request, false)
			if err != nil {
				io.WriteString(w, err.Error())
				io.WriteString(w, "\r\n\r\n")
			} else {
				_, err = w.Write(bs)
				if err != nil {
					io.WriteString(w, err.Error())
					io.WriteString(w, "\r\n\r\n")
				}
			}

			if request.Body != nil {
				request.Body = &teeReader{
					r: request.Body,
					w: w,
				}
			}
		}
	}

	response, err := client.Do(request)

	if w != nil {
		if err != nil {
			io.WriteString(w, err.Error())
			io.WriteString(w, "\r\n\r\n")
		} else {
			bs, e := httputil.DumpResponse(response, false)
			if e != nil {
				io.WriteString(w, e.Error())
				io.WriteString(w, "\r\n\r\n")
			} else {
				_, e = w.Write(bs)
				if e != nil {
					io.WriteString(w, e.Error())
					io.WriteString(w, "\r\n\r\n")
				}
			}
			if response.Body != nil {
				responseWriter, _ := dumpTo.NewBody(request)

				needClose = false
				response.Body = &teeReader2{
					r:  response.Body,
					w1: w,
					w2: responseWriter,
				}
			}
		}
	}
	return response, err
}

type teeReader struct {
	r io.ReadCloser
	w io.Writer
}

func (t *teeReader) Close() error {
	return t.r.Close()
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

type teeReader2 struct {
	r  io.ReadCloser
	w1 io.WriteCloser
	w2 io.WriteCloser
}

func (t *teeReader2) Close() error {
	if t.w1 != nil {
		t.w1.Close()
	}
	if t.w2 != nil {
		t.w2.Close()
	}
	return t.r.Close()
}

func (t *teeReader2) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if t.w1 != nil {
			t.w1.Write(p[:n])
		}
		if t.w2 != nil {
			t.w2.Write(p[:n])
		}
	}
	return n, err
}
