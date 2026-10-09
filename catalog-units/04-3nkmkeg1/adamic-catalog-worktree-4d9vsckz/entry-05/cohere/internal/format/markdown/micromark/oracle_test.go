package micromark

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
)

// The event oracle: the fork's own micromark, run by Node with the extensions parse-markdown.js passes,
// asked for the event stream of each input. The Go port must produce the same stream, every token's type,
// direction and both positions, before mdast-util-from-markdown is ever involved, so a construct is
// measured at the layer it was written in.
//
// Node runs the production build (each package's default export), which is what the bundle carries. The
// one thing Node cannot reproduce is the bundle's rewritten Unicode punctuation class, which is why the
// Go side generates that table from the bundle instead; inputs here should not depend on a code point
// whose classification changed between the bundle's Unicode and this Node's.

// forkRoot is the Prettier fork whose node_modules the oracle imports from (internal/corpus): a skip naming
// COHERE_PRETTIER_FORK when it is unset, and a failure when it is set but lacks node_modules/micromark.
func forkRoot(t *testing.T) string {
	t.Helper()
	corpus.Resolve(t, "prettier-fork:node_modules/micromark")
	return corpus.PrettierFork.Root(t)
}

const oracleScript = `
import { parse, postprocess, preprocess } from "FORK/node_modules/micromark/index.js";
import { gfm } from "FORK/node_modules/micromark-extension-gfm/index.js";
import { math } from "FORK/node_modules/micromark-extension-math/index.js";
import { syntax as wikiLinkSyntax } from "FORK/node_modules/@braindb/micromark-extension-wiki-link/dist/index.js";
import { liquidSyntax } from "FORK/src/language-markdown/parse/micromark/micromark-extension-liquid.js";
import { overrideHtmlTextSyntax } from "FORK/src/language-markdown/parse/micromark/micromark-extension-html-text.js";
import { readFileSync } from "node:fs";

const options = {
  extensions: [
    gfm(),
    math(),
    wikiLinkSyntax({ aliasDivider: { charCodeAt: () => Number.NaN } }),
    liquidSyntax(),
    overrideHtmlTextSyntax(),
  ],
};

const inputs = JSON.parse(readFileSync(0, "utf8"));
const outputs = inputs.map((input) => {
  try {
    const events = postprocess(parse(options).document().write(preprocess()(input, undefined, true)));
    return events.map(([kind, token]) => describe(kind, token));
  } catch (error) {
    return ["error: " + error.message];
  }
});
process.stdout.write(JSON.stringify(outputs));

function describe(kind, token) {
  const { start, end } = token;
  return kind + " " + token.type + " " + start.line + ":" + start.column + ":" + start.offset + "-" + end.line + ":" + end.column + ":" + end.offset;
}
`

// upstreamEvents runs the oracle over inputs, one Node process for all of them.
func upstreamEvents(t *testing.T, inputs []string) [][]string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed; the event oracle needs it")
	}
	root := forkRoot(t)

	directory := t.TempDir()
	script := filepath.Join(directory, "oracle.mjs")
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(oracleScript, "FORK", root)), 0o644); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", script)
	command.Stdin = strings.NewReader(string(encoded))
	output, err := command.Output()
	if err != nil {
		if exitError, isExit := err.(*exec.ExitError); isExit {
			t.Fatalf("the event oracle failed: %v\n%s", err, exitError.Stderr)
		}
		t.Fatal(err)
	}
	var outputs [][]string
	if err := json.Unmarshal(output, &outputs); err != nil {
		t.Fatal(err)
	}
	return outputs
}

// portEvents describes the Go port's events the way the oracle describes upstream's.
func portEvents(input string) []string {
	return portEventsFrom(input, nil)
}

// portEventsFrom is portEvents with the parse taking its tokens and tokenizers from memory.
func portEventsFrom(input string, memory *Memory) (described []string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			described = []string{fmt.Sprintf("panic: %v", recovered)}
		}
	}()
	events := Parse(SourceUnits(input), MarkdownConstructs(), memory)
	for _, event := range events {
		kind := "exit"
		if event.Enter {
			kind = "enter"
		}
		start, end := event.Token.Start, event.Token.End
		described = append(described, fmt.Sprintf("%s %s %d:%d:%d-%d:%d:%d", kind, event.Token.Type,
			start.Line, start.Column, start.Offset, end.Line, end.Column, end.Offset))
	}
	return described
}

// compareEvents reports the first divergence per input.
func compareEvents(t *testing.T, inputs []string) (failures int) {
	t.Helper()
	expected := upstreamEvents(t, inputs)
	for index, input := range inputs {
		actual := portEvents(input)
		if difference := firstEventDifference(expected[index], actual); difference != "" {
			failures++
			t.Errorf("input %q: %s", truncateInput(input), difference)
		}
	}
	return failures
}

func firstEventDifference(expected []string, actual []string) string {
	for index := 0; index < len(expected) || index < len(actual); index++ {
		var want, got string
		if index < len(expected) {
			want = expected[index]
		}
		if index < len(actual) {
			got = actual[index]
		}
		if want != got {
			return fmt.Sprintf("event %d: want %q, got %q", index, want, got)
		}
	}
	return ""
}

func truncateInput(input string) string {
	if len(input) > 80 {
		return input[:80] + "…"
	}
	return input
}

// TestEventsMatchUpstream runs every fixture through both sides.
func TestEventsMatchUpstream(t *testing.T) {
	t.Parallel()
	compareEvents(t, eventFixtures)
}

// TestEventOracleCanFail proves the comparison sees a difference: a fixture compared against the events
// of a different input must fail.
func TestEventOracleCanFail(t *testing.T) {
	t.Parallel()
	expected := upstreamEvents(t, []string{"*a*"})
	if firstEventDifference(expected[0], portEvents("**a**")) == "" {
		t.Fatal("the comparison found no difference between the events of *a* and **a**")
	}
}

// TestCorpusEventsMatchUpstream compares every markdown file under COHERE_MARKDOWN_CORPUS, a measuring run.
func TestCorpusEventsMatchUpstream(t *testing.T) {
	t.Parallel()
	root := os.Getenv("COHERE_MARKDOWN_CORPUS")
	if root == "" {
		t.Skip("set COHERE_MARKDOWN_CORPUS to a directory to compare every .md file under it")
	}
	var paths []string
	var inputs []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == ".git") {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(path), ".md") {
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			paths = append(paths, path)
			// Prettier normalizes line endings and strips a byte order mark before parsing.
			text := strings.ReplaceAll(strings.ReplaceAll(string(source), "\r\n", "\n"), "\r", "\n")
			inputs = append(inputs, strings.TrimPrefix(text, "\uFEFF"))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Node reads its input and writes its events as one string each, capped at 512MB, and the events of a
	// tree run to many times its text, so the oracle is asked in batches of about 4MB of input.
	const batchBytes = 4 << 20
	failures := 0
	for start := 0; start < len(inputs); {
		end, size := start, 0
		for end < len(inputs) && (end == start || size+len(inputs[end]) <= batchBytes) {
			size += len(inputs[end])
			end++
		}
		expected := upstreamEvents(t, inputs[start:end])
		for index := start; index < end; index++ {
			if difference := firstEventDifference(expected[index-start], portEvents(inputs[index])); difference != "" {
				failures++
				if failures <= 40 {
					t.Errorf("%s: %s", paths[index], difference)
				}
			}
		}
		start = end
	}
	t.Logf("%d of %d files match upstream's events", len(inputs)-failures, len(inputs))
}
