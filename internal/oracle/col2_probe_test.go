package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

var col2Probes = map[string]string{
	"c01_class": "class Mine implements Intl.Collator {\n\tcompare(x: string, y: string): number {\n\t\treturn 2;\n\t}\n\tresolvedOptions(): Intl.ResolvedCollatorOptions {\n\t\treturn { locale: 'en', usage: 'sort', sensitivity: 'variant', ignorePunctuation: false, collation: 'default', caseFirst: 'false', numeric: true };\n\t}\n}\nconst collator: Intl.Collator = new Mine();\nconsole.log(String(collator.compare('a', 'b')));\n",
	"c02_literal": "const collator: Intl.Collator = {\n\tcompare: (x: string, y: string): number => 2,\n\tresolvedOptions: (): Intl.ResolvedCollatorOptions => ({ locale: 'en', usage: 'sort', sensitivity: 'variant', ignorePunctuation: false, collation: 'default', caseFirst: 'false', numeric: true }),\n};\nconsole.log(String(collator.compare('a', 'b')));\n",
	"c03_parameter": "class Mine {\n\tcompare(x: string, y: string): number {\n\t\treturn 2;\n\t}\n\tresolvedOptions(): Intl.ResolvedCollatorOptions {\n\t\treturn { locale: 'en', usage: 'sort', sensitivity: 'variant', ignorePunctuation: false, collation: 'default', caseFirst: 'false', numeric: true };\n\t}\n}\nfunction use(collator: Intl.Collator): number {\n\treturn collator.compare('a', 'b');\n}\nconsole.log(String(use(new Mine())));\n",
	"c04_literal_untyped": "const mine = {\n\tcompare: (x: string, y: string): number => 2,\n\tresolvedOptions: (): Intl.ResolvedCollatorOptions => ({ locale: 'en', usage: 'sort', sensitivity: 'variant', ignorePunctuation: false, collation: 'default', caseFirst: 'false', numeric: true }),\n};\nfunction use(collator: Intl.Collator): number {\n\treturn collator.compare('a', 'b');\n}\nconsole.log(String(use(mine)));\n",
	"h_interface": "interface Sorter {\n\tcompare(x: string, y: string): number;\n}\nconst collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nconst sorter: Sorter = collator;\nconsole.log(String(sorter.compare('a', 'b')));\n",
	"h_field": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nconst holder: { readonly sorter: Intl.Collator } = { sorter: collator };\nconsole.log(String(holder.sorter.compare('a', 'b')));\n",
	"h_generic": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nfunction use<T extends Intl.Collator>(sorter: T): number {\n\treturn sorter.compare('a', 'b');\n}\nconsole.log(String(use(collator)));\n",
	"h_identity": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nfunction same<T>(value: T): T {\n\treturn value;\n}\nconsole.log(String(same(collator).compare('a', 'b')));\n",
	"k_string": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nconsole.log(String(collator));\n",
	"k_template": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nconsole.log(`${collator}`);\n",
	"k_destructure": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nconst { compare } = collator;\nconsole.log(String(compare('a', 'b')));\n",
	"k_json": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nconsole.log(JSON.stringify(collator) ?? 'undefined');\n",
	"n_alias": "const C = Intl.Collator;\nconst collator = new C(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log(typeof collator);\n",
	"n_keyed": "const collator = new Intl['Collator'](undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log(typeof collator);\n",
	"n_global": "const collator = new globalThis.Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log(typeof collator);\n",
	"n_instanceof": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nconst value: unknown = collator;\nif (value instanceof Intl.Collator) {\n\tconsole.log(String(value.compare('a', 'b')));\n}\n",
	"e_empty": "console.log('before');\nconst collator = new Intl.Collator('', { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log(`collator ${typeof collator}`);\n",
	"o_order": "const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });\nconsole.log('collator');\nfunction first(): string {\n\tconsole.log('first');\n\treturn 'a10';\n}\nfunction second(): string {\n\tconsole.log('second');\n\treturn 'a2';\n}\nconst result = collator.compare(first(), second());\nconsole.log(`compared ${result < 0 ? 'before' : result > 0 ? 'after' : 'equal'}`);\n",
}

func TestCol2Probe(t *testing.T) {
	t.Parallel()
	for name, source := range col2Probes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/intl_collator/probe", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Logf("PROBE %s start", name)
			node := onNode(t, path)
			t.Logf("PROBE %s node: exit %d stdout %q stderr %.300q", name, node.exitCode, node.stdout, node.stderr)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s lower: %T %v", name, err, err)
				return
			}
			native, _ := natively(t, program)
			t.Logf("PROBE %s native: exit %d stdout %q stderr %.300q", name, native.exitCode, native.stdout, native.stderr)
			backend := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s javascript: exit %d stdout %q stderr %.300q", name, backend.exitCode, backend.stdout, backend.stderr)
		})
	}
}
