package prettier

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

/*
 * Why a Go linter depends on a JavaScript interpreter.
 *
 * cohere replaces Prettier, and replacing Prettier means matching it byte for byte on a tree that is
 * already Prettier-formatted. Anything short of that produces a diff nobody asked for and buries
 * real changes in noise. The alternative was measured rather than assumed: typescript-go ships a
 * formatter, and its FormatCodeSettings has 21 fields and no printWidth, so it has no line-breaking
 * engine at all. It leaves a 130-character line at 130 characters. 21.8% of tracked files diverge
 * from our Prettier after every settings-reachable fix.
 *
 * So the only thing that reproduces our formatter is our formatter. goja runs it in-process, which
 * costs about 9.6 MB in the stripped binary and buys byte-identical output on every file type
 * Prettier handles, including the four our own fork customizes.
 */

// ForkPathVariable selects loading from a fork checkout instead of the embedded bundles.
//
// The meaning changed when the bundles were vendored, and the old meaning is worth stating because
// anything set in a CI config still carries it. It used to answer "where is the fork", back when
// disk was the only mode and the alternative was a hardcoded home directory. It now answers "load
// from disk rather than from this binary", and the default is no longer a path at all -- it is the
// embedded copy, which needs no fork present.
//
// Being set is the entire mode switch, evaluated in one place (Bundles), so there is no second
// condition to keep in agreement with this one. That matters here more than it usually would: the
// fork was previously reachable by two independent paths, this engine's own constant and the release
// guard's environment lookup, and neither consulted the other. Overriding one left the other pinned,
// so a CI runner that set this variable got a release guard reporting green over an engine still
// reading a home directory on one laptop. Measured rather than reasoned: with the variable pointed
// at a nonexistent path, `New` returned a working engine.
//
// So the engine that loads the bundles owns the name, and every other consumer resolves through it.
// `prettier` imports nothing internal, which is what makes it the safe home; `release` already
// depends on it, so the reference points down rather than sideways. Anyone adding a third consumer
// of the fork should find one place here rather than guess which of two to copy.
const ForkPathVariable = "COHERE_PRETTIER_FORK"

// BundleFiles are the Prettier bundles the engine evaluates, in dependency order.
//
// standalone.js first because the plugins register themselves against it. Everything else is one
// language, and the set is what parserForExtension can ask for.
//
// Exported so the vendoring build step can copy exactly what the engine loads rather than keeping a
// second list that drifts from this one. A bundle vendored but not loaded is dead weight; a bundle
// loaded but not vendored is a build that fails at runtime on a machine without the fork.
var BundleFiles = []string{
	"standalone.js",
	"plugins/estree.js",
	"plugins/typescript.js",
	"plugins/babel.js",
	"plugins/postcss.js",
	"plugins/markdown.js",
	"plugins/graphql.js",
	"plugins/yaml.js",
}

// loadBundles returns each bundle's source, in evaluation order.
//
// The ordering is this function's job and Bundles' map cannot carry it: standalone.js has to be
// evaluated before the plugins that register against it, and a map has no order. So Bundles decides
// where the bytes come from and this decides what order they are fed to the interpreter, which are
// genuinely different questions.
//
// A missing bundle is a hard error rather than a skipped language, and Bundles has already enforced
// that -- an engine that quietly loaded seven of eight would format the eighth's files by falling
// through to no parser at all, and the failure would look like a file type nobody formats rather
// than a broken build.
func loadBundles() ([]namedSource, error) {
	bundles, err := Bundles()
	if err != nil {
		return nil, err
	}
	sources := make([]namedSource, 0, len(BundleFiles))
	for _, name := range BundleFiles {
		sources = append(sources, namedSource{name: name, text: string(bundles.Files[name])})
	}
	return sources, nil
}

// namedSource is one bundle and where it came from, so an evaluation failure can name the file.
type namedSource struct {
	name string
	text string
}

// drainPromise resolves a value that may be a promise, and this is the least obvious part of the
// engine.
//
// prettier.format returns a promise. goja implements promises but has no event loop, because an
// event loop is a host concern rather than a language one -- the interpreter has a job queue and
// nothing that pumps it. So a promise handed back to Go sits Pending forever, and a caller that
// simply stringifies the result gets the literal text "[object Promise]" written into the file. That
// is not hypothetical; it is what the first working version of this produced.
//
// The pump is goja's own: every RunString drains pending jobs as part of finishing, so evaluating a
// trivial expression advances the queue. Looping on that is what settles the promise.
//
// The iteration cap is a deadlock guard rather than a timeout. Prettier's formatting is synchronous
// under the hood, so a promise that has not settled after this many pumps is not slow, it is stuck,
// and returning an error beats hanging a build.
func drainPromise(vm *goja.Runtime, value goja.Value) (string, error) {
	promise, isPromise := value.Export().(*goja.Promise)
	if !isPromise {
		return value.String(), nil
	}

	const maxPumps = 100000
	for pump := 0; pump < maxPumps && promise.State() == goja.PromiseStatePending; pump++ {
		if _, err := vm.RunString("0"); err != nil {
			return "", fmt.Errorf("pumping the job queue: %w", err)
		}
	}

	switch promise.State() {
	case goja.PromiseStateFulfilled:
		return promise.Result().String(), nil
	case goja.PromiseStateRejected:
		// Go whitespace: a rejection's message, trimmed for a person to read, and compared with nothing.
		return "", fmt.Errorf("%s", strings.TrimSpace(promise.Result().String()))
	default:
		return "", fmt.Errorf("the formatter's promise never settled after %d pumps", maxPumps)
	}
}
