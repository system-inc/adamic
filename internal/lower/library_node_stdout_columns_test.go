package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"testing"
)

func TestNodeStdoutColumnsReadAdmission(test *testing.T) {
	test.Parallel()
	for _, ext := range []string{"a", "ts"} {
		source, readError := os.ReadFile(filepath.Join("testdata/stdout", "columns."+ext))
		if readError != nil {
			test.Fatal(readError)
		}
		program, lowerSourceAsError := lowerSourceAs(test, string(source), "."+ext)
		if lowerSourceAsError != nil {
			test.Fatal(lowerSourceAsError)
		}
		_ = program
	}
	for _, source := range []string{
		`import process from 'node:process'; const width = process.stdout.columns ?? 0;`,
		`import host from 'node:process'; const missing = (host.stdout).columns === undefined;`,
	} {
		if _, lowerSourceError := lowerSource(test, source); lowerSourceError != nil {
			test.Fatal(lowerSourceError)
		}
	}
	if (ir.NodeFSFile{Operation: "host_stdout_columns", Of: ir.MaybeNumber}).MayThrow() {
		test.Fatal("columns read throws")
	}
}

func TestNodeStdoutColumnsRefusesOtherUses(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ source, reason string }{
		{
			`process.stdout.columns = 80;`,
			"node:process.Process.stdout.columns written (fix: read process.stdout.columns; terminal status is read-only)",
		},
		{
			`process.stdout.columns ||= 80;`,
			"node:process.Process.stdout.columns written (fix: read process.stdout.columns; terminal status is read-only)",
		},
		{
			`delete process.stdout.columns;`,
			"node:process.Process.stdout.columns deleted (fix: read process.stdout.columns; terminal status is read-only)",
		},
		{
			`const value = process.stdout?.columns;`,
			"node:process.Process.stdout.columns read through an optional chain (fix: read process.stdout.columns directly)",
		},
		{
			`const value = process.stderr.columns;`,
			"node:tty.WriteStream.columns on another stream (fix: read process.stdout.columns)",
		},
		{
			`const value = process.stdout.rows;`,
			"node:tty.WriteStream.rows is not bound (fix: use process.stdout.columns for terminal width)",
		},
		{`const value = process.stdout['columns'];`, "node:process.Process.stdout used as a value"},
		{`const {columns} = process.stdout;`, "node:process.Process.stdout used as a value"},
	} {
		for _, ext := range []string{".a", ".ts"} {
			test.Run(one.source+ext, func(test *testing.T) {
				_, lowerSourceAsError := lowerSourceAs(test, "import process from 'node:process';"+one.source, ext)
				var ny *NotYet
				if !errors.As(lowerSourceAsError, &ny) || ny.What != one.reason {
					test.Fatalf("want %q, got %v", one.reason, lowerSourceAsError)
				}
			})
		}
	}
}
