package jsutils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dop251/goja"
)

type HttpResult struct {
	runtime    *goja.Runtime
	self       goja.Value
	success    bool
	output     string
	statusCode int
}

func (er *HttpResult) OnSuccess(call goja.FunctionCall) goja.Value {
	if er.statusCode >= 200 && er.statusCode < 300 && er.success && len(call.Arguments) == 1 {
		if fn, ok := goja.AssertFunction(call.Arguments[0]); ok {
			_, _ = fn(nil, er.runtime.ToValue(er.output))
		}
	}
	return er.self
}

func (er *HttpResult) OnFail(call goja.FunctionCall) goja.Value {
	if er.statusCode >= 300 && len(call.Arguments) == 1 {
		if fn, ok := goja.AssertFunction(call.Arguments[0]); ok {
			_, _ = fn(nil, er.runtime.ToValue(er.output), er.runtime.ToValue(er.statusCode))
		}
	}
	return er.self
}

func (ctx *Easykube) Http() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {

		htr := &HttpResult{runtime: ctx.AddonCtx.vm, statusCode: -1}
		obj := ctx.AddonCtx.NewObject()
		htr.self = obj

		// bind methods
		_ = obj.Set("onSuccess", htr.OnSuccess)
		_ = obj.Set("onFail", htr.OnFail)

		// Extract arguments
		url := call.Argument(0).String()
		method := "GET"
		if len(call.Arguments) > 1 {
			method = call.Argument(1).String()
		}
		headers := make(map[string]string)
		if len(call.Arguments) > 2 {
			if h, ok := call.Argument(2).Export().(map[string]interface{}); ok {
				for k, v := range h {
					headers[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		body := ""
		if len(call.Arguments) > 3 {
			body = call.Argument(3).String()
		}

		// Make the HTTP request
		req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
		if err != nil {
			htr.output = err.Error()
			htr.success = false
			return obj

		} else {
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				htr.output = err.Error()
				htr.success = false
				return obj
			}

			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				htr.output = err.Error()
				htr.success = false
				htr.statusCode = -1
				return obj
			}

			htr.success = true
			htr.output = string(respBody)
			htr.statusCode = resp.StatusCode

		}

		return obj
	}
}
