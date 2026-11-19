package oracle

import "testing"

// Scratch probe for #6c4swvt piece 4, never cut: tsc's 3076 shape through a readonly worker, honest and narrow. It
// prints what Node and each backend do, or the stop's exact words, and fails only to make the job show them.
func TestProbe3076(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"probe3076_honest.ts", "probe3076_narrow.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			node := onNode(t, path)
			t.Logf("PROBE %s node: %s", name, optionsBaseShown(node))
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s stops: %v", name, err)
				return
			}
			native, _ := natively(t, program)
			t.Logf("PROBE %s native: %s", name, optionsBaseShown(native))
			t.Logf("PROBE %s javascript: %s", name, optionsBaseShown(onJavaScriptBackend(t, program)))
		})
	}
}
