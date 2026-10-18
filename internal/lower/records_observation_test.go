package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestRecordDiscardedReadProof(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const r:Record<string,number>={}; r['toString'];`,
		`function show(r:Record<string,number>):void {if(Object.hasOwn(r,'constructor')) {console.log(String(r['constructor']));}}`,
		`const r:Record<string,number>={}; const value = r['constructor'];`,
		`function show(r:Record<string,number>):void {const value=r['constructor']; if(Object.hasOwn(r,'constructor')){console.log(String(value));}}`,
		`function show(r:Record<string,number>,k:string):void {const value=r[k]; if(Object.hasOwn(r,k)){console.log(String(value));}}`,
	} {
		t.Run(source, func(t *testing.T) {
			p, err := lowerSource(t, source)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, f := range p.Functions {
				found = found || f.Name == "record_own_read"
			}
			if !found {
				t.Fatal("discarded read did not use the own-only helper")
			}
		})
	}
}

func TestRecordObservationMustNotEscapeGuard(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const r:Record<string,number>={}; const value=r['toString']; console.log(String(value));`,
		`function show(r:Record<string,number>):void {if(Object.hasOwn(r,'toString')) {delete r['toString']; console.log(String(r['toString']));}}`,
		`function show(r:Record<string,number>):void {if(Object.hasOwn(r,'other')) {console.log(String(r['toString']));}}`,
		`function show(r:Record<string,number>,k:string):void {const value=r[k]; console.log('between'); if(Object.hasOwn(r,k)){console.log(String(value));}}`,
		`function show(r:Record<string,number>,k:string):void {const value=r[k]; if(Object.hasOwn(r,k)){console.log(String(value));} else {console.log(String(value));}}`,
		`function show(r:Record<string,number>,k:string):void {const value=r[k]; if(Object.hasOwn(r,'other')){console.log(String(value));}}`,
		`const r:Record<string,number>={}; console.log(String(r['toString'] === undefined));`,
	} {
		t.Run(source, func(t *testing.T) {
			p, err := lowerSource(t, source)
			if strings.Contains(source, "r['toString']") {
				var refusal *Refused
				if !errors.As(err, &refusal) {
					t.Fatalf("want literal-member refusal, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range p.Functions {
				if f.Name == "record_own_read" {
					t.Fatal("observable read escaped the loud lookup")
				}
			}
		})
	}
}
