package javascript

import (
	"fmt"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Every record the backend makes carries adamicRecords' mark, so a checked narrow tells a record from a fixed object as
// native's shape does (#6c4swvt): a record literal and an entries record, the two IR nodes that make one natively,
// each pass a narrow to a record and stop one to a fixed object, and a plain object literal the other way round.
// Mutant: make either maker without the mark (records.go's recordLiteral, or the ObjectLiteral's Record arm), and its
// row narrows to a fixed object without stopping.
func TestRecordMakersCarryTheMark(t *testing.T) {
	t.Parallel()
	value := ir.StringConstant{Index: 0}
	for _, maker := range []struct {
		name   string
		made   ir.Expression
		record bool
	}{
		{"record literal", ir.RecordLiteral{Element: ir.String, Entries: []ir.RecordEntry{{Key: ir.StringConstant{Index: 1}, Value: value}}}, true},
		{"entries record", ir.ObjectLiteral{Record: true, Fields: []ir.Field{{Name: "k", Value: value}}}, true},
		{"fixed object", ir.ObjectLiteral{Fields: []ir.Field{{Name: "k", Value: value}}}, false},
	} {
		t.Run(maker.name, func(t *testing.T) {
			t.Parallel()
			e := emitter{program: &ir.Program{Strings: []string{"a", "k"}}}
			made := e.valueRaw(maker.made)
			kept, refused := 14, 4
			if !maker.record {
				kept, refused = 4, 14
			}
			source := viewTestRuntime + fieldReadinessRuntime + recordRuntime +
				"const made = " + made + ";\n" +
				"console.log(String(adamicRecords.has(made)));\n" +
				"console.log(String(adamicPlaceholderNarrow(made, " + fmt.Sprint(kept) + ", \"kept\") === made));\n" +
				"adamicPlaceholderNarrow(made, " + fmt.Sprint(refused) + ", \"refused\");\n" +
				"console.log('read through');\n"
			mark := "true"
			if !maker.record {
				mark = "false"
			}
			runViewNode(t, source, mark+"\ntrue\n", "adamic: panic: refused\n", 70)
		})
	}
}
