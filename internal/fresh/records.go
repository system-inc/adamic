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
	case "defineHidden":
		a.write(WriteMapEntry, c.Site, "", args[0], args[2], elementKey)
		return value{}
	case "values":
		return a.fresh(elementKey, a.load(args[0], elementKey))
	case "entries":
		return a.fresh(elementKey, a.fresh(anyField, a.load(args[0], elementKey)))
	case "keys":
		return a.fresh(elementKey, value{})
	case "forInKeys":
		// The list the record's table may keep for every for...in (runtime record.c), held there too, so
		// it's never this function's alone.
		return outsideValue()
	}
	return value{}
}
