package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewDictionaryKeys(t *testing.T) {
	for _, shape := range []string{"fixed", "producer"} {
		for _, name := range []string{"good", "empty", "read-wrong"} {
			t.Run(shape+"-"+name, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/keys-"+shape+"-"+name)
				truth := onNode(t, path)
				want := "2\n10\nz\na\n"
				if name == "empty" {
					want = ""
				}
				if name == "read-wrong" {
					want += "42\n"
				}
				if truth.exitCode != 0 || string(truth.stdout) != want {
					t.Fatalf("Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if name == "read-wrong" {
						diagnostic := "adamic: panic: field read failed: view.table['z']; expected string | undefined, found number\n"
						if got.exitCode != 70 || string(got.stderr) != diagnostic || string(got.stdout) != "2\n10\nz\na\n" {
							t.Fatalf("lazy enumeration/lookup: %#v", got)
						}
					} else if d := disagreement(truth, got); d != "" {
						t.Fatal(d)
					}
				}
				if name != "read-wrong" {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

func TestCheckedViewDictionaryKeysEagerMutant(t *testing.T) {
	program, path := interfaceFixture(t, "dictionaries/source/keys-producer-good")
	truth := onNode(t, path)
	c := native.C(program)
	call := "adamic_view_dictionary_source_keys("
	if !strings.Contains(c, call) {
		t.Fatal("no keys hook")
	}
	wrapper := `static adamic_array *keys_mutant(const adamic_object *object) {
 adamic_array *keys=adamic_view_dictionary_source_keys(object);
 for(size_t i=0;i<keys->length;i++) (void)adamic_view_dictionary_source_read(object,keys->elements[i].reference,1u << adamic_view_union_string,0,"Object.keys[key]","string");
 return keys;
 }`
	c = "#include \"view_dictionaries.h\"\n" + wrapper + "\n" + strings.ReplaceAll(c, call, "keys_mutant(")
	binary := filepath.Join(t.TempDir(), "keys-mutant")
	if err := native.Build(c, binary, native.Options{}); err != nil {
		t.Fatal("semantic mutant must compile: ", err)
	}
	code := javascript.JavaScript(program)
	if !strings.Contains(code, "Object.keys(") {
		t.Fatal("no JS keys hook")
	}
	js := filepath.Join(t.TempDir(), "keys-mutant.mjs")
	code = `const keysMutant=object=>{const keys=Object.keys(object);for(const key of keys)adamicViewDictionaryRead(object,key,['string'],0,'Object.keys[key]','string');return keys;};` + "\n" + strings.ReplaceAll(code, "Object.keys(", "keysMutant(")
	if err := os.WriteFile(js, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{execute(t, binary), onNode(t, js)} {
		if got.exitCode != 70 || disagreement(truth, got) == "" {
			t.Fatalf("eager element check escaped: %#v", got)
		}
		t.Logf("key-only control catches eager element mutant: exit=%d stderr=%q", got.exitCode, got.stderr)
	}
}
