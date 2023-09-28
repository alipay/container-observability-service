package spans

import (
	lua "github.com/yuin/gopher-lua"
)

var helperFunctions = make(map[string]lua.LGFunction)

func RegisterLuaHelperFunc(name string, helper func(L *lua.LState) int) {
	helperFunctions[name] = helper
}

func Loader(L *lua.LState) int {
	// register functions to the table
	mod := L.SetFuncs(L.NewTable(), helperFunctions)

	// returns the module
	L.Push(mod)
	return 1
}

func InjectHelperFuncToLua(L *lua.LState) {
	L.PreloadModule("delivery", Loader)
}
