package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

func TestModuleNamespaceLimitsStayLoud(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ source, provider, reason string }{
		{"const escaped = performance;", "export const value = 7;", "ESM namespace object"},
		{"console.log(Object.keys(performance).join(','));", "export const value = 7;", "ESM namespace object"},
		{"console.log(`${performance['value']}`);", "export const value = 7;", "ESM namespace object"},
		{"performance.value = 2;", "export const value = 7;", ""}, // The checker rejects an ESM export write.
		{
			"console.log(`${performance.read()}`);",
			"export function read(this: { value: number }): number { return this.value; } export const value = 7;",
			"observing its receiver",
		},
	} {
		test.Run(probe.source, func(test *testing.T) {
			test.Parallel()
			directory := test.TempDir()
			entry := filepath.Join(directory, "main.a")
			for name, source := range map[string]string{
				"main.a":     "import * as performance from './provider.a';\n" + probe.source,
				"provider.a": probe.provider,
			} {
				if writeError := os.WriteFile(
					filepath.Join(directory, name),
					[]byte(source),
					0o644,
				); writeError != nil {
					test.Fatal(writeError)
				}
			}
			program, loadError := load.Load([]string{entry})
			if probe.reason == "" {
				if loadError == nil {
					test.Fatal("checker allowed writing an ESM export")
				}
				return
			}
			if loadError != nil {
				test.Fatal(loadError)
			}
			_, lowerError := Lower(context.Background(), program)
			var notYet *NotYet
			if !errors.As(lowerError, &notYet) || !strings.Contains(lowerError.Error(), probe.reason) {
				test.Fatalf("got %v, want NotYet %s", lowerError, probe.reason)
			}
		})
	}
}

func TestModuleNamespaceInitializedReadProof(test *testing.T) {
	test.Parallel()
	path := filepath.Join("..", "oracle", "testdata", "module_namespace_reads", "direct_initialized.a")
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	requireLoweredOutput(test, lowered)
	if strings.Contains(native.C(lowered), "ReferenceError: Cannot access 'value'") {
		test.Fatal("provider initialized before consumer, but the qualified read retained a readiness check")
	}
}
