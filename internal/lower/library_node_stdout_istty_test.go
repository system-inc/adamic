package lower

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNodeStdoutIsTTYReadsAreAdmitted(t *testing.T) {
	t.Parallel()
	for _, ext := range []string{"a", "ts"} {
		source, err := os.ReadFile(filepath.Join("testdata/stdout", "isTTY."+ext))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := lowerSourceAs(t, string(source), "."+ext); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{
		`import process from 'node:process'; const pretty = !!process.stdout.isTTY;`,
		`import host from 'node:process'; const missing = host.stdout.isTTY === undefined;`,
		`import process from 'node:process'; const tty = (process.stdout).isTTY;`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNodeStdoutIsTTYRefusesOtherStreams(t *testing.T) {
	t.Parallel()
	for _, stream := range []string{"stderr", "stdin"} {
		for _, ext := range []string{".a", ".ts"} {
			_, err := lowerSourceAs(t, "import process from 'node:process'; const value=process."+stream+".isTTY; console.log(value === undefined ? 'undefined' : 'true');", ext)
			var ny *NotYet
			want := "node:process.Process." + stream + ".isTTY is not bound (fix: read process.stdout.isTTY)"
			if !errors.As(err, &ny) || ny.What != want {
				t.Fatalf("want %q, got %v", want, err)
			}
		}
	}
}

func TestNodeStdoutIsTTYRefusesOpaqueUses(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const value = process.stdout['isTTY'];`,
		`const { isTTY } = process.stdout;`,
	} {
		for _, ext := range []string{".a", ".ts"} {
			t.Run(source+ext, func(t *testing.T) {
				_, err := lowerSourceAs(t, "import process from 'node:process';"+source, ext)
				var ny *NotYet
				want := "node:process.Process.stdout used as a value"
				if !errors.As(err, &ny) || ny.What != want {
					t.Fatalf("want %q, got %v", want, err)
				}
			})
		}
	}
}

func TestNodeStdoutIsTTYRefusesMutation(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ source, reason string }{
		{`process.stdout.isTTY = true;`, "written (fix: read process.stdout.isTTY; terminal status is read-only)"},
		{`process.stdout.isTTY ||= true;`, "written (fix: read process.stdout.isTTY; terminal status is read-only)"},
		{`delete process.stdout.isTTY;`, "deleted (fix: read process.stdout.isTTY; terminal status is read-only)"},
		{`console.log(process.stdout?.isTTY);`, "read through an optional chain (fix: read process.stdout.isTTY directly)"},
	} {
		for _, ext := range []string{".a", ".ts"} {
			_, err := lowerSourceAs(t, "import process from 'node:process';"+one.source, ext)
			var ny *NotYet
			want := "node:process.Process.stdout.isTTY " + one.reason
			if !errors.As(err, &ny) || ny.What != want {
				t.Fatalf("%s: want %q, got %v", one.source, want, err)
			}
		}
	}
}
