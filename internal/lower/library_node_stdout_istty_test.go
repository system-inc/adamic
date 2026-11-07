package lower

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNodeStdoutIsTTYReadsAreAdmitted(test *testing.T) {
	test.Parallel()
	for _, ext := range []string{"a", "ts"} {
		source, readError := os.ReadFile(filepath.Join("testdata/stdout", "isTTY."+ext))
		if readError != nil {
			test.Fatal(readError)
		}
		if _, lowerSourceAsError := lowerSourceAs(test, string(source), "."+ext); lowerSourceAsError != nil {
			test.Fatal(lowerSourceAsError)
		}
	}
	for _, source := range []string{
		`import process from 'node:process'; const pretty = !!process.stdout.isTTY;`,
		`import host from 'node:process'; const missing = host.stdout.isTTY === undefined;`,
		`import process from 'node:process'; const tty = (process.stdout).isTTY;`,
	} {
		if _, lowerSourceError := lowerSource(test, source); lowerSourceError != nil {
			test.Fatal(lowerSourceError)
		}
	}
}

func TestNodeStdoutIsTTYRefusesOtherStreams(test *testing.T) {
	test.Parallel()
	for _, stream := range []string{"stderr", "stdin"} {
		for _, ext := range []string{".a", ".ts"} {
			_, lowerSourceAsError := lowerSourceAs(test, "import process from 'node:process'; const value=process."+
				stream+
				".isTTY; console.log(value === undefined ? 'undefined' : 'true');", ext)
			var ny *NotYet
			want := "node:process.Process." + stream + ".isTTY is not bound (fix: read process.stdout.isTTY)"
			if !errors.As(lowerSourceAsError, &ny) || ny.What != want {
				test.Fatalf("want %q, got %v", want, lowerSourceAsError)
			}
		}
	}
}

func TestNodeStdoutIsTTYRefusesOpaqueUses(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`const value = process.stdout['isTTY'];`,
		`const { isTTY } = process.stdout;`,
	} {
		for _, ext := range []string{".a", ".ts"} {
			test.Run(source+ext, func(test *testing.T) {
				_, lowerSourceAsError := lowerSourceAs(test, "import process from 'node:process';"+source, ext)
				var ny *NotYet
				want := "node:process.Process.stdout used as a value"
				if !errors.As(lowerSourceAsError, &ny) || ny.What != want {
					test.Fatalf("want %q, got %v", want, lowerSourceAsError)
				}
			})
		}
	}
}

func TestNodeStdoutIsTTYRefusesMutation(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ source, reason string }{
		{`process.stdout.isTTY = true;`, "written (fix: read process.stdout.isTTY; terminal status is read-only)"},
		{`process.stdout.isTTY ||= true;`, "written (fix: read process.stdout.isTTY; terminal status is read-only)"},
		{`delete process.stdout.isTTY;`, "deleted (fix: read process.stdout.isTTY; terminal status is read-only)"},
		{
			`console.log(process.stdout?.isTTY);`,
			"read through an optional chain (fix: read process.stdout.isTTY directly)",
		},
	} {
		for _, ext := range []string{".a", ".ts"} {
			_, lowerSourceAsError := lowerSourceAs(test, "import process from 'node:process';"+one.source, ext)
			var ny *NotYet
			want := "node:process.Process.stdout.isTTY " + one.reason
			if !errors.As(lowerSourceAsError, &ny) || ny.What != want {
				test.Fatalf("%s: want %q, got %v", one.source, want, lowerSourceAsError)
			}
		}
	}
}
