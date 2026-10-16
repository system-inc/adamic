package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestLibrarySpreadRegExpUnion(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Cursor { readonly lastIndex: number; }
function copy(value: Cursor | { readonly label: string }): void { const result = { ...value }; }
copy(/a/g);`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class RegExp") {
		t.Fatalf("want RegExp refusal, got %v", err)
	}
}

func TestLibrarySpreadErrorView(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Message { readonly message: string; }
function copy(value: Message): void { const result = { ...value }; }
copy(new Error('bad'));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class Error") {
		t.Fatalf("want Error refusal, got %v", err)
	}
}

func TestLibrarySpreadPlainProducers(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Cursor { readonly lastIndex: number; }
class CursorRecord { readonly lastIndex = 2; }
function copy(value: Cursor): void { const result = { ...value }; console.log(Object.keys(result).join(',')); }
const unrelated = /a/g;
copy({ lastIndex: 1 });
copy(new CursorRecord());`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLibrarySpreadMapView(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Sized { readonly size: number; }
function copy(value: Sized): void { const result = { ...value }; }
copy(new Map<string, number>());`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class Map") {
		t.Fatalf("want Map refusal, got %v", err)
	}
}

func TestLibrarySpreadSetView(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Sized { readonly size: number; }
function copy(value: Sized): void { const result = { ...value }; }
copy(new Set<number>());`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class Set") {
		t.Fatalf("want Set refusal, got %v", err)
	}
}

func TestLibrarySpreadDateView(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Time { readonly getTime: () => number; }
function copy(value: Time): void { const result = { ...value }; }
copy(new Date(0));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class Date") {
		t.Fatalf("want Date refusal, got %v", err)
	}
}

func TestLibrarySpreadTypedArrayView(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Length { readonly length: number; }
function copy(value: Length): void { const result = { ...value }; }
copy(new Uint8Array(2));`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "library class Uint8Array") {
		t.Fatalf("want Uint8Array refusal, got %v", err)
	}
}
