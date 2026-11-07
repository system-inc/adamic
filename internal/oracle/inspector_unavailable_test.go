package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/hostconfiguration"
)

// CommonJS measurement, separate from the source oracle's ES-module outcome.
func TestInspectorUnavailableIsNodes(test *testing.T) {
	test.Parallel()
	configuration := &hostconfiguration.Configuration{
		Name:      "node-without-inspector",
		Node:      "v24.19.0",
		Configure: "--without-inspector",
		Ruling:    "#cn4yapb",
	}
	if why := hostconfiguration.Skip(1, configuration); why != "" {
		test.Skipf(
			"inspector measurement row %d: the CommonJS load needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
			1,
		)
	}
	binary := hostconfiguration.Require(test, 1, configuration)
	probe := filepath.Join(test.TempDir(), "inspector.cjs")
	source := `for (const name of ['inspector','node:inspector','inspector/promises']) {
 try { require(name); console.log('unexpected success'); }
 catch (error) { console.log([error.name,error.code,error.message,String(error),Object.keys(error).join(',')].join('\n')); }
 }`
	if writeError := os.WriteFile(probe, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	actual := executeWith(test, []string{"NODE_PATH="}, binary, probe)
	row := strings.Join([]string{
		inspectorUnavailableName,
		inspectorUnavailableCode,
		inspectorUnavailableMessage,
		inspectorUnavailableString,
		"code",
		"",
	}, "\n")
	if actual.exitCode != 0 || len(actual.stderr) != 0 || string(actual.stdout) != strings.Repeat(row, 3) {
		test.Fatalf(
			"configured inspector load: exit %d stdout %q stderr %q; want %q",
			actual.exitCode,
			actual.stdout,
			actual.stderr,
			strings.Repeat(row, 3),
		)
	}
}

// This pin runs on stock runners; only the independent CommonJS measurement opts in.
func TestInspectorUnavailableBackends(test *testing.T) {
	test.Parallel()
	for _, suffix := range []string{"", "_uncaught"} {
		test.Run("node_inspector_unavailable"+suffix, func(test *testing.T) {
			fixture := "internal/oracle/testdata/node_inspector_unavailable" + suffix + ".a"
			path := filepath.Join(repository, fixture)
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			truth := onSliceNode(test, path)
			actual, _ := nativelyUncached(test, program)
			if why := sourceDisagreement(fixture, truth, actual, false); why != "" {
				test.Fatal(why)
			}
			actual = onSliceJavaScript(test, program)
			if why := sourceDisagreement(fixture, truth, actual, true); why != "" {
				test.Fatal(why)
			}
		})
	}
}
