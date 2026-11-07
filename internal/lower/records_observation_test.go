package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestRecordDiscardedReadProof(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`const r:Record<string,number>={}; r['toString'];`,
		`function show(r:Record<string,number>):void {if(Object.hasOwn(r,'constructor')) {console.log(String(r['constructor']));}}`,
		`const r:Record<string,number>={}; const value = r['constructor'];`,
		`function show(r:Record<string,number>):void {const value=r['constructor']; if(Object.hasOwn(r,'constructor')){console.log(String(value));}}`,
		`function show(r:Record<string,number>,k:string):void {const value=r[k]; if(Object.hasOwn(r,k)){console.log(String(value));}}`,
	} {
		test.Run(source, func(test *testing.T) {
			program, lowerSourceError := lowerSource(test, source)
			if lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
			found := false
			for _, function := range program.Functions {
				found = found || function.Name == "record_own_read"
			}
			if !found {
				test.Fatal("discarded read did not use the own-only helper")
			}
		})
	}
}

func TestRecordObservationMustNotEscapeGuard(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`const r:Record<string,number>={}; const value=r['toString']; console.log(String(value));`,
		`function show(r:Record<string,number>):void {if(Object.hasOwn(r,'toString')) {delete r['toString']; console.log(String(r['toString']));}}`,
		`function show(r:Record<string,number>):void {if(Object.hasOwn(r,'other')) {console.log(String(r['toString']));}}`,
		`function show(r:Record<string,number>,k:string):void {const value=r[k]; console.log('between'); if(Object.hasOwn(r,k)){console.log(String(value));}}`,
		`function show(r:Record<string,number>,k:string):void {const value=r[k]; if(Object.hasOwn(r,k)){console.log(String(value));} else {console.log(String(value));}}`,
		`function show(r:Record<string,number>,k:string):void {const value=r[k]; if(Object.hasOwn(r,'other')){console.log(String(value));}}`,
		`const r:Record<string,number>={}; console.log(String(r['toString'] === undefined));`,
	} {
		test.Run(source, func(test *testing.T) {
			program, lowerSourceError := lowerSource(test, source)
			if strings.Contains(source, "r['toString']") {
				var refusal *Refused
				if !errors.As(lowerSourceError, &refusal) {
					test.Fatalf("want literal-member refusal, got %v", lowerSourceError)
				}
				return
			}
			if lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
			for _, function := range program.Functions {
				if function.Name == "record_own_read" {
					test.Fatal("observable read escaped the loud lookup")
				}
			}
		})
	}
}
