package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestLibrarySpreadRegExpUnion(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `interface Cursor { readonly lastIndex: number; }
function copy(value: Cursor | { readonly label: string }): void { const result = { ...value }; }
copy(/a/g);`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class RegExp") {
		test.Fatalf("want RegExp refusal, got %v", err)
	}
}

func TestLibrarySpreadErrorView(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `interface Message { readonly message: string; }
function copy(value: Message): void { const result = { ...value }; }
copy(new Error('bad'));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class Error") {
		test.Fatalf("want Error refusal, got %v", err)
	}
}

func TestLibrarySpreadPlainProducers(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `interface Cursor { readonly lastIndex: number; }
class CursorRecord { readonly lastIndex = 2; }
function copy(value: Cursor): void { const result = { ...value }; console.log(Object.keys(result).join(',')); }
const unrelated = /a/g;
copy({ lastIndex: 1 });
copy(new CursorRecord());`)
	if err != nil {
		test.Fatal(err)
	}
}

func TestLibrarySpreadMapView(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `interface Sized { readonly size: number; }
function copy(value: Sized): void { const result = { ...value }; }
copy(new Map<string, number>());`)
	// A Map going where Sized, a plain object type, is held stops at the argument (shape_fit.go), before
	// the spread could see it.
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(
		err.Error(),
		"main.a:3:6: stage 0 can't lower a Map or Set going where a plain object type is held yet",
	) {
		test.Fatalf("want the Map's NotYet at the argument, got %v", err)
	}
}

func TestLibrarySpreadSetView(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `interface Sized { readonly size: number; }
function copy(value: Sized): void { const result = { ...value }; }
copy(new Set<number>());`)
	// A Set going where Sized, a plain object type, is held stops at the argument (shape_fit.go), before
	// the spread could see it.
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(
		err.Error(),
		"main.a:3:6: stage 0 can't lower a Map or Set going where a plain object type is held yet",
	) {
		test.Fatalf("want the Set's NotYet at the argument, got %v", err)
	}
}

func TestLibrarySpreadDateView(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `interface Time { readonly getTime: () => number; }
function copy(value: Time): void { const result = { ...value }; }
copy(new Date(0));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class Date") {
		test.Fatalf("want Date refusal, got %v", err)
	}
}

func TestLibrarySpreadTypedArrayView(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `interface Length { readonly length: number; }
function copy(value: Length): void { const result = { ...value }; }
copy(new Uint8Array(2));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class Uint8Array") {
		test.Fatalf("want Uint8Array refusal, got %v", err)
	}
}
