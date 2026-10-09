package oracle

import "testing"

func init() { additionalFixtureCounts = append(additionalFixtureCounts, checkedViewArrayCounts) }

func checkedViewArrayCounts(t *testing.T) []string {
	rows := []string{}
	for _, name := range []string{"array-lazy", "array-iteration", "array-map", "array-pop", "array-bounds", "array-second", "array-length-kind", "array-missing", "array-boolean", "array-iteration-bad", "array-map-bad", "array-alias", "array-undefined", "array-string-literal", "nullable-array-string", "nullable-array-element", "generic-array-element", "array-field", "array-element", "object-element", "non-array"} {
		rows = append(rows, counted(t, "stage3/interface-downcasts/lane2/"+name+".a", false, nil, false, false))
	}
	return rows
}
