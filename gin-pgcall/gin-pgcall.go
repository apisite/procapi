// ginpgcall is a gin bindings for pgcall
package ginpgcall

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/birkirb/loggers.v1"
)

// Caller allows to process and mock pgcall.Caller independently
type Caller interface {
	Call(r *http.Request, method string, args map[string]interface{}) (interface{}, error)
}

// callError allows to process pgcall.CallError errors independently
type callError interface {
	IsNotFound() bool
	IsBadRequest() bool
	Code() string
	Error() string
	Data() map[string]interface{}
}

// Server holds pgcall server
type Server struct {
	Caller                    //	*pgcall.Server
	log    loggers.Contextual // for debugging only
}

// NewServer returns pgcall server object
func NewServer(log loggers.Contextual, pgcall Caller) *Server {
	return &Server{Caller: pgcall, log: log}
}

// Route registers supported locations in gin
func (srv *Server) Route(prefix string, r *gin.Engine) error {
	uri := prefix + "/:method"
	r.GET(uri, srv.handler(binding.Query))
	r.POST(uri, srv.handler(binding.JSON))
	return nil
}

// handler handles location request
func (srv *Server) handler(bind binding.Binding) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Param("method")
		//	srv.Log.Debugf("ginapi - handler (%s.%s)", nsp, name)
		var a map[string]interface{}
		if ctx.Request.Method == "POST" {
			// TODO: if Content-Type: application/json ... else form-data
			// TODO: file upload
			err := json.NewDecoder(ctx.Request.Body).Decode(&a)
			//		err := ctx.ShouldBindWith(&a, bind)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "data": err})
				return
			}
		} else {
			a = map[string]interface{}{}
			vals := ctx.Request.URL.Query()
			for k, v := range vals {
				a[k] = v[0]
			}
		}
		result, err := srv.Call(ctx.Request, name, a)
		if err != nil {
			// AbortWithStatusJSON
			var status int
			data := map[string]interface{}{
				"message": err.Error(),
			}

			perr, ok := err.(callError)
			if ok {
				if perr.IsNotFound() {
					status = http.StatusNotFound
				} else if perr.IsBadRequest() {
					status = http.StatusBadRequest
				} else {
					status = http.StatusServiceUnavailable
				}
				data["code"] = perr.Code()
				data["data"] = perr.Data()
			} else {
				status = http.StatusServiceUnavailable
			}
			ctx.JSON(status, data) //gin.H{"code": "METHOD_ERROR", "data": err.Error()})
			return
		}
		//srv.Log.Debugf("ginapi - result(%+v)", result)

		ctx.JSON(http.StatusOK, result)
	}
}

// SetFuncBlank appends function templates and not related to request functions to funcs
func (srv *Server) SetFuncBlank(funcs template.FuncMap) {

	funcs["makeSlice"] = func(param ...interface{}) interface{} {
		return param
	}
	funcs["makeMap"] = MakeMap
	funcs["api"] = func(method string, dict ...interface{}) (interface{}, error) {
		return "", nil
	}
	funcs["api_map"] = func(method string, args map[string]interface{}) (interface{}, error) {
		return "", nil
	}
	funcs["json"] = func(in interface{}) (template.HTML, error) {
		out, err := json.MarshalIndent(in, "", "  ")
		return template.HTML(out), err
	}
	funcs["param"] = func(key string) string { return "" }
	funcs["get"] = func(keys ...string) *map[string]interface{} { return nil }
	funcs["item"] = func(in map[string][]string, key string) *string {
		val, ok := in[key]
		if !ok {
			return nil
		}
		return &val[0]
	}
}

type RequestContext interface {
	// Get(key string) (value interface{}, exists bool)
}

// SetFuncRequest appends related to request functions to funcs
func (srv *Server) SetFuncRequest(funcs template.FuncMap, ctx *gin.Context) {
	funcs["param"] = func(key string) string { return ctx.Param(key) } // TODO: use original
	funcs["api"] = func(method string, dict ...interface{}) (interface{}, error) {
		//log.Debugf("Call for %s - Dict: %+v", method, dict)
		argsMap, err := MakeMap(dict...)
		if err != nil {
			return nil, err
		}
		return srv.Call(ctx.Request, method, *argsMap)
	}
	funcs["api_map"] = func(method string, args map[string]interface{}) (interface{}, error) {
		//log.Debugf("Call apimap for %s - Dict: %+v", method, args)
		return srv.Call(ctx.Request, method, args)
	}
	funcs["get"] = func(keys ...string) *map[string]interface{} {
		rv := map[string]interface{}{}
		for _, k := range keys {
			val, ok := ctx.Request.URL.Query()[k]
			if !ok || len(val) == 0 || len(val) == 1 && val[0] == "" {
				continue
			}
			rv[k] = val[0]
		}
		return &rv
	}
}

// MakeMap makes a map from key,value pairs
func MakeMap(args ...interface{}) (*map[string]interface{}, error) {
	if len(args) == 1 {
		// already map
		a := args[0].(map[string]interface{})
		return &a, nil
	}
	if len(args)%2 != 0 {
		// log.Printf("Args: %+v", args)
		return nil, errors.New("arg count must be even")
	}

	dict := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		key, isset := args[i].(string)
		if !isset {
			return nil, errors.Errorf("not string key in position %d", i)
		}
		dict[key] = args[i+1]
	}
	return &dict, nil
}
