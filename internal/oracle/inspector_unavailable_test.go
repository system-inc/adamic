package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/hostconfiguration"
)

// CommonJS measurement, separate from the source oracle's ES-module outcome.
func TestInspectorUnavailableIsNodes(t *testing.T) {
	t.Parallel()
	configuration := &hostconfiguration.Configuration{Name: "node-without-inspector", Node: "v24.19.0", Configure: "--without-inspector", Ruling: "#cn4yapb"}
	if why := hostconfiguration.Skip(1, configuration); why != "" {
		t.Skipf("inspector measurement row %d: the CommonJS load needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR", 1)
	}
	binary := hostconfiguration.Require(t, 1, configuration)
	probe := filepath.Join(t.TempDir(), "inspector.cjs")
	source := `for (const name of ['inspector','node:inspector','inspector/promises']) {
 try { require(name); console.log('unexpected success'); }
 catch (error) { console.log([error.name,error.code,error.message,String(error),Object.keys(error).join(',')].join('\n')); }
 }`
	if err := os.WriteFile(probe, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"NODE_PATH="}, binary, probe)
	row := strings.Join([]string{inspectorUnavailableName, inspectorUnavailableCode, inspectorUnavailableMessage, inspectorUnavailableString, "code", ""}, "\n")
	if actual.exitCode != 0 || len(actual.stderr) != 0 || string(actual.stdout) != strings.Repeat(row, 3) {
		t.Fatalf("configured inspector load: exit %d stdout %q stderr %q; want %q", actual.exitCode, actual.stdout, actual.stderr, strings.Repeat(row, 3))
	}
}

// This pin runs on stock runners; only the independent CommonJS measurement opts in.
func TestInspectorUnavailableBackends(t *testing.T) {
	t.Parallel()
	for _, suffix := range []string{"", "_uncaught"} {
		t.Run("node_inspector_unavailable"+suffix, func(t *testing.T) {
			fixture := "internal/oracle/testdata/node_inspector_unavailable" + suffix + ".a"
			path := filepath.Join(repository, fixture)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onSliceNode(t, path)
			actual, _ := nativelyUncached(t, program)
			if why := sourceDisagreement(fixture, truth, actual, false); why != "" {
				t.Fatal(why)
			}
			actual = onSliceJavaScript(t, program)
			if why := sourceDisagreement(fixture, truth, actual, true); why != "" {
				t.Fatal(why)
			}
		})
	}
}
