package lower

import (
	"errors"
	"testing"
)

func TestOverloadMixedComparisonRemainsNotYet(t *testing.T) {
	_, err := lowerSource(t, `function compare(a:number,b:string):boolean; function compare(a:string|number,b:string|number):boolean {return a < b;} console.log(compare(2,"10").toString());`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "a BinaryExpression with a union of differently held members and a union of differently held members" {
		t.Fatalf("want mixed comparison unsupported, got %v", err)
	}
}
func TestOverloadReassignedComparisonRemainsNotYet(t *testing.T) {
	_, err := lowerSource(t, `function compare(a:number,b:number):boolean; function compare(a:string|number,b:string|number):boolean {function replacement():string|number {return "10";} a=replacement(); return a < b;} console.log(compare(2,3).toString());`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "a BinaryExpression with a union of differently held members and a union of differently held members" {
		t.Fatalf("want reassigned comparison unsupported, got %v", err)
	}
}
