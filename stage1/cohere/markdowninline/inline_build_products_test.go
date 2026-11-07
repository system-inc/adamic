package markdowninline

import (
	"path/filepath"
	"testing"
)

var inlineMutations = []inlineMutation{
	{
		"one-case disagreement",
		"main.ts",
		"console.log(encode(formatLeaf(line.slice(0, 1), decode(line.slice(1)))));",
		"console.log(line === 'w😀😀😀😀' ? 'planted disagreement' : encode(formatLeaf(line.slice(0, 1), decode(line.slice(1)))));",
	},
	{
		"escaped delimiter parity",
		"inline.ts",
		"(found.preceding - position) % 2 === 1",
		"(found.preceding - position) % 2 === 0",
	},
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
func inlineLoweredProduct(test *testing.T, index int) string {
	test.Helper()
	build := buildInlineLowered
	if index >= 0 {
		build = inlineMutations[index].buildLowered
	}
	return inlineBuild(test, inlineLoweredInputs(index), build)
}
func inlineNativeProduct(test *testing.T, index int, sanitize bool) string {
	test.Helper()
	source := inlineLoweredProduct(test, index)
	suffix := " native"
	build := inlineSource(source).buildNative
	if sanitize {
		suffix = " sanitized"
		build = inlineSource(source).buildSanitized
	}
	return inlineBuild(test, inlineNativeInputs(test, inlineProductName(index)+suffix, source, sanitize), build)
}
func inlineNodeProduct(test *testing.T) string {
	test.Helper()
	source, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return inlineBuild(
		test,
		inlineNodeInputs(test, "markdowninline Node oracle", source),
		inlineSource(source).buildNode,
	)
}
func inlineMutationNodeProduct(test *testing.T, index int) string {
	test.Helper()
	source := inlineLoweredProduct(test, index)
	return inlineBuild(
		test,
		inlineNodeInputs(test, inlineProductName(index)+" Node", source),
		inlineSource(source).buildNode,
	)
}

func TestProduct_MarkdownInlineFindsGoOverlay(test *testing.T) {
	test.Parallel()
	inlineOverlayBuild(test)
}

func TestProduct_MarkdownInlineNode(test *testing.T) {
	test.Parallel()
	inlineNodeProduct(test)
}

func TestProduct_MarkdownInlineOriginalLowered(test *testing.T) {
	test.Parallel()
	inlineLoweredProduct(test, -1)
}

func TestProduct_MarkdownInlineOriginalNative(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, -1, false)
}

func TestProduct_MarkdownInlineOriginalSanitized(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, -1, true)
}

func TestProduct_MarkdownInlinePlantedLowered(test *testing.T) {
	test.Parallel()
	inlineLoweredProduct(test, 0)
}

func TestProduct_MarkdownInlinePlantedNative(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, 0, false)
}

func TestProduct_MarkdownInlinePlantedSanitized(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, 0, true)
}

func TestProduct_MarkdownInlineParityLowered(test *testing.T) {
	test.Parallel()
	inlineLoweredProduct(test, 1)
}

func TestProduct_MarkdownInlineParityNative(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, 1, false)
}

func TestProduct_MarkdownInlineParitySanitized(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, 1, true)
}

func TestProduct_MarkdownInlineParityNode(test *testing.T) {
	test.Parallel()
	inlineMutationNodeProduct(test, 1)
}

func TestProduct_MarkdownInlinePipeLowered(test *testing.T) {
	test.Parallel()
	inlineLoweredProduct(test, 2)
}

func TestProduct_MarkdownInlinePipeNative(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, 2, false)
}

func TestProduct_MarkdownInlinePipeSanitized(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, 2, true)
}

func TestProduct_MarkdownInlinePipeNode(test *testing.T) {
	test.Parallel()
	inlineMutationNodeProduct(test, 2)
}

func TestProduct_MarkdownInlineFenceLowered(test *testing.T) {
	test.Parallel()
	inlineLoweredProduct(test, 3)
}

func TestProduct_MarkdownInlineFenceNative(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, 3, false)
}

func TestProduct_MarkdownInlineFenceSanitized(test *testing.T) {
	test.Parallel()
	inlineNativeProduct(test, 3, true)
}

func TestProduct_MarkdownInlineFenceNode(test *testing.T) {
	test.Parallel()
	inlineMutationNodeProduct(test, 3)
}
