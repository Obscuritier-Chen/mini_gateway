package lua

import  (
	"fmt"
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

func (e *Engine) ExecuteRule(path, userAgent string) (int, string, error) {
	L := e.pool.Get().(*lua.LState)
	defer e.pool.Put(L)

	fn := L.GetGlobal("check_request")
	if fn.Type() != lua.LTFunction {
		return 500, "", fmt.Errorf("failed to find lua func check_request")
	}

	err := L.CallByParam(lua.P{
		Fn: fn,
		NRet: 2, //两个返回值
		Protect: true,
	}, lua.LString(path), lua.LString(userAgent))

	if err != nil {
		return 500, "", err
	}

	msg := L.ToString(-1)
	statusCode := L.ToInt(-2)
	L.Pop(2)

	return statusCode, msg, nil
}