package jsutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dop251/goja"
)

func (ctx *Easykube) Http() func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
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
			panic(ctx.AddonCtx.vm.NewGoError(err))
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(ctx.AddonCtx.vm.NewGoError(err))
		}
		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(ctx.AddonCtx.vm.NewGoError(err))
		}

		// Return the response as a JavaScript object
		obj := ctx.AddonCtx.vm.NewObject()
		obj.Set("status", resp.StatusCode)
		obj.Set("body", string(respBody))
		return obj
	}
}
