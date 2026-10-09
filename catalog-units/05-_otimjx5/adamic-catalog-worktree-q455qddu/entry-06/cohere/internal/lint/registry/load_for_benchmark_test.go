package registry

import (
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/lint/configuration"
)

// BenchmarkLoadForTheLiveConfig loads a real project's config against the real registry, which is what
// every run pays before its first phase. It was 590ms on ahra when every pair of keys rescanned the
// registry, and is about 7ms with each key asked once (#kgv1pry).
//
// It lives here rather than beside LoadFor because it needs the whole rule set, and every test binary that
// links the rule set relinks on every rule edit: in configuration it cost each rule edit a link and a
// second compile of the registry (#6gct10n).
//
// The config is ahra's, which is private, so the benchmark skips where the ahra corpus is not set (#sycrdr6).
func BenchmarkLoadForTheLiveConfig(b *testing.B) {
	path := corpus.Ahra.Path(b, "CohereSettings.json")
	names := Names()
	for b.Loop() {
		if _, err := configuration.LoadFor(path, names); err != nil {
			b.Fatal(err)
		}
	}
}
