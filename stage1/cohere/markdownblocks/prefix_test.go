package markdownblocks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

// Not parallel: shared markdownMemory configuration; existing helper controls parallel execution.
func TestMarkdownParserPrefixes(test *testing.T) {
	parallelMarkdown(test)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, files := auditCorpus(test, root)
	// This file sweep is small beside the mandatory generated and fixed checks.
	selection := selectMarkdownFiles(test, root, inputs, files, 1)
	if selection.Sample {
		test.Log(selection.Log(test.Name()))
	}
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
	for _, text := range []string{
		"",
		"\ufeff",
		">",
		"> ",
		">\t",
		" > a",
		"   > a",
		"    > a",
		"\t> a",
		"\t > a",
		">> a",
		"> > a",
		"\ufeff> a",
		"\r\na",
		"\ra",
		"a\r",
		"a\r\nb",
		"😀\t> a",
	} {
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/prefix/short/%d", len(inputs)), Text: text})
	}
	units := []string{"a", " ", "\t", "\r", "\n", "\x00", "\ufeff", "😀", ">"}
	for _, a := range units {
		for _, b := range units {
			for _, c := range units {
				text := a + b + c
				inputs = append(inputs, auditInput{Name: "generated/prefix/" + text, Text: text})
			}
		}
	}
	for column := 0; column <= 12; column++ {
		for _, prefix := range []string{" ", "\t", "  \t", "\t ", "\ufeff"} {
			for _, suffix := range []string{"", "a", " ", "\t", "\r\n", "😀"} {
				text := strings.Repeat(" ", column) + prefix + ">" + suffix
				inputs = append(inputs, auditInput{Name: "generated/prefix/indent/" + text, Text: text})
			}
		}
	}
	// All scalar ranges in one source check UTF-16 chunks and tab columns together.
	var unicode strings.Builder
	for code := rune(0); code <= 0xffff; code++ {
		if code >= 0xd800 && code <= 0xdfff {
			continue
		}
		unicode.WriteRune(code)
	}
	inputs = append(inputs, auditInput{Name: "generated/prefix/unicode", Text: unicode.String() + "😀\t\U0010ffff"})
	dir := test.TempDir()
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	var batch strings.Builder
	for _, input := range inputs {
		batch.WriteString("x" + encode.Replace(input.Text) + "\n")
	}
	cases := filepath.Join(dir, "cases.txt")
	write(test, cases, []byte(batch.String()))
	goBinary := prefixGoOracle(test)
	want := execute(test, nil, goBinary, cases)
	clean(test, "actual Go parser primitives", want)
	main, absolutePathError := filepath.Abs("testdata/prefix_probe.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program := lowered(test, main)
	binary := filepath.Join(probeNative(test, prefixProbe, native.Options{Sanitize: true}), "port")
	answer := sanitizedRun(test, binary, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native parser prefixes", answer},
		{"source Node parser prefixes", onNode(test, main, cases)},
		{"backend parser prefixes", onJavaScriptBackend(test, program, cases)},
	} {
		clean(test, side.name, side.result)
		if !bytes.Equal(side.result.stdout, want.stdout) {
			offset := firstDifference(string(side.result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			test.Fatalf("%s differs at byte %d in %q", side.name, offset, inputs[index].Name)
		}
	}
	if report := leaks(test, program, binary, cases); report != "" {
		test.Fatal(report)
	}
	for _, mutation := range []struct{ name, file, from, to string }{
		{"CRLF chunk", "preprocess.ts", "code: -3", "code: -4"},
		{"tab virtual spaces", "preprocess.ts", "column < next", "column <= next"},
		{"quote continuation limit", "quotePrefix.ts", "index >= 3", "index >= 4"},
	} {
		test.Run(mutation.name, func(test *testing.T) {
			scratch := test.TempDir()
			if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			for _, name := range []string{"preprocess.ts", "quotePrefix.ts", "codec.ts"} {
				content, readError := os.ReadFile(name)
				if readError != nil {
					test.Fatal(readError)
				}
				if name == mutation.file {
					if strings.Count(string(content), mutation.from) != 1 {
						test.Fatal("mutant site count")
					}
					content = []byte(strings.Replace(string(content), mutation.from, mutation.to, 1))
				}
				write(test, filepath.Join(scratch, name), content)
			}
			content, readError := os.ReadFile(main)
			if readError != nil {
				test.Fatal(readError)
			}
			mutantMain := filepath.Join(scratch, "testdata/prefix_probe.ts")
			write(test, mutantMain, content)
			result := onNode(test, mutantMain, cases)
			clean(test, "source Node output-only prefix mutant", result)
			if bytes.Equal(result.stdout, want.stdout) {
				test.Fatal("mutant survived")
			}
			offset := firstDifference(string(result.stdout), string(want.stdout))
			index := bytes.Count(want.stdout[:offset], []byte("\n"))
			test.Logf("output-only source Node mutant caught in %q at byte %d", inputs[index].Name, offset)
		})
	}
	fast := filepath.Join(probeNative(test, prefixProbe, native.Options{}), "port")
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"actual Go preprocessing/constructs", goBinary, []string{cases}},
		{"native preprocessing/prefix recognition", fast, []string{cases}},
		{
			"source Node preprocessing/prefix recognition",
			"node",
			[]string{"--disable-warning=ExperimentalWarning", runner, main, cases},
		},
	} {
		var elapsed time.Duration
		for round := 0; round < 3; round++ {
			start := time.Now()
			result := execute(test, nil, side.command, side.args...)
			elapsed += time.Since(start)
			clean(test, side.name, result)
			equal(test, side.name, result.stdout, want.stdout)
		}
		test.Logf(
			"parser primitive throughput %s %.1f texts/s, three runs %.6fs; source preprocessing/chunks/prefix recognition and startup/protocol/output included; Go also drains tokenizer and emits construct events, native does not",
			side.name,
			float64(len(inputs)*3)/elapsed.Seconds(),
			elapsed.Seconds(),
		)
	}
	test.Logf(
		"%d physical files, %d generated, %d sources; exact chunk boundaries/codes and %d quote prefix contexts match actual Go/source/native/backend; ASan/UBSan/leaks",
		files,
		len(inputs)-files,
		len(inputs),
		len(inputs)*8,
	)
}

// prefixGoOracle is the Go side, cohere with the driver and bridge its overlay lays over it, a product built ahead
// (#5qykzj5).
func prefixGoOracle(test *testing.T) string { test.Helper(); return oracles.PrefixGoOracle(test) }
