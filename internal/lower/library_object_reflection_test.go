package lower

import (
	"strings"
	"testing"
)

func TestObjectReflectionUnknownProducerRefusals(t *testing.T) {
	for _, source := range []string{
		"function count(value:Record<string,number>):number{return Object.entries(value).length;}",
		"let values:Record<string,number>={x:1}; Object.entries(values);",
		"const values:Record<string,number>={x:1}; values.next=2; Object.entries(values);",
		"const values:Record<string,number>={x:1}; const alias=values; alias.next=2; Object.entries(values);",
	} {
		if _, err := lowerSource(t, source); err == nil {
			t.Fatalf("unknown producer or expando accepted: %s", source)
		}
	}
}

func TestObjectReflectionKnownIndexProducer(t *testing.T) {
	source := "const values:{[key:string]:number}={x:1}; Object.entries(values);"
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	if program.ReflectionChecks != 1 {
		t.Fatalf("want one visible check, got %d", program.ReflectionChecks)
	}
}

func TestObjectReflectionNullMembersStayNotYet(t *testing.T) {
	_, err := lowerSource(t, "const values:Record<string,number|undefined>={x:1}; Object.entries(values);")
	if err == nil || !strings.Contains(err.Error(), "distinct null metadata") {
		t.Fatalf("want distinct null metadata stop, got %v", err)
	}
}
