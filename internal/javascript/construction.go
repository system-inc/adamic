package javascript

import "github.com/system-inc/adamic/internal/ir"

func jsConstructionFields(fields []ir.Field) bool {
	for _, field := range fields {
		if field.Absent {
			return true
		}
	}
	return false
}
