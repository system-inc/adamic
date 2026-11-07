package fresh

import "github.com/system-inc/adamic/internal/ir"

func (analysis *analysis) recordCall(call ir.RecordCall) value {
	args := make([]value, len(call.Arguments))
	for index, arg := range call.Arguments {
		args[index] = analysis.value(arg)
	}
	switch call.Method {
	case "get":
		return analysis.load(args[0], elementKey)
	case "set":
		analysis.write(WriteMapEntry, call.Site, "", args[0], args[2], elementKey)
		return args[2]
	case "defineHidden":
		analysis.write(WriteMapEntry, call.Site, "", args[0], args[2], elementKey)
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
