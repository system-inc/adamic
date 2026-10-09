package tailwind

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const physicalDirectionFile = "/repository/source/components/Thing.tsx"

func TestNoPhysicalDirectionFires(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
		count      int
	}{
		// One per family, because each is a separate table entry and a rule missing one would pass
		// every fixture written for the others.
		{"margin left", "export const c = 'flex ml-4';\n", 1},
		{"margin right", "export const c = 'flex mr-4';\n", 1},
		{"padding left", "export const c = 'flex pl-4';\n", 1},
		{"padding right", "export const c = 'flex pr-4';\n", 1},
		{"text left", "export const c = 'flex text-left';\n", 1},
		{"text right", "export const c = 'flex text-right';\n", 1},
		{"position left", "export const c = 'flex left-0';\n", 1},
		{"position right", "export const c = 'flex right-0';\n", 1},
		{"border left bare", "export const c = 'flex border-l';\n", 1},
		{"border left with value", "export const c = 'flex border-l-2';\n", 1},
		{"border right bare", "export const c = 'flex border-r';\n", 1},
		{"border right with value", "export const c = 'flex border-r-2';\n", 1},
		{"rounded left bare", "export const c = 'flex rounded-l';\n", 1},
		{"rounded left with value", "export const c = 'flex rounded-l-md';\n", 1},
		{"rounded right bare", "export const c = 'flex rounded-r';\n", 1},
		{"rounded right with value", "export const c = 'flex rounded-r-md';\n", 1},
		{"scroll margin left", "export const c = 'flex scroll-ml-4';\n", 1},
		{"scroll margin right", "export const c = 'flex scroll-mr-4';\n", 1},
		{"scroll padding left", "export const c = 'flex scroll-pl-4';\n", 1},
		{"scroll padding right", "export const c = 'flex scroll-pr-4';\n", 1},
		// Variant prefixes strip before the base is tested.
		{"a responsive prefix", "export const c = 'flex md:ml-4';\n", 1},
		{"stacked prefixes", "export const c = 'flex md:hover:ml-4';\n", 1},
		// A negative value on a family the original allows one on.
		{"a negative margin", "export const c = 'flex -ml-4';\n", 1},
		// Several violations in one string report separately, matching the original's per-violation
		// report loop.
		{"two violations in one string", "export const c = 'ml-4 mr-4';\n", 2},
		// Reached through a JSX attribute, a call argument, and a bare array entry. The last is the
		// shape all three tree findings actually take, and the one a class-surface reader misses.
		{"a jsx attribute", "export const c = <div className=\"flex ml-4\" />;\n", 1},
		{"a call argument", "export const c = cn('flex ml-4');\n", 1},
		{"a bare array entry", "export const theme = [\n    'flex items-center',\n    'fixed left-[50%]',\n];\n", 1},
		// Template literals: the static head and the piece after an interpolation. A rule reading
		// only the head would pass the first and miss the second.
		{"a template with no holes", "export const c = `flex ml-4`;\n", 1},
		{"a template head", "export const c = `flex ml-4 ${other}`;\n", 1},
		{"a template tail after a hole", "export const c = `flex ${other} ml-4`;\n", 1},
		// The exact strings the tree carries, so the fixtures and the answer key agree.
		{"the dialog theme string", "export const c = 'fixed top-[50%] left-[50%] z-50 outline-none';\n", 1},
		{"the scroll area string", "export const c = 'before:absolute before:top-1/2 before:left-1/2 before:size-full';\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoPhysicalDirection, physicalDirectionFile, testCase.sourceText)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "useLogicalClass"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestNoPhysicalDirectionStaysSilent(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The logical classes the rule is asking for. These are the replacement text itself, so a
		// rule that matched its own output would loop forever under the fixer.
		{"logical margin", physicalDirectionFile, "export const c = 'flex ms-4 me-4';\n"},
		{"logical padding", physicalDirectionFile, "export const c = 'flex ps-4 pe-4';\n"},
		{"logical text alignment", physicalDirectionFile, "export const c = 'flex text-start text-end';\n"},
		{"logical border and rounded", physicalDirectionFile, "export const c = 'border-s-2 rounded-e-md';\n"},
		{"logical position", physicalDirectionFile, "export const c = 'flex start-0 end-0';\n"},
		// The direction-aware escape hatch, which is the one case where a physical class is right.
		{"an rtl prefix", physicalDirectionFile, "export const c = 'flex rtl:ml-4';\n"},
		{"an ltr prefix", physicalDirectionFile, "export const c = 'flex ltr:mr-4';\n"},
		{"a direction prefix among others", physicalDirectionFile, "export const c = 'flex md:rtl:ml-4';\n"},
		// Classes that begin with the same letters but are not the family. `mladder` shares three
		// characters with `ml-`, and a prefix test without the dash would match it.
		{"a class merely starting with the letters", physicalDirectionFile, "export const c = 'flex mlauto';\n"},
		{"a bare prefix with no value", physicalDirectionFile, "export const c = 'flex ml-';\n"},
		// Unrelated classes that contain the words.
		{"a class containing left elsewhere", physicalDirectionFile, "export const c = 'flex overflow-hidden';\n"},
		{"the word left in prose", physicalDirectionFile, "export const message = 'You have items left in your cart';\n"},
		{"a sentence", physicalDirectionFile, "export const message = 'Align the text to the left side';\n"},
		// Strings that are not classes at all, which is what the shape test exists for. The rule
		// visits every literal in the file, so these are the population it has to stay quiet on.
		// These three are what the shape test alone stands on. Each contains a token that IS a
		// physical-direction class by name, so the mapping table matches it, and each is rejected
		// because no token in the string passes the class-shape test. Without them, deleting the
		// shape test kills no fixture, and the one guard between this rule and 139,000 string
		// literals is unpinned. Mutation found exactly that.
		{"a documentation path naming a class", physicalDirectionFile, "export const link = 'See docs/text-left.md';\n"},
		{"a class name inside braces", physicalDirectionFile, "export const token = '{{ml-4}}';\n"},
		{"a class name with trailing punctuation", physicalDirectionFile, "export const note = 'ml-4;';\n"},
		{"a url", physicalDirectionFile, "export const url = 'https://example.com/left/right';\n"},
		{"a file path", physicalDirectionFile, "export const path = '/repository/source/Thing.ts';\n"},
		{"an identifier-like string", physicalDirectionFile, "export const key = 'someIdentifierName';\n"},
		{"an empty string", physicalDirectionFile, "export const c = '';\n"},
		// A negative on an exact-match family. The original's pattern for these has no negative
		// group, so `-text-left` is not a violation to the gate even though it looks like one.
		{"a negative on an exact family", physicalDirectionFile, "export const c = 'flex -text-left';\n"},
		// The file gate: only .ts and .tsx.
		{"a javascript file", "/repository/source/Thing.js", "export const c = 'flex ml-4';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoPhysicalDirection, testCase.fileName, testCase.sourceText))
		})
	}
}
