package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The ruled guard needs every other own-enumerability road closed too. Both backends use this
// lowering pass; an admitted road here needs a new representation proof before either may run it.
func TestJSONErrorEnumerableRoadsRefuse(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, body, stdout, refusal string }{
		{"define_name", "Object.defineProperty(e, 'name', {value: 'X', enumerable: true});", "{\"name\":\"X\"}\n", "Adamic 0.1 refuses Object.defineProperty; property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores"},
		{"define_message", "Object.defineProperty(e, 'message', {value: 'y', enumerable: true});", "{\"message\":\"y\"}\n", "Adamic 0.1 refuses Object.defineProperty; property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores"},
		{"assign_name", "Object.assign(e, {name: 'X'});", "{\"name\":\"X\"}\n", "stage 0 can't lower Object.assign on a shape not proven by a plain literal or its const binding yet"},
		{"spread", "const copy = {...e}; console.log(`${copy instanceof Error}`); console.log(JSON.stringify(copy) ?? '<undefined>');", "false\n{}\n{}\n", "stage 0 can't lower object spread of Error storage with non-enumerable properties; copy the desired fields explicitly yet"},
		{"reflect_set", "Reflect.set(e, 'name', 'X');", "{\"name\":\"X\"}\n", "stage 0 can't lower a namespace member without a lowered binding yet"},
		{"reflect_define", "Reflect.defineProperty(e, 'name', {value: 'X', enumerable: true});", "{\"name\":\"X\"}\n", "stage 0 can't lower a namespace member without a lowered binding yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), probe.name+".a")
			source := "const e = new Error('x');\n" + probe.body + "\nconsole.log(JSON.stringify(e) ?? '<undefined>');\n"
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if why := disagreement(run{stdout: []byte(probe.stdout)}, truth); why != "" {
				t.Fatalf("Node: %s: %#v", why, truth)
			}
			_, err := lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), probe.refusal) {
				t.Fatalf("own-enumerability road lost its refusal %q: %v", probe.refusal, err)
			}
			t.Log(err)
		})
	}
}
