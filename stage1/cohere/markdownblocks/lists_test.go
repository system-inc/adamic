package markdownblocks

import (
	"testing"

	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

// Not parallel: listLayoutShared setup cache and shared markdownMemory configuration; existing helper controls parallel execution. shared layout fixtures and memory-budget admission are handled by the setup helper. shared listLayoutShared and listLayoutBuckets; existing helper controls parallel execution.
// Not parallel: list-layout fixture and markdown memory configuration.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownListLayout(t *testing.T) {
	t.Parallel()
	buildListLayoutSetup(t)
}

// Not parallel: quoteLayoutShared is initialized for the parallel quote layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Not parallel: quote lowered and Go build products initialized before parallel shards.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownQuoteLayout(t *testing.T) {
	quoteLayoutSetup(t)
}

// Not parallel: tableLayoutSharedProducts and tableLayoutInputs are initialized for the parallel table layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Legacy entry point reports shared preparation completed before m.Run.
func TestMarkdownTableLayout(t *testing.T) {
	t.Parallel()
	testMarkdownTableLayout(t)
}

type layoutFixture struct {
	selection                                                                  gatesample.Selection
	mutantNativeCases                                                          string
	mutantInputs                                                               []oracles.AuditInput
	mutantWant                                                                 []byte
	root, directory, nativeCases, canonicalCases, main, fork, goLayout, script string
	inputs                                                                     []oracles.AuditInput
	files                                                                      int
	want                                                                       []byte
	program                                                                    *ir.Program
}

// Not parallel: shared leaf layout shard setup and markdownMemory configuration; existing helper controls parallel execution. shared layout fixtures and memory-budget admission are handled by the setup helper. shared leafCompositionPrepared and leafCompositionBuilds; existing helper controls parallel execution.
// Not parallel: markdown memory configuration checked by parallelMarkdown before parallel admission.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownLeafComposition(t *testing.T) {
	t.Parallel()
	configureMarkdownMemory(t)
	testMarkdownLeafShards(t)
}

// Not parallel: structureLayoutComplete is initialized for the parallel structure layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Not parallel: structure-layout fixture and markdown memory configuration.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownStructureLayout_Setup(t *testing.T) {
	configureMarkdownMemory(t)
	structureLayoutShared(t)
}
