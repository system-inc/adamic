package lower

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const phantomPresenceFix = "read the brand member's undefined value without testing its presence; use a separate real field or Map for observable presence"
const phantomWriteFix = "keep the brand member phantom and unwritten; store observable state in a separate real field or Map"

func TestPhantomArrayPresenceRefusals(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"marker: void", "marker: undefined", "marker?: undefined"} {
		for _, probe := range []struct{ name, body string }{
			{"in", "'marker' in a;"},
			{"Object.hasOwn", "Object.hasOwn(a, 'marker');"},
			{"hasOwnProperty", "a.hasOwnProperty('marker');"},
			{"Object.keys", "Object.keys(a);"},
			{"Object.values", "Object.values(a);"},
			{"Object.entries", "Object.entries(a);"},
			{"Object.getOwnPropertyNames", "Object.getOwnPropertyNames(a);"},
			{"object spread", "const copied = {...a};"},
			{"Object.assign", "Object.assign({}, a);"},
			{"JSON.stringify", "JSON.stringify(a);"},
		} {
			t.Run(field+"/"+probe.name, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSource(t, "type Brand = readonly number[] & { "+field+" }; function observe(a: Brand): void { "+probe.body+" }")
				wantPhantomRefusal(t, err, "presence of phantom array brand member marker via "+probe.name, phantomPresenceFix)
			})
		}
		for _, body := range []string{"a.marker = undefined;", "a['marker'] = undefined;", "a.marker ??= undefined;", "[a.marker] = [undefined];", "Object.assign(a, {marker: undefined});"} {
			t.Run(field+"/write/"+body, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSource(t, "type Brand = readonly number[] & { "+field+" }; function write(a: Brand): void { "+body+" }")
				wantPhantomRefusal(t, err, "a write to phantom array brand member marker", phantomWriteFix)
			})
		}
	}
}

func TestPhantomArrayPresenceNarrowing(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../review/phantom-brands/array-presence-in.a")
	if err != nil {
		t.Fatal(err)
	}
	path, err := filepath.Abs("../../review/phantom-brands/array-presence-in.a")
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path).CombinedOutput()
	if err != nil || string(output) != "absent\n" {
		t.Fatalf("Node: %v: %q", err, output)
	}
	_, err = lowerSource(t, string(source))
	wantPhantomRefusal(t, err, "presence of phantom array brand member  __sortedArrayBrand via in", phantomPresenceFix)
}

func wantPhantomRefusal(t *testing.T, err error, what, fix string) {
	t.Helper()
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != what || refused.Fix != fix {
		t.Fatalf("got %v, want Adamic 0.1 refuses %s; %s", err, what, fix)
	}
}

func TestPhantomArrayPresenceViews(t *testing.T) {
	t.Parallel()
	for _, declaration := range []string{
		"interface Brand extends ReadonlyArray<number> { marker: undefined; }",
		"interface Brand extends ReadonlyArray<number> { marker: void; }",
		"interface Brand extends ReadonlyArray<number> { marker?: undefined; }",
	} {
		for _, parameter := range []string{"a: Brand", "a: Brand | readonly number[]", "a: Item"} {
			t.Run(declaration+parameter, func(t *testing.T) {
				t.Parallel()
				generic := ""
				if parameter == "a: Item" {
					generic = "<Item extends Brand>"
				}
				_, err := lowerSource(t, declaration+" function observe"+generic+"("+parameter+"): void { Object.keys(a); }")
				wantPhantomRefusal(t, err, "presence of phantom array brand member marker via Object.keys", phantomPresenceFix)
			})
		}
	}
}

func TestPhantomArrayComputedPresence(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, body string }{
		{"Object.hasOwn", "const key = 'marker'; Object['hasOwn'](a, key);"},
		{"hasOwnProperty", "a['hasOwnProperty']('marker');"},
		{"propertyIsEnumerable", "a.propertyIsEnumerable('marker');"},
		{"Object.keys", "(Object).keys(a);"},
		{"hasOwnProperty", "function key(): string { return 'marker'; } a.hasOwnProperty(key());"},
		{"Object.assign", "Object.assign({}, {}, a);"},
	} {
		t.Run(probe.body, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, "type Brand = readonly number[] & {marker: undefined}; function observe(a: Brand): void { "+probe.body+" }")
			wantPhantomRefusal(t, err, "presence of phantom array brand member marker via "+probe.name, phantomPresenceFix)
		})
	}
}

func TestPhantomArrayWritesNameTheMember(t *testing.T) {
	t.Parallel()
	for _, body := range []string{"a.second = undefined;", "a['second'] = undefined;", "({value: a.second} = {value: undefined});"} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, "type Brand = number[] & {first: void; second: undefined}; function write(a: Brand): void { "+body+" }")
			wantPhantomRefusal(t, err, "a write to phantom array brand member second", phantomWriteFix)
		})
	}
}

func TestPhantomArrayRealMembersRemainUsable(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "type Brand = number[] & {marker: undefined}; function use(a: Brand): void { a[0] = 1; a.hasOwnProperty('length'); }")
	if err != nil {
		t.Fatal(err)
	}
}
