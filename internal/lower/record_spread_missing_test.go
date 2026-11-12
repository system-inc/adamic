package lower

import (
	"strings"
	"testing"
)

// A spread into record storage from a fixed object that may be undefined keeps main's stop (#bzrdem0,
// s4): only a source whose present arm is record storage is admitted, so Point | undefined is judged
// as Point, a fixed object, and stops where main stops it. The admitted spreads agree with Node in
// internal/oracle's TestRecordSpreadFromMissingAgreesWithNode.
func TestRecordSpreadFromMissingFixedObjectStops(t *testing.T) {
	t.Parallel()
	const source = "interface Point {\n\tx: number;\n\ty: number;\n}\nfunction spread(point: Point | undefined): Record<string, number> {\n\treturn { ...point, z: 1 };\n}\nconsole.log(String(Object.keys(spread(undefined)).length));\n"
	const want = "6:11: stage 0 can't lower a spread into record storage from a fixed object yet"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			if err == nil || !strings.HasSuffix(err.Error(), "/main"+extension+":"+want) {
				t.Errorf("got %v, want main%s:%s", err, extension, want)
			}
		})
	}
}
