package fresh

import "github.com/system-inc/adamic/internal/ir"

func (analysis *analysis) recordCall(c ir.RecordCall) value {
	args := make([]value, len(c.Arguments))
	for i, arg := range c.Arguments {
		args[i] = analysis.value(arg)
	}
	switch c.Method {
	case "get":
		return analysis.load(args[0], elementKey)
	case "set":
		analysis.write(WriteMapEntry, c.Site, "", args[0], args[2], elementKey)
		return args[2]
	case "defineHidden":
		analysis.write(WriteMapEntry, c.Site, "", args[0], args[2], elementKey)
		return value{}
	case "values":
		return analysis.fresh(elementKey, analysis.load(args[0], elementKey))
	case "entries":
		return analysis.fresh(elementKey, analysis.fresh(anyField, analysis.load(args[0], elementKey)))
	case "keys":
		return analysis.fresh(elementKey, value{})
	}
	return value{}
}
