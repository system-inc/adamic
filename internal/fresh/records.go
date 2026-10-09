package fresh

import "github.com/system-inc/adamic/internal/ir"

func (a *analysis) recordCall(c ir.RecordCall) value {
	args := make([]value, len(c.Arguments))
	for i, arg := range c.Arguments {
		args[i] = a.value(arg)
	}
	switch c.Method {
	case "get":
		return a.load(args[0], elementKey)
	case "set":
		a.write(WriteMapEntry, c.Site, "", args[0], args[2], elementKey)
		return args[2]
	case "values":
		return a.fresh(elementKey, a.load(args[0], elementKey))
	case "entries":
		return a.fresh(elementKey, a.fresh(anyField, a.load(args[0], elementKey)))
	case "keys":
		return a.fresh(elementKey, value{})
	}
	return value{}
}
