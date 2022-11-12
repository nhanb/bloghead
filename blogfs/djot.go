package blogfs

import (
	_ "embed"
	"html/template"

	lua "github.com/yuin/gopher-lua"
)

//go:embed djot.lua
var djotlua string

var ls *lua.LState

func djotToHtml(input string) template.HTML {
	if ls == nil {
		ls = lua.NewState()
		if err := ls.DoString(djotlua); err != nil {
			panic(err)
		}
	}
	if err := ls.CallByParam(lua.P{
		Fn:      ls.GetGlobal("djot_to_html"),
		NRet:    1,
		Protect: true,
	}, lua.LString(input)); err != nil {
		panic(err)
	}
	ret := ls.Get(-1) // returned value
	ls.Pop(1)         // remove received value
	return template.HTML(ret.String())
}
