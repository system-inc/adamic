# Slot 04 helper contracts

Three independently claimed helpers, each in its own `.a` file. These remove 48
listed dependency edges for 36 distinct rules in the frozen readiness ledger.
Only `@next/next/no-img-element` loses its final listed blocker from this slot
alone. See [the conservative residual ledger](slot04_readiness.json) and
[the validation report](slot04_REPORT.md).

| File | Go symbol | Contract |
|---|---|---|
| `jsx_element_parts.a` | `jsx.ElementParts` | Return original tag and attributes links for opening and self-closing elements; other present nodes return two nil links. |
| `tailwind_settings_key.a` | `ClassLiteralSettings.key` | Ordered attribute, callee and variable lists, each name terminated by NUL, lists separated by SOH. |
| `tailwind_compiled_reader.a` | `compiledClassLiteralReader` | One compiled reader per key; first settings win; later hits return the same reader without calling the factory. |

`JsxArena` is the inventory's common AST-adapter contract, not a JSX parser.
Indices identify nodes, `-1` represents a nil link, and kind names omit Go's
`Kind` prefix. `elementParts` requires a present input node, like Go's helper,
which panics on nil. Its result preserves identity through indices. The adapter
must populate raw tag/attributes fields; guessing from generic child order is
not sufficient. The tests project those named fields from Go's parsed trees
independently of the real helper's answers. Invalid arena indices panic.

```a
const parts = elementParts(arena, elementIndex);
const key = classLiteralSettingsKey(settings);
const reader = compiledReaders.reader(key, settings, createReader);
```

`ClassLiteralSettings04` has three readonly lists: `attributeNames`,
`calleeNames`, and `variablePatterns`. Null Go lists adapt to empty lists.
Order, duplicates, empty names, control characters and Unicode are preserved.
This key is the exact upstream encoding, including its collisions when names
contain separator bytes. It is not a new collision-free serialization. Adamic
strings represent decoded configuration text, not arbitrary invalid UTF-8 Go
byte strings.

Create one `CompiledClassLiteralReaderCache<Reader>` per lint run and share it
among all rules. `Reader` is a present object type. Supply the separately owned
Go-compatible `NewClassLiteralReader` factory; this file does not implement
regex compilation or AST class-value reading. A caller with already cached key
`key` gets its earlier reader even if it supplies different settings or a
different factory. Cache instances have separate lifetimes; replacing the
instance starts a new run. The implementation covers serial execution; Go's
concurrent `sync.Map` and racing `LoadOrStore` behavior are not ported here.

The test factory is an identity witness, not a production reader. The Go oracle
calls the real `compiledClassLiteralReader` and records the real factory's
inputs without changing its results. Comparisons check actual pointer sharing,
factory call counts and first input bytes. This isolates caching from the
independent reader-factory dependency and avoids using expected cache results
as adapter input.

## Reproduce

Source the environment printed by `bash cloud/setup.sh` first. From the root:

```
python3 stage1/cohere/lint/helpers/testdata/slot04/capture.py > /tmp/slot04-jsx-capture.log 2>&1
python3 stage1/cohere/lint/helpers/testdata/slot04/capture.py github.com/system-inc/cohere/internal/lint/rules/tailwind.compiledClassLiteralReader > /tmp/slot04-tailwind-capture.log 2>&1
python3 stage1/cohere/lint/helpers/testdata/slot04/readiness.py > /tmp/slot04-readiness.log 2>&1
go test ./stage1/cohere/lint/helpers -run '^TestSlot04' -count=1 -v -timeout=10m > /tmp/slot04-tests.log 2>&1
```

Capture overlays alter no cohere worktree file. Next, React and structure
fixture packages run normally. Tailwind installs pinned `tailwindcss@4.3.3`
under a temporary directory, redirects the two fixture search roots there,
and runs the named fixture and helper tests. External live-repository
population tests remain outside this capture. The generator requires at least
one captured input for every consumer in the ledger. Cache-call inputs are
sorted to remove parallel test scheduling from regeneration.

Captured corpora use pinned cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`.
Boundary corpora include both opening forms, fragments, component references,
malformed source, every settings-list order in a small product, duplicates,
separator characters, Unicode, repeated and colliding keys. Metadata coverage
has an omission mutant. Each production helper has a compiling semantic mutant
checked independently on Node source and sanitized native against Go.
