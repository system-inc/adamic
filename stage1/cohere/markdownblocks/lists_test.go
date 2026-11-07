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
func TestMarkdownListLayout(test *testing.T) {
	test.Parallel()
	buildListLayoutSetup(test)
}

// Not parallel: quoteLayoutShared is initialized for the parallel quote layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Not parallel: quote lowered and Go build products initialized before parallel shards.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownQuoteLayout(test *testing.T) {
	quoteLayoutSetup(test)
}

// Not parallel: tableLayoutSharedProducts and tableLayoutInputs are initialized for the parallel table layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Legacy entry point reports shared preparation completed before m.Run.
func TestMarkdownTableLayout(test *testing.T) {
	test.Parallel()
	testMarkdownTableLayout(test)
}

type layoutFixture struct {
	selection                                                                  gatesample.Selection
	mutantNativeCases                                                          string
	mutantInputs                                                               []auditInput
	mutantWant                                                                 []byte
	root, directory, nativeCases, canonicalCases, main, fork, goLayout, script string
	inputs                                                                     []auditInput
	files                                                                      int
	want                                                                       []byte
	program                                                                    *ir.Program
}

// Not parallel: shared leaf layout shard setup and markdownMemory configuration; existing helper controls parallel execution. shared layout fixtures and memory-budget admission are handled by the setup helper. shared leafCompositionPrepared and leafCompositionBuilds; existing helper controls parallel execution.
// Not parallel: markdown memory configuration checked by parallelMarkdown before parallel admission.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownLeafComposition(test *testing.T) {
	test.Parallel()
	configureMarkdownMemory(test)
	testMarkdownLeafShards(test)
}

// Not parallel: structureLayoutComplete is initialized for the parallel structure layout shards. shared layout fixtures and memory-budget admission are handled by the setup helper.
// Not parallel: structure-layout fixture and markdown memory configuration.
// Not parallel: shared layout fixtures and memory-budget admission are handled by the setup helper.
func TestMarkdownStructureLayout_Setup(test *testing.T) {
	configureMarkdownMemory(test)
	structureLayoutShared(test)
}

func blockCorpus(test *testing.T, root, slice string) ([]auditInput, int) {
	return oracles.BlockCorpus(test, root, slice)
}
