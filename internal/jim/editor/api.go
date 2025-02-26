package editor

import (
	"bytes"

	"github.com/jstotz/jim/internal/jim/command"
	lua "github.com/yuin/gopher-lua"
)

type APIModule struct {
	editor *Editor
}

func NewAPIModule(e *Editor) *APIModule {
	return &APIModule{
		editor: e,
	}
}

func (m *APIModule) Load() {
	l := m.editor.luaState
	m.editor.Logger.Debug("Loading API module")
	mod := l.NewTable()
	apiMod := l.SetFuncs(l.NewTable(), m.exports())
	l.SetField(mod, "api", apiMod)
	l.SetGlobal("print", m.printFunction(l))
	l.SetGlobal("jim", mod)
}

func (m *APIModule) printFunction(l *lua.LState) *lua.LFunction {
	return l.NewFunction(func(l *lua.LState) int {
		m.editor.Logger.Debug("printing")
		var buf bytes.Buffer
		top := l.GetTop()
		for i := 1; i <= top; i++ {
			buf.WriteString(l.Get(i).String())
			if i != top {
				buf.WriteString("\t")
			}
		}
		// Final newline
		buf.WriteString("\n")

		m.editor.Logger.Debug(buf.String())

		return 0 // number of results
	})
}

func (m *APIModule) exports() map[string]lua.LGFunction {
	expts := map[string]lua.LGFunction{
		"delete": m.apiDelete,
	}
	for name, fn := range expts {
		expts[name] = m.wrapAPIFunction(name, fn)
	}
	return expts
}

func (m *APIModule) wrapAPIFunction(name string, fn lua.LGFunction) lua.LGFunction {
	return func(l *lua.LState) int {
		m.editor.Logger.Debug("Calling API method", "method", name)
		defer func() {
			if r := recover(); r != nil {
				m.editor.Logger.Debug("recovered", "method", name)
			}
		}()
		return fn(l)
	}
}

func (m *APIModule) runCommand(l *lua.LState, cmd command.Command) int {
	if err := m.editor.runCommand(cmd); err != nil {
		l.RaiseError("command failed: %s", err.Error())
	}
	return 0
}

func (m *APIModule) apiDelete(l *lua.LState) int {
	return m.runCommand(l, command.DeleteText{Length: 1})
}
