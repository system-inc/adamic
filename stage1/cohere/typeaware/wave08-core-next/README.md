# Wave 08 core continuation

The claim was pushed in 2ca1f5401 before implementation. This directory owns
`require-atomic-updates`, `require-await`, and `symbol-description`. React
reservations are parked in the claim file. No shared registration generator,
harness, parser, or compiler implementation is edited.

Each rule exports numeric `listenerKinds`; the listener manifests retain the
shared generator's SyntaxKind names. The owned profile indexes those numeric
kinds and passes cached nodes to listeners. The current parser still exposes
strings, so `numeric_node.a` converts them once outside the rules. Related-node
lookups read the cache. Atomic updates listens to SourceFile, as Go does, and
builds native flow for that file's suspending functions.

Atomic updates ports the pinned ECMAScript CFG, adding read and post-evaluation
write hooks, deferred suspension/property-store placement, symbol-based escape
scanning, two-set fixed-point flow, and assignment-position deduplication.
`--allow-properties` suppresses only property findings. The owned CFG uses the
parser's five-child conditional layout; its inherited process CFG remains
untouched. Await ports syntax exemptions, generic context-path replay,
independent type-parameter substitutions, contextual and heritage contracts,
thenability, exact head spans, naming, and semicolon-safe removeAsync suggestions.
Symbol uses the first original declaration's declaration-file flag, including
Go's decision not to follow aliases.

Four raw questions live here in matching Go and Adamic files:

- `symbol_origins`: ordered original declaration-file flags.
- `syntax_snapshot`: numeric syntax metadata and identifier symbol identities.
- `await_contract`: contextual/heritage types and declared/resolved signatures.
- `await_type`: union/apparent/index/property edges and call signatures.

They return facts, never lint verdicts, CFGs, escape decisions, or thenability.
`overlay_bridge.py` registers them only in an owned scratch dispatcher. This
preserves the instruction to keep changes inside owned rule directories.
Integration must install their production dispatcher hooks and shared factories;
these profiles are not represented as installed shared-harness rules.

Build and compare (source the cloud toolchain environment first):

```sh
python3 stage1/cohere/typeaware/wave08-core-next/overlay_bridge.py /tmp/wave08-core-bridge
CC=clang go build -overlay /tmp/wave08-core-bridge/overlay.json -buildmode=c-archive -o /tmp/wave08-core-bridge/checker.a ./bridge/tsgo/archive
CC=clang CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all' go build -overlay /tmp/wave08-core-bridge/overlay.json -buildmode=c-archive -o /tmp/wave08-core-bridge/checker-asan.a ./bridge/tsgo/archive
adamic build stage1/cohere/typeaware/wave08-core-next/suite.a -o /tmp/wave08-core-native --tsgo /tmp/wave08-core-bridge/checker.a
adamic build stage1/cohere/typeaware/wave08-core-next/suite.a -o /tmp/wave08-core-asan --tsgo /tmp/wave08-core-bridge/checker-asan.a --sanitize
python3 stage1/cohere/typeaware/wave08-core-next/capture_core.py /tmp/wave08-core-fixtures
python3 stage1/cohere/typeaware/wave08-core-next/compare_core.py --artifacts /tmp/wave08-core-compare --fixtures /tmp/wave08-core-fixtures --native /tmp/wave08-core-native --compiler-root /path/to/TypeScript
python3 stage1/cohere/typeaware/wave08-core-next/validate_mutants.py --artifacts /tmp/wave08-core-mutants --baseline /tmp/wave08-core-compare --stage0 /path/to/adamic --archive /tmp/wave08-core-bridge/checker-asan.a --fixtures /tmp/wave08-core-fixtures
```

The profile defaults to Symbol; `--atomic` or `--await` selects the other rule.
Test output must be redirected to log files. `compare_core.py` compares the
production registry's complete finding/fix/suggestion bytes. Capturing preserves
upstream assertions; one deliberately checker-free Symbol guard witness is
excluded from the typed native comparison. The three typed profiles cover 254
upstream programs: 94 atomic / 59 findings, 126 await / 97 findings, and 34 Symbol
/ 13 findings. All three also compare the frozen 77 compiler and 287 repository
files; those corpora produce zero findings for this zero-volume batch. Additional
Symbol controls cover unfollowed imports, source ambient declarations, Unicode,
comments, explicit type arguments and optional calls. Corpus emptiness is not
used as evidence that the rules fire; the positive controls and mutants do that.
