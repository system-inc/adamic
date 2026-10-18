package oracle

import "testing"

func init() {
	for _, name := range []string{"array-holes-view", "array-lazy", "array-iteration", "array-map", "array-pop", "array-bounds", "array-second", "array-length-kind", "array-missing", "array-boolean", "array-iteration-bad", "array-map-bad", "array-alias", "array-undefined", "array-string-literal", "nullable-array-string", "nullable-array-missing", "nullable-array-element", "generic-array-element", "array-field", "array-element", "object-element", "non-array"} {
		path := "stage3/interface-downcasts/lane2/" + name + ".a"
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) { return counted(t, path, false, nil, false, false) }})
	}
}
