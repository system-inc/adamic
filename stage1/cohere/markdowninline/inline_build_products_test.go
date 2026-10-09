package markdowninline

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"path/filepath"
	"sync"
	"testing"
)

var inlineCompilerOnce sync.Once
var inlineCompilerRecipe buildcache.Inputs

func inlineCompilerInputs(t *testing.T) buildcache.Inputs {
	t.Helper()
	inlineCompilerOnce.Do(func() { inlineCompilerRecipe = collectInlineCompilerInputs(t) })
	return inlineCompilerRecipe
}

var inlineMutations = []inlineMutation{
	{"one-case disagreement", "main.ts", "console.log(encode(formatLeaf(line.slice(0, 1), decode(line.slice(1)))));", "console.log(line === 'w😀😀😀😀' ? 'planted disagreement' : encode(formatLeaf(line.slice(0, 1), decode(line.slice(1)))));"},
	{"escaped delimiter parity", "inline.ts", "(found.preceding - position) % 2 === 1", "(found.preceding - position) % 2 === 0"},
	{"table pipe escaping", "inline.ts", "if(table)", "if(!table)"},
	{"minimum absent fence", "inline.ts", "while(runs.includes(count))", "while(runs.includes(count) && count < 1)"},
}

func inlineProductName(index int) string {
	if index < 0 {
		return "markdowninline"
	}
	if index == 0 {
		return "markdowninline planted"
	}
	return "markdowninline mutant " + inlineMutations[index].name
}
func inlineLoweredProduct(t *testing.T, index int) string {
	t.Helper()
	name := inlineProductName(index) + " lowered"
	build := buildInlineLowered
	if index < 0 {
		name = "markdowninline lowered program"
	} else {
		build = inlineMutations[index].buildLowered
	}
	in := inlineProgramInputs(inlineCompilerInputs(t), name)
	in.Files = append(append([]string{}, in.Files...), "stage1/cohere/markdowninline/inline_build_products_test.go")
	return inlineBuild(t, in, build)
}
func inlineNativeProduct(t *testing.T, index int, sanitize bool) string {
	t.Helper()
	source := inlineLoweredProduct(t, index)
	suffix := " native"
	build := inlineSource(source).buildNative
	if sanitize {
		suffix = " sanitized"
		build = inlineSource(source).buildSanitized
	}
	return inlineBuild(t, inlineNativeInputs(t, inlineProductName(index)+suffix, source, sanitize), build)
}
func inlineNodeProduct(t *testing.T) string {
	t.Helper()
	source, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return inlineBuild(t, inlineNodeInputs(t, "markdowninline Node oracle", source), inlineSource(source).buildNode)
}
func inlineMutationNodeProduct(t *testing.T, index int) string {
	t.Helper()
	source := inlineLoweredProduct(t, index)
	return inlineBuild(t, inlineNodeInputs(t, inlineProductName(index)+" Node", source), inlineSource(source).buildNode)
}

func TestProduct_MarkdownInlineGoOverlay(t *testing.T) {
	t.Parallel()
	inlineOverlayBuild(t)
}

func TestProduct_MarkdownInlineNode(t *testing.T) {
	t.Parallel()
	inlineNodeProduct(t)
}

func TestProduct_MarkdownInlineOriginalLowered(t *testing.T) {
	t.Parallel()
	inlineLoweredProduct(t, -1)
}

func TestProduct_MarkdownInlineOriginalNative(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, -1, false)
}

func TestProduct_MarkdownInlineOriginalSanitized(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, -1, true)
}

func TestProduct_MarkdownInlinePlantedLowered(t *testing.T) {
	t.Parallel()
	inlineLoweredProduct(t, 0)
}

func TestProduct_MarkdownInlinePlantedNative(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, 0, false)
}

func TestProduct_MarkdownInlinePlantedSanitized(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, 0, true)
}

func TestProduct_MarkdownInlineParityLowered(t *testing.T) {
	t.Parallel()
	inlineLoweredProduct(t, 1)
}

func TestProduct_MarkdownInlineParityNative(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, 1, false)
}

func TestProduct_MarkdownInlineParitySanitized(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, 1, true)
}

func TestProduct_MarkdownInlineParityNode(t *testing.T) {
	t.Parallel()
	inlineMutationNodeProduct(t, 1)
}

func TestProduct_MarkdownInlinePipeLowered(t *testing.T) {
	t.Parallel()
	inlineLoweredProduct(t, 2)
}

func TestProduct_MarkdownInlinePipeNative(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, 2, false)
}

func TestProduct_MarkdownInlinePipeSanitized(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, 2, true)
}

func TestProduct_MarkdownInlinePipeNode(t *testing.T) {
	t.Parallel()
	inlineMutationNodeProduct(t, 2)
}

func TestProduct_MarkdownInlineFenceLowered(t *testing.T) {
	t.Parallel()
	inlineLoweredProduct(t, 3)
}

func TestProduct_MarkdownInlineFenceNative(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, 3, false)
}

func TestProduct_MarkdownInlineFenceSanitized(t *testing.T) {
	t.Parallel()
	inlineNativeProduct(t, 3, true)
}

func TestProduct_MarkdownInlineFenceNode(t *testing.T) {
	t.Parallel()
	inlineMutationNodeProduct(t, 3)
}
