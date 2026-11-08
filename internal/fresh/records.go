package fresh

import "github.com/system-inc/adamic/internal/ir"

func (analysis *analysis) recordCall(call ir.RecordCall) value {
	arguments := make([]value, len(call.Arguments))
	for index, argument := range call.Arguments {
		arguments[index] = analysis.value(argument)
	}
	switch call.Method {
	case "get":
		return analysis.load(arguments[0], elementKey)
	case "set":
		analysis.write(WriteMapEntry, call.Site, "", arguments[0], arguments[2], elementKey)
		return arguments[2]
	case "defineHidden":
		analysis.write(WriteMapEntry, call.Site, "", arguments[0], arguments[2], elementKey)
		return value{}
	case "values":
		return analysis.fresh(elementKey, analysis.load(arguments[0], elementKey))
	case "entries":
		return analysis.fresh(elementKey, analysis.fresh(anyField, analysis.load(arguments[0], elementKey)))
	case "keys":
		return analysis.fresh(elementKey, value{})
	}
	return value{}
}
