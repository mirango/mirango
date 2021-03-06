package mirango

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/mirango/framework"
)

type Response struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
	encoding      string
	encoders      framework.Encoders
}

func NewResponse(w http.ResponseWriter, encoders framework.Encoders) *Response {
	return &Response{
		ResponseWriter: w,
		encoders:       encoders,
	}
}

func (w *Response) WriteEntity(status int, value interface{}) error {
	_, err := WriteEntity(w, status, value, w.encoding)
	return err
}

func (w *Response) WriteAsXml(status int, value interface{}, writeHeader bool) error {
	_, err := WriteAsXml(w, status, value, writeHeader)
	return err
}

func (w *Response) WriteAsJson(status int, value interface{}) error {
	_, err := WriteAsJson(w, status, value)
	return err
}

func (w *Response) WriteJson(status int, value interface{}, contentType string) error {
	_, err := WriteJson(w, status, value, contentType)
	return err
}

func (w *Response) WriteString(status int, str string) error {
	_, err := WriteString(w, status, str)
	return err
}

func (w *Response) WriteHeader(httpStatus int) {
	if w.statusCode == 0 {
		if httpStatus == 0 {
			httpStatus = http.StatusOK
		}
		w.statusCode = httpStatus
		w.ResponseWriter.WriteHeader(httpStatus)
	}
}

func (w *Response) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func (w *Response) Write(bytes []byte) (int, error) {
	written, err := w.ResponseWriter.Write(bytes)
	w.contentLength += written
	return written, err
}

func (w *Response) ContentLength() int {
	return w.contentLength
}

func (w *Response) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *Response) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *Response) StreamAsJson(status int, d time.Duration, f func(int64) (interface{}, bool)) error {
	var i int64 = 0
	var err error
	for {
		e, stop := f(i)
		if e != nil {
			eerr := w.WriteJson(status, e, framework.MIME_JSON)
			if eerr != nil {
				err = eerr
				break
			}
			_, eerr = w.Write([]byte("\n"))
			w.Flush()
			i++
		}
		if stop {
			break
		}
		time.Sleep(d)
	}

	return err
}

func (w *Response) StreamAsXml(status int, d time.Duration, f func(int64) (interface{}, bool)) error {
	var i int64 = 0
	var err error
	for {
		e, stop := f(i)
		if e != nil {
			eerr := w.WriteAsXml(status, e, i == 0)
			if eerr != nil {
				err = eerr
				break
			}
			_, eerr = w.Write([]byte("\n"))
			w.Flush()
			i++
		}
		if stop {
			break
		}
		time.Sleep(d)
	}

	return err
}

func (w *Response) Render(c *Context, data interface{}) error {
	encoder := w.encoders.Get(w.encoding)
	if encoder != nil {
		// check data type
		b, err := encoder.Encode(data)
		if err != nil {
			return err
		}
		w.WriteHeader(200)
		_, err = w.Write(b)
		return err
	}
	return nil
}

func (w *Response) Stream(status int, d time.Duration, f func(int64) (interface{}, bool)) error {
	if f == nil {
		return nil
	}

	switch w.encoding {
	case framework.MIME_JSON:
		return w.StreamAsJson(status, d, f)
	case framework.MIME_XML:
		return w.StreamAsXml(status, d, f)
	}

	return nil
}

func WriteEntity(w http.ResponseWriter, status int, value interface{}, encoding string) (int, error) {
	if value == nil {
		return 0, nil
	}

	switch encoding {
	case framework.MIME_JSON:
		return WriteAsJson(w, status, value)
	case framework.MIME_XML:
		return WriteAsXml(w, status, value, true)
	}

	return 0, nil
}

func WriteAsXml(w http.ResponseWriter, status int, value interface{}, writeHeader bool) (int, error) {
	var output []byte
	var err error

	if value == nil {
		return 0, nil
	}

	if err != nil {
		return WriteString(w, http.StatusInternalServerError, err.Error())
	}
	w.Header().Set(framework.HEADER_ContentType, framework.MIME_XML)
	w.WriteHeader(status)
	if writeHeader {
		cl, err := w.Write([]byte(xml.Header))
		if err != nil {
			return cl, err
		}
	}
	return w.Write(output)

}

func WriteAsJson(w http.ResponseWriter, status int, value interface{}) (int, error) {
	return WriteJson(w, status, value, framework.MIME_JSON)
}

func WriteJson(w http.ResponseWriter, status int, value interface{}, contentType string) (int, error) {
	var output []byte
	var err error

	if value == nil {
		return 0, nil
	}

	if err != nil {
		return WriteString(w, http.StatusInternalServerError, err.Error())
	}
	w.Header().Set(framework.HEADER_ContentType, contentType)
	w.WriteHeader(status)
	return w.Write(output)
}

func WriteString(w http.ResponseWriter, status int, str string) (int, error) {
	w.WriteHeader(status)
	return w.Write([]byte(str))
}
