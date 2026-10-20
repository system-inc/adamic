package native

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// A framework directory resolves <Probe/probe.h> to Probe.framework/Headers/probe.h, or PrivateHeaders, under it: a
// framework planted in an earlier -iframework directory shadows a later one, and a framework header a __has_include
// asks after moves the key when it's planted. macOS's clang lists -iframework directories as framework directories.
// It's macOS's alone, so it lives in a _darwin file rather than skipping elsewhere, a skip the census would refuse
// (chg_42qt811t's red).
func TestUnitKeyHoldsFrameworks(t *testing.T) {
	t.Parallel()
	compiler := unitKeyCompiler(t)
	environment := unitKeyEnvironment()
	for _, plan := range []struct {
		name, header, plant, value string
		before, after              string
	}{
		{"Shadowing", "#include <Probe/probe.h>\n", "earlier/Probe.framework/Headers/probe.h", "PROBE", "1", "2"},
		{"HasInclude", "#if __has_include(<Probe/maybe.h>)\n#define MAYBE 2\n#else\n#define MAYBE 0\n#endif\n", "later/Probe.framework/PrivateHeaders/maybe.h", "MAYBE", "0", "2"},
	} {
		t.Run(plan.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeUnitKeyFile(t, filepath.Join(root, "later", "Probe.framework", "Headers", "probe.h"), "#define PROBE 1\n")
			if err := os.MkdirAll(filepath.Join(root, "earlier"), 0o755); err != nil {
				t.Fatal(err)
			}
			common := []runtimeFile{{"units.h", []byte(plan.header)}}
			flags := append(slices.Clone(unitKeyFlags), "-iframework", filepath.Join(root, "earlier"), "-iframework", filepath.Join(root, "later"))
			unit := unitKeyProbe(plan.value)
			keying := testUnitKeying(t, unitFirst, common, flags, compiler, environment)
			directory := t.TempDir()
			if got := observeUnitKey(t, compiler, compileUnitKey(t, unit, common, flags, compiler, keying, directory)); got != plan.before {
				t.Fatalf("before the plant the unit returned %s, want %s", got, plan.before)
			}
			writeUnitKeyFile(t, filepath.Join(root, plan.plant), "#define "+plan.value+" 2\n")
			planted := testUnitKeying(t, unitFirst, common, flags, compiler, environment)
			if planted.key(unit) == keying.key(unit) {
				t.Fatalf("%s was planted and the key held", plan.plant)
			}
			if got := observeUnitKey(t, compiler, compileUnitKey(t, unit, common, flags, compiler, planted, directory)); got != plan.after {
				t.Fatalf("after the plant the unit returned %s, want %s", got, plan.after)
			}
		})
	}
}
