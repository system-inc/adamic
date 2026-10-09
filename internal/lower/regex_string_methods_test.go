package lower

import "testing"

func TestRegexStringCatchAndCall(t *testing.T) {
	for _, source := range []string{
		`function fail(): void { 'a'.matchAll(/a/); } try { fail(); } catch (error) { if (error instanceof Error) { console.log(error.message); } }`,
		`function fail(): void { 'a'.replaceAll(/a/, 'b'); } try { fail(); } catch {} finally { console.log('finally'); }`,
		`console.log(String.prototype.split.call('aba', /a/g).join('|'));`,
		`String.prototype.match.call('aba', /a/g);`,
		`String.prototype.matchAll.call('aba', /a/g).next();`,
		`String.prototype.search.call('aba', /a/g);`,
		`console.log(String.prototype.replace.call<string, [RegExp, string], string>('aba', /a/g, '!'));`,
		`console.log(String.prototype.replaceAll.call<string, [RegExp, string], string>('aba', /a/g, '!'));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("must lower %s: %v", source, err)
		}
	}
}
