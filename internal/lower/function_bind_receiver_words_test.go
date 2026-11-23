package lower

import (
	"errors"
	"strings"
	"testing"
)

// TestBindReceiverWordsNameWhatTheReceiverIs holds the stop for a method bind whose receiver isn't a const or
// a never-assigned parameter: it says what the receiver is, never that it was assigned when it wasn't
// (#d6jncyx, Adamic on #bbssdcy, family 66; tsc's builder.ts:1750, :1838 and :1937 bind a destructured host).
func TestBindReceiverWordsNameWhatTheReceiverIs(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, words string }{
		{"destructured", "class Writer {\n\twrite(): string { return \"written\"; }\n}\nfunction bind({ host }: { host: Writer }): () => string {\n\treturn host.write.bind(host);\n}\nconsole.log(bind({ host: new Writer() })());\n",
			"method bind whose receiver host isn't a const or a parameter the function never assigns (a destructured binding)"},
		{"let", "class Writer {\n\twrite(): string { return \"written\"; }\n}\nlet host = new Writer();\nconst bound = host.write.bind(host);\nconsole.log(bound());\n",
			"method bind whose receiver host isn't a const or a parameter the function never assigns (a let or var)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var later *NotYet
			if !errors.As(err, &later) || later.What != probe.words {
				t.Fatalf("got %v, want the NotYet %q", err, probe.words)
			}
			if strings.Contains(err.Error(), "assigned receiver") {
				t.Fatalf("the old words are back: %v", err)
			}
		})
	}
}
