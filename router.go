package ctxrouter

import (
	"net/http"
	"reflect"
	"strconv"
	"errors"
)

var router *Router

type (
	Router struct {
		tree   *node
		params []reflect.Value
	}
        //not use in single pakage
	ContextInterface interface {
		Init(http.ResponseWriter, *http.Request)
		DecodeRequest() error
	}
)


type leaf struct {
	data map[string]Value
}

type Value  struct {
	CallV     reflect.Value
	CallT     reflect.Type
	V         interface{}
	ParamsV   []reflect.Value
	ParamsT   []reflect.Type
	HasParams bool //faster when callback
}

func New() *Router {
	router = &Router{
		tree: &node{},
	}
	return router
}

func (this *Router) Add(path, method string, v interface{}) error {
	if method == "" {
		method = "default"
	}
	val := Value{
		V:v,
		CallV:reflect.ValueOf(v),
		CallT:reflect.TypeOf(v),
	}
	if reflect.TypeOf(v).Kind() == reflect.Func {
		val.CallT = reflect.TypeOf(v).In(0).Elem()
		paramsLen := val.CallV.Type().NumIn()
		val.HasParams = paramsLen > 1
		for i := 0; i < paramsLen; i++ {
			if i > 0 {
				if i == 1 {
					val.ParamsT = make([]reflect.Type, 0)
					val.ParamsT = append(val.ParamsT, val.CallV.Type().In(i))
				} else if i > 1 {
					val.ParamsT = append(val.ParamsT, val.CallV.Type().In(i))
				}
			}
		}
	}
	if vMap, _, _ := this.tree.getValue(path); vMap != nil {
		if vMap, ok := vMap.(*leaf); ok {
			vMap.data[method] = val
			return nil
		} else {
			panic("router value is not a value map")
		}
	}
	if err := this.tree.addRoute(path, &leaf{data: map[string]Value{method:val}}); err != nil {
		return err
	}
	return nil
}


func (this *Router) Match(method, path string) (val Value, p []string) {
	if v, p, _ := this.tree.getValue(path); v != nil {
		if v, ok := v.(*leaf); ok {
			if v.data[method].V != nil {
				val = v.data[method]
			} else {
				val = v.data["default"]
			}
			if val.V != nil && p != nil {
				val.ParamsV = make([]reflect.Value, 0)
				for i, n := range p {
					pt := val.ParamsT[i]
					pv, err := strConv(n, pt)
					if err == nil {
						val.ParamsV = append(val.ParamsV, pv)
					} else {
						return Value{}, nil
					}
				}
			}
			return val, p
		}
		panic("router value is not valueMap")
	}
	return val, p
}

func (this *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	val, _ := this.Match(r.Method, r.URL.Path)
	if val.V == nil {
		http.NotFoundHandler().ServeHTTP(w, r)
		return
	}
	ctx := reflect.New(val.CallT).Interface().(ContextInterface)
	ctx.Init(w, r)
	if err := ctx.DecodeRequest(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	in := []reflect.Value{reflect.ValueOf(ctx)}
	if val.HasParams {
		in = append(in, val.ParamsV...)
	}
	val.CallV.Call(in)
}

func (this *Router) Get(path string, controller interface{}) {
	if err := this.Add(path, "GET", controller); err != nil {
		panic(err)
	}
}

func (this *Router) Post(path string, controller interface{}) {
	if err := this.Add(path, "POST", controller); err != nil {
		panic(err)
	}
}

func (this *Router) Patch(path string, controller interface{}) {
	if err := this.Add(path, "PATCH", controller); err != nil {
		panic(err)
	}
}

func (this *Router) Put(path string, controller interface{}) {
	if err := this.Add(path, "PUT", controller); err != nil {
		panic(err)
	}
}

func (this *Router) Delete(path string, controller interface{}) {
	if err := this.Add(path, "DELETE", controller); err != nil {
		panic(err)
	}
}

func (this *Router) All(path string, controller interface{}) {
	if err := this.Add(path, "", controller); err != nil {
		panic(err)
	}
}
//strConv convert string params to function params
func strConv(src string, t reflect.Type) (rv reflect.Value, err error) {
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(src), nil
	case reflect.Int:
		v, err := strconv.Atoi(src)
		if err == nil {
			rv = reflect.ValueOf(v)
		}
		return rv, err
	case reflect.Int64:
		v, err := strconv.ParseInt(src, 10, 64)
		if err == nil {
			rv = reflect.ValueOf(v)
		}
		return rv, err
	case reflect.Bool:
		v, err := strconv.ParseBool(src)
		if err == nil {
			rv = reflect.ValueOf(v)
		}
		return rv, err
	case reflect.Float64:
		v, err := strconv.ParseFloat(src, 64)
		if err == nil {
			rv = reflect.ValueOf(v)
		}
		return rv, err
	case reflect.Float32:
		v, err := strconv.ParseFloat(src, 32)
		if err == nil {
			rv = reflect.ValueOf(float32(v))
		}
		return rv, err
	case reflect.Uint64:
		v, err := strconv.ParseUint(src, 10, 64)
		if err == nil {
			rv = reflect.ValueOf(v)
		}
		return rv, err
	case reflect.Uint32:
		v, err := strconv.ParseUint(src, 10, 32)
		if err == nil {
			rv = reflect.ValueOf(uint32(v))
		}
		return rv, err
	default:
		return rv, errors.New("elem of invalid type")
	}
}