# What a native binary can't do, and what Adamic does instead

A native binary answers to the oracle, Node running the source, byte for byte on stdout, stderr and exit
code. A few things Node does from the source have no answer a compiled program can reproduce. This page
lists each, what Adamic does there, and which check holds that answer. The rule is the one
[wasm-limits.md](wasm-limits.md) keeps: Adamic does the same thing another way that a ruling names and a
test pins on every side, or stops loudly. It never does something different quietly.

Each difference here is a row of `ruledBackendDivergences` (`internal/oracle/backend_divergences_test.go`).
A row pins Node's outcome, native's and the JavaScript backend's, each byte for byte, so membership never
excuses a change on any side. wasm32 is a native build and is held to the row's native outcome.

## Loading CommonJS modules

An Adamic program is an ES module, and a native binary has no module loader. tsc's `sys.ts` requires two
things inside a `try` whose `catch` handles a failure, so those two requires lower as the throw the catch
sees (Adamic's ruling, task #redxes2, Nov 3). Every other `require` stops with a named NotYet: a literal one
of anything else is never loaded, so this is never a way to run code.

### installed-layout

`require("source-map-support")` throws what Node throws where tsc 6.0.3 is installed, since it ships no
dependencies: an `Error` with code `MODULE_NOT_FOUND` and the message `Cannot find module
'source-map-support'`. Node's message goes on with a `Require stack:` naming the requiring file, and its error
has a `requireStack` own property holding the same path. That path is the running module's own, which has no
reproducible native answer yet, so native gives the name, the code and the message's first line only. tsc's
catch ignores the error, so tsc observes nothing of the difference.

Held by: `node_commonjs_require_missing.a` and its uncaught variant, and `TestInstalledLayoutIsNodes`, which
measures Node's answer from a `typescript/lib/_tsc.js` laid out as npm installs it, so a change in Node's
answer fails it.

### no-javascript-plugins

A `require` of a path computed as the program runs is tsc's plugin load. It evaluates the path, then throws an
`Error`, with no code, whose message is `native tsc can't load JavaScript modules: ` and the path. Stock Node
would load the plugin; a native binary can't run JavaScript. tsc's catch returns the error as the load's
result.

Held by: `node_commonjs_require_computed.a` and its uncaught variant.

## The source on Node

The oracle runs every program as an ES module (`oracle/node.mjs`), where `require` isn't defined: the source
of each fixture above throws a `ReferenceError` there, before a computed path is ever evaluated. That is
pinned in each row too, as Node's side.
