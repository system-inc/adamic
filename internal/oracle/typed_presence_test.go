package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Family 65's typed half (#2hzryxh, Adamic's road (b) on #bbssdcy): a spread copies presence per slot, so `in`
// with a fixed key on a receiver the checker proves one object type answers from what the copy holds, then from
// the prototype's names, as Node does. Each of Language's three conditions has its program: a class getter seen
// through a { readonly y: number } view isn't copied (65-getter-view.ts, false), prototype names still answer
// (65-prototype.ts), and present-but-undefined is present while an absent spread adds nothing (65-undefined.ts,
// 65-typed-presence.ts's first line). The rest are shapes tsc's own code doesn't write: a copy made in another
// module, a copy of a copy, an in in a while condition, a source chosen by a ternary, an optional source given
// {}, a write that publishes a field the absent spread left out, printed by JSON too, and a receiver typed by a
// type parameter, as builder.ts's isFullChain is. Before this change each stopped with the family's words, or
// agreed already where no spread may be absent.
func TestTypedPresenceAgrees(t *testing.T) {
	t.Parallel()
	for file, node := range map[string]string{
		"65-typed-presence.ts":  "false\ntrue\n",
		"65-getter-view.ts":     "false\n",
		"65-prototype.ts":       "true\ntrue\nfalse\n",
		"65-undefined.ts":       "true\n",
		"65-across-modules.ts":  "false false undefined\ntrue true 3\n",
		"65-copy-of-copy.ts":    "false\ntrue\nundefined 4\n",
		"65-while.ts":           "0\n10\n20\n3\n",
		"65-ternary-source.ts":  "true\nfalse\ntrue\n",
		"65-optional-source.ts": "false\ntrue\nfalse\n",
		"65-later-write.ts":     "false\ntrue\n8\n{} {\"code\":8}\n",
		"65-generic.ts":         "true\nfalse\nfalse\ntrue\n",
	} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_presence", file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// What the typed half leaves stopped, each where and in the words it stopped at main. Both forms of the
// overwrite { ...source, k: undefined } stay #xy74m5x's stop (Adamic's 19:32:48Z ruling): the number form, and
// the optional form pinned from spread_overwrite's own file, beside a copied {} reaching an optional source.
// They turn into running pins when a spread writes its own layout (#7nhzq7h; typed_presence/README.md). The
// unknown half is held, so an in on a receiver narrowed to object keeps the family's stop though every copy now
// carries its bits. Lowering stops before either backend, so one check covers both. If a stop goes missing, the
// program is run natively under the sanitizers and its run is part of the failure.
func TestTypedPresenceStops(t *testing.T) {
	t.Parallel()
	const held = "an overwrite of k with value after a spread whose source holds k as number, so the write would take a scalar slot for a reference or a reference for a scalar"
	const optional = "an overwrite of k after a spread whose source may not have it (its type makes k optional), so the copied object may have no slot to write"
	const family = "stage 0 can't lower in when a possibly absent spread has no property presence descriptors yet"
	for _, pin := range []struct{ file, node, at, words string }{
		{"typed_presence/65-leading-overwrite.ts", "true\nundefined\ntrue\nundefined\n", "k: undefined", held},
		{"spread_overwrite/optional_number.ts", "true\nundefined\ntrue\nundefined\n", "k: undefined", optional},
		{"typed_presence/copied-empty-object.ts", "undefined\nundefined\n", "k: undefined", optional},
		{"typed_presence/65-unknown-presence.ts", "false\ntrue\n", `"code" in value`, family},
	} {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", pin.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(pin.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			want, err := spreadOverwriteLocation(path, pin.at)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err == nil {
				native, _ := natively(t, program)
				t.Fatalf("lowered; want a stop at %s saying %q. Sanitized native: exit %d, stdout %q, stderr %q", want, pin.words, native.exitCode, native.stdout, native.stderr)
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want a NotYet, got %v", err)
			}
			if notYet.Where != want || !strings.Contains(err.Error(), pin.words) {
				t.Fatalf("got %q at %s; want %q at %s", err.Error(), notYet.Where, pin.words, want)
			}
		})
	}
}

// Adamic's two 05:57Z refusals, each with its words: the overwrite written as a literal before a spread is
// refused by stock tsc (TS2783 at 1:16), and through a view by adamic/single-spread at 2:26, so neither reaches
// the bits.
func TestTypedPresenceOverwriteRefusals(t *testing.T) {
	t.Parallel()
	t.Run("65-overwrite.ts", func(t *testing.T) {
		t.Parallel()
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_presence/65-overwrite.ts"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = load.Load([]string{path})
		var checked *load.CheckError
		const diagnostic = ":1:16: error TS2783: 'k' is specified more than once, so this usage will be overwritten."
		if !errors.As(err, &checked) || !strings.Contains(err.Error(), filepath.ToSlash(path)+diagnostic) {
			t.Fatalf("want tsc's %q, got %v", diagnostic, err)
		}
	})
	t.Run("65-overwrite-view.ts", func(t *testing.T) {
		t.Parallel()
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/typed_presence/65-overwrite-view.ts"))
		if err != nil {
			t.Fatal(err)
		}
		want, err := spreadOverwriteLocation(path, "...source")
		if err != nil {
			t.Fatal(err)
		}
		_, err = lowered(t, path)
		var refused *lower.Refused
		const words = "Adamic 0.1 refuses a spread after the first field; spread once, first: { ...source, field: value } (adamic/single-spread)"
		if !errors.As(err, &refused) || refused.Where != want || !strings.Contains(err.Error(), words) {
			t.Fatalf("want %q at %s, got %v", words, want, err)
		}
	})
}
