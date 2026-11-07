package buildcache

import "testing"

// A covered test binary's build info holds one setting a plain one doesn't, `-cover=true`, and a product key never
// holds it, so a plain driver asks under the key its covered owner settled. Only that exact setting is dropped: one
// that only starts like it still moves the key, as every other setting does. Mutants: the strip dropped (a covered
// binary's settings differ from a plain one's); the strip widened to any setting starting with `-cover` (the
// lookalike no longer moves the value).
func TestACoveredBinarysSettingsValueAsAPlainOnes(test *testing.T) {
	test.Parallel()
	plain := []string{
		"-buildmode=exe",
		"-compiler=gc",
		"-trimpath=true",
		"CGO_ENABLED=1",
		"GOARCH=amd64",
		"GOOS=linux",
		"vcs.revision=abc",
	}
	covered := append([]string{"-cover=true"}, plain...)
	if settingsValue(covered) != settingsValue(plain) {
		test.Fatalf(
			"a covered binary's settings value %q, a plain one's %q: want them equal",
			settingsValue(covered),
			settingsValue(plain),
		)
	}
	for _, other := range []string{"-cover=false", "-covermode=set", "-coverpkg=example.com/m", "-race=true"} {
		if settingsValue(append([]string{other}, plain...)) == settingsValue(plain) {
			test.Errorf(
				"the setting %q left the settings value as a plain binary's: want only %q dropped",
				other,
				coverageSetting,
			)
		}
	}
}
