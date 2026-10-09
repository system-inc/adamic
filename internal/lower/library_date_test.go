package lower

import (
	"strings"
	"testing"
)

func TestDateRuling10Refusals(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`new Date(0).toString();`, `new Date(0).toDateString();`, `new Date(0).toTimeString();`,
		`new Date(0).toLocaleString();`, `new Date(0).toLocaleDateString();`, `new Date(0).toLocaleTimeString("en-US",{timeZone:"UTC"});`,
		`new Date(0).getTimezoneOffset();`, `new Date(0).getHours();`, `new Date(0).setHours(1);`,
		`Date.prototype.getFullYear.call(new Date(0));`, `new Date(0)["toString"]();`,
		`Date.parse("December 4, 1995");`, `Date.parse("2020-01-01T00:00:00");`, `Date.parse("-000000-01-01");`,
		`function parse(value:string):number {return Date.parse(value);}`, `new Date("1/1/2020");`, `new Date(2020,0);`, `Date();`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			if err == nil || !strings.Contains(err.Error(), "ruling 10") {
				t.Fatalf("want ruling 10 refusal, got %v", err)
			}
		})
	}
}
func TestDateInternalSlotAndValidationBoundaries(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`Date.prototype.getTime.call({});`,
		`const date=new Date(0); const view:{valueOf:()=>number}=date; view.valueOf();`,
		`try {new Date(NaN).toISOString();} catch (error:unknown) {console.log('caught');}`,
		`const method=Date.prototype.getTime;`,
		`const date = new Date(0); date.getTime = (): number => 123;`,
		`Date.now = (): number => 0;`,
		`Date.prototype.valueOf = (): number => 42;`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			if err == nil {
				t.Fatal("unrepresented boundary compiled")
			}
		})
	}
}

// Not parallel: changes the process-wide TZ environment variable with t.Setenv.
func TestDateUTCFormsDoNotRequirePinnedBuildTimezone(t *testing.T) {
	t.Setenv("TZ", "America/New_York")
	for _, source := range []string{
		`Date.parse("2020-01-01T00:00:00Z");`, `new Date(0).toISOString();`, `new Date(0).toUTCString();`,
		`new Date(0).getUTCHours();`, `Date.UTC(2020,0);`, `const clock=Date.now;clock();`,
		`const left=new Date(0); const right=new Date(1); const before=left<right;`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("%s: %v", source, err)
		}
	}
}
