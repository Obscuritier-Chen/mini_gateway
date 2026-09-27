package lua

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/yuin/gopher-lua"
)

type Engine struct {
	pool sync.Pool
	scriptPath string
}

func New(scriptPath string) (*Engine, error) {
	testL := lua.NewState()
	defer testL.Close()
	if err := testL.DoFile(scriptPath); err!=nil {
		return nil, fmt.Errorf("failed to load lua script: %w", err)
	}

	e := &Engine{
		scriptPath: scriptPath,
	}

	e.pool = sync.Pool{
		New: func() interface{} {
			L := lua.NewState()
			_ = L.DoFile(scriptPath)
			return L
		},
	}

	return e, nil
}

func (e *Engine) ExecuteRule(r *http.Request, requestID string) (int, string, error) {
	L := e.pool.Get().(*lua.LState)
	defer e.pool.Put(L)

	fn := L.GetGlobal("check_request")
	if fn.Type() != lua.LTFunction {
		return 500, "", fmt.Errorf("failed to find lua func check_request")
	}

	reqTable := L.NewTable()
	reqTable.RawSetString("method", lua.LString(r.Method))
	reqTable.RawSetString("path", lua.LString(r.URL.Path))
	reqTable.RawSetString("request_id", lua.LString(requestID))

	headersTable := L.NewTable()
	for k, v := range r.Header {
		if len(v) >0 {
			headersTable.RawSetString(k, lua.LString(v[0]))
		}
	}
	reqTable.RawSetString("headers", headersTable)

	queryTable := L.NewTable()
	for k, v := range r.URL.Query() {
		if len(v) >0 {
			queryTable.RawSetString(k, lua.LString(v[0]))
		}
	}
	reqTable.RawSetString("query", queryTable)

	err := L.CallByParam(lua.P{
		Fn: fn,
		NRet: 2, //两个返回值
		Protect: true,
	}, reqTable)

	if err != nil {
		return 500, "", err
	}

	msg := L.ToString(-1)
	statusCode := L.ToInt(-2)
	L.Pop(2)

	return statusCode, msg, nil
}