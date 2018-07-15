package ctxrouter

import (
	"encoding/json"
	"net/http"
	"reflect"
)

//Param get params form request (It is faster than most other function, because there is no extra compute )
//req http.Request
func Params(req *http.Request) []string {
	return req.Header[paramHeader]
}

const paramHeader = "X-Ctxrouter-Params"

type ContextInterface interface {
	Init(http.ResponseWriter, *http.Request)
	DecodeRequest() error
}

func New() *Router {
	return &Router{
		handlers: make(map[string][]Handler),
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	val, _, params, err := r.Match(req.Method, req.URL.Path)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	if val.callT == nil {
		req.Header[paramHeader] = params
		if h, ok := val.callV.Interface().(http.HandlerFunc); ok {
			h.ServeHTTP(w, req)
		} else if hf, ok := val.callV.Interface().(func(http.ResponseWriter, *http.Request)); ok {
			hf(w, req)
		}
		return
	}
	ctx := reflect.New(val.callT).Interface().(ContextInterface)
	ctx.Init(w, req)
	if err := ctx.DecodeRequest(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in := []reflect.Value{reflect.ValueOf(ctx)}
	if val.hasParams {
		in = append(in, val.paramsV...)
	}
	rets := val.callV.Call(in)
	if len(rets) == 1 {
		if ret := rets[0]; !ret.IsNil() {
			if data, ok := rets[0].Interface().(interface{}); ok {
				if d, err := json.Marshal(data); err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(d)
				}
			}
		}
	} else if len(rets) == 2 {
		if rets[1].IsNil() {
			if data, ok := rets[0].Interface().(interface{}); ok {
				if d, err := json.Marshal(data); err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(d)
				}
			}
		} else {
			if httpError, ok := rets[1].Interface().(ErrorInterface); ok {
				d, _ := json.Marshal(httpError)
				w.Header().Set("Content-Type", "application/json")
				statusCode := httpError.StatusCode()
				if statusCode > 0 {
					w.WriteHeader(statusCode)
				} else {
					w.WriteHeader(400)
				}
				w.Write(d)
			} else if errCommon, ok := rets[1].Interface().(error); ok {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(errCommon.Error()))
			}
		}
	}
}

func (r *Router) Get(path string, controller interface{}) {
	if err := r.Handle("GET", path, controller); err != nil {
		panic(err)
	}
}

func (r *Router) Post(path string, controller interface{}) {
	if err := r.Handle("POST", path, controller); err != nil {
		panic(err)
	}
}

func (r *Router) Patch(path string, controller interface{}) {
	if err := r.Handle("PATCH", path, controller); err != nil {
		panic(err)
	}
}

func (r *Router) Put(path string, controller interface{}) {
	if err := r.Handle("PUT", path, controller); err != nil {
		panic(err)
	}
}

func (r *Router) Delete(path string, controller interface{}) {
	if err := r.Handle("DELETE", path, controller); err != nil {
		panic(err)
	}
}

func (r *Router) Head(path string, controller interface{}) {
	if err := r.Handle("HEAD", path, controller); err != nil {
		panic(err)
	}
}

func (r *Router) Options(path string, controller interface{}) {
	if err := r.Handle("OPTIONS", path, controller); err != nil {
		panic(err)
	}
}
func (r *Router) All(path string, controller interface{}) {
	if err := r.Handle("*", path, controller); err != nil {
		panic(err)
	}
}
