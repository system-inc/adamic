package lower

import (
	"errors"
	"strings"
	"testing"
)

// A view the program declares that names a lib-declared member the runtime shape lacks is NotYet at
// the relation, at any depth (#sn788gx, Adamic's ruling at 23:3xZ Nov 7). lib declares stack and cause
// on Error, and the runtime's Error shape holds name and message alone, so on main a real Error seen as
// { readonly stack?: string } read stack as absent natively, printing none where Node printed the stack,
// in .ts and .a both. A view naming the members the shape holds, and the Error type itself, still lower.
// Mutant: libraryUnstoredMember answering "" for every member, and every row lowers again, the silent none among them.
func TestLibraryUnstoredMemberViewIsNotYet(t *testing.T) {
	t.Parallel()
	const made = "const made = new Error(\"m\");\n"
	for _, probe := range []struct{ name, source, line, member string }{
		{"stack through a declared view", "interface Traced { readonly stack?: string }\n" + made + "const seen: Traced = made;\nconst text = seen.stack ?? \"none\";\nconsole.log(text.startsWith(\"Error: m\") ? \"stack\" : text);\n", ":3:", "stack"},
		{"stack one field deep", "interface Traced { readonly stack?: string }\ninterface Holder { readonly inner: Traced }\n" + made + "const holder: Holder = { inner: made };\nconsole.log(holder.inner.stack ?? \"none\");\n", ":4:", "stack"},
		{"stack through a parameter", "interface Traced { readonly stack?: string }\nfunction show(seen: Traced): string { return seen.stack ?? \"none\"; }\n" + made + "console.log(show(made));\n", ":4:", "stack"},
		{"cause through a declared view", "interface Caused { readonly cause?: unknown }\n" + made + "const seen: Caused = made;\nconsole.log(seen.cause === undefined ? \"none\" : \"cause\");\n", ":3:", "cause"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "stage 0 can't lower a view naming " + probe.member + ", which lib declares on Error but the runtime's Error doesn't hold (leave " + probe.member + " out of the view's type) yet"
				var pending *NotYet
				if !errors.As(err, &pending) || !strings.HasSuffix(err.Error(), want) || !strings.Contains(err.Error(), "main"+extension+probe.line) {
					t.Fatalf("want NotYet %q at %s, got %v", want, probe.line, err)
				}
			})
		}
	}
	// The members Error's shape holds, and Error seen as itself, still lower and agree with Node.
	lowersAndAgreesWithNodeAs(t, "interface Said { readonly name: string; readonly message: string }\n"+made+"const seen: Said = made;\nconsole.log(seen.name + \": \" + seen.message);\nconst same: Error = made;\nconsole.log(same.message);\n", ".a")
}
