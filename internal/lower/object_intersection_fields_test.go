package lower

import (
	"errors"
	"strings"
	"testing"
)

// An object intersection is held as an object whatever its fields hold, so long as a record would
// hold each one: an object field and one that holds the intersection again lower. A Weak field and
// a two-word optional boolean have no such storage (recordStorageType), and a closure field waits on
// its own ruling, so each stays NotYet. So does an intersection with a constructable member, an
// index signature or a class instance, as a checked view over it would. A readonly index signature is
// stopped by the dictionary storage rule before any intersection, but a mutable one (a Record) and a
// class instance with fields alone are kept out only by the member's family (unsupportedViewFamily).
func TestObjectIntersectionFields(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, source string
		admitted     bool
	}{
		{
			"object field",
			"interface Label { readonly text: string; }\ntype Labeled = { readonly id: number } & { readonly label: Label };\nconst value: Labeled = { id: 1, label: { text: 'built' + '-label' } };\nconsole.log(value.label.text);\n",
			true,
		},
		{
			"itself",
			"type Chain = { readonly value: number } & { readonly next: Chain | undefined };\nconst chain: Chain = { value: 1, next: { value: 2, next: undefined } };\nconsole.log(`${chain.value + (chain.next === undefined ? 0 : chain.next.value)}`);\n",
			true,
		},
		{
			"weak field",
			"import type { Weak } from 'adamic';\ninterface Label { readonly text: string; }\ntype Held = { readonly id: number } & { readonly label: Weak<Label> };\nconst value: Held = { id: 1, label: undefined };\nconsole.log(`${value.id}`);\n",
			false,
		},
		{
			"optional boolean field",
			"type Flagged = { readonly id: number } & { readonly flag?: boolean };\nconst value: Flagged = { id: 1, flag: true };\nconsole.log(`${value.id}`);\n",
			false,
		},
		{
			"closure field",
			"type Called = { readonly id: number } & { readonly run: () => number };\nconst value: Called = { id: 1, run: (): number => 2 };\nconsole.log(`${value.id + value.run()}`);\n",
			false,
		},
		{
			"constructable member",
			"class Made { readonly made: number = 1; }\ntype Maker = { readonly id: number } & { new (): Made };\nfunction find(): Maker | undefined { return undefined; }\nconsole.log(typeof find());\n",
			false,
		},
		{
			"index signature member",
			"type Counts = { readonly id: number } & { readonly [key: string]: number };\nconst value: Counts = { id: 1 };\nconsole.log(`${value.id}`);\n",
			false,
		},
		{
			"record member",
			"type Counts = Record<string, number> & { readonly id: number };\nconst value: Counts = { id: 1, more: 2 };\nconsole.log(`${value.id}`);\n",
			false,
		},
		{
			"class instance member",
			"class Box { readonly v: number = 1; }\ntype Tagged = Box & { readonly tag: string };\nconst value: Tagged = { v: 2, tag: 'built' + '-tag' };\nconsole.log(`${value.v} ${value.tag}`);\n",
			false,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			if probe.admitted {
				if lowerSourceError != nil {
					test.Fatal(lowerSourceError)
				}
				return
			}
			var stop *NotYet
			if !errors.As(lowerSourceError, &stop) {
				test.Fatalf("want NotYet, got %v", lowerSourceError)
			}
		})
	}
}

// adamic_object_assign retains and releases by the source slot's reference flag, which is safe only
// while a name is held alike on both sides. A field that's a string in the target and a boxed
// string | number in the source is refused, and so is assigning an object with an object field to
// itself: Object.assign copies only number, string and boolean fields.
func TestObjectAssignKeepsFieldsHeldAlike(test *testing.T) {
	test.Parallel()
	for name, source := range map[string]string{
		"string and boxed union":      "const target = { value: 'built' + '-text' };\nconst source = { value: 1 as string | number };\nObject.assign(target, source);\nconsole.log(target.value);\n",
		"itself with an object field": "const target = { count: 1, label: { text: 'built' + '-label' } };\nconst same = Object.assign(target, target);\nconsole.log(`${same.count} ${same.label.text}`);\n",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, source)
			var refusal *Refused
			if !errors.As(lowerSourceError, &refusal) ||
				!strings.Contains(lowerSourceError.Error(), "intersection result") {
				test.Fatalf("want the intersection result refusal, got %v", lowerSourceError)
			}
		})
	}
}
