package main

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/stage3/census/latent/privatefields"
)

// Audit every field against the old copier's contract, including exported fields
// and the shared identity field which must never become a skipped field.
func TestLatentPrivateIRCacheAnswersMatchIRAudit20pb346(t *testing.T) {
	t.Parallel()
	owners := []reflect.Type{reflect.TypeFor[ir.Program](), reflect.TypeFor[ir.BoundaryStringTable](), reflect.TypeFor[ir.GeneratedType](), reflect.TypeFor[ir.Function]()}
	for _, owner := range owners {
		policies := map[string]privatefields.Policy{}
		for i := 0; i < owner.NumField(); i++ {
			name := owner.Field(i).Name
			if policy, ok := irPrivateFields[owner.Name()+"."+name]; ok {
				switch policy {
				case privateNil:
					policies[name] = privatefields.Nil
				case privateEmpty:
					policies[name] = privatefields.Empty
				}
			}
		}
		plan := privatefields.New(owner, policies)
		value := reflect.New(owner).Elem()
		for i := 0; i < owner.NumField(); i++ {
			key := owner.Name() + "." + owner.Field(i).Name
			want := key == "Program.argumentFacts" || key == "Program.methodTables" || key == "BoundaryStringTable.indexes"
			if got := plan.Skip(value, i); got != want {
				t.Errorf("%s: skip=%v, want %v", key, got, want)
			}
		}
	}
}
