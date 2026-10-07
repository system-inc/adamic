package lower

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestShapeErasureFixtures(t *testing.T) {
	for _, name := range []string{"proven", "nonconforming", "uninitialized", "host"} {
		t.Run(name, func(t *testing.T) {
			checked, err := load.Load([]string{"../../stage3/interface-downcasts/lane3/" + name + ".a"})
			if err != nil {
				t.Fatal(err)
			}
			program, err := Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			casts, reads := 0, 0
			inspect := func(node any) bool {
				if literal, ok := node.(ir.ObjectLiteral); ok && name == "uninitialized" {
					for _, field := range literal.Fields {
						if field.Name == "ready" && (field.Certificate == nil || field.Certificate.DeclaredType != "boolean") {
							t.Fatal("staged slot lost its declared boolean type")
						}
					}
				}
				if _, ok := node.(ir.CheckedCast); ok {
					casts++
				}
				if property, ok := node.(ir.Property); ok && property.View != "" {
					reads++
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			t.Logf("remaining casts=%d checked reads=%d", casts, reads)
			if name == "proven" && (casts != 0 || reads != 0) {
				t.Fatal("proven factory retained runtime checks")
			}
			if name != "proven" && reads == 0 {
				t.Fatal("unsafe erasure removed field checks")
			}
		})
	}
}

func TestShapeCertificatesSeparateNumberAndBoolean(t *testing.T) {
	program, err := lowerSource(t, `const number={value:0};const boolean={value:false};console.log(String(number.value));console.log(String(boolean.value));`)
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]bool{}
	walk(program.Main, func(node any) bool {
		if literal, ok := node.(ir.ObjectLiteral); ok {
			for _, field := range literal.Fields {
				if field.Name == "value" && field.Certificate != nil {
					declared[field.Certificate.DeclaredType] = true
				}
			}
		}
		return true
	})
	if !declared["number"] || !declared["boolean"] {
		t.Fatalf("semantic certificates conflated physical scalar slots: %v", declared)
	}
}
