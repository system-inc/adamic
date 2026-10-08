Built a shared exact-shape proof for fresh const Node option objects, including parentheses and as const.
Branch codex/host-stats-dirent-union, from 70a2a970; this checkpoint is the const-options implementation commit.
Lower/IR/flow and Linux counts pass; six unchanged combined host fixtures agree with Node on native and JavaScript.
All four requested proof mutants and an extra hidden-initializer cast mutant are caught by the direct predicate assertion.
Whole oracle with WASI is running at this checkpoint; remaining inline cases are listed below.

The compiler hook at internal/lower/optional_widening.go:135 is unchanged.
Both the exemption and the fs-file field reader call nodeHostExactOptions.
Source-declared scalar fields are read from their declared type; a missing
optional field defaults only after the exact absence proof. No wider object is
stored, returned or passed to the runtime. Existing named NotYet overload and
encoding restrictions remain.

A binding must be an unannotated const with a fresh object-literal initializer.
Only parentheses and as const are unwrapped; structural casts, aliases, function
returns, parameters, let, spreads, methods and object-valued options fail.
Symbol identity distinguishes shadowed names. The source file is checked for
writes (including destructuring and loop targets), deletion, updates, receiver
method calls, aliases, storage, prior calls and captures. Declared scalar reads
are allowed. Exported objects fail the proof.

Assumption, following the request that the six real sys.ts fixtures turn green:
a private module const captured only as this proven consuming host argument is
consumption, not an earlier escaping capture. Every other capture is refused,
including closures returning the object and local captured bindings. For a module
consumer, every other object use is rejected regardless of text order: an escape
written after the function declaration can still precede its invocation.
This exception never admits general closure capture or a wider stored view.

The scratch combined proof is 3de95428, the existing merge of d18d2a4 with the
Stats/Dirent runtime dispatch fix. Only this library proof and the variable-options
condition were applied; its newer boxed-constant handling was preserved.
No host fixture, status.json, compiler refusal hook or language rule was changed.
The exact Node stdout, stderr and exit status were checked for each:

| Fixture | Native | JavaScript |
| --- | --- | --- |
| 06_fileExists | green | green |
| 07_directoryExists | green | green |
| 08_getDirectories | green | green |
| 10_getModifiedTime | green | green |
| 13_createDirectory | green | green |
| 24_useCaseSensitiveFileNames | green | green |

The new node_options_const_capture.a fixture holds the module consumer to Node
on native, JavaScript and WASI. The existing node_options_widening.a uses a
parenthesized as const initializer. Its second host call now uses a fresh inline
literal: the direct binding has already been passed to a call. Node output
before/after is byte-identical: directory followed by absent=true.
node_fs_file_stat.a also uses a parenthesized as const initializer and agrees
on native/JavaScript. No other existing fs fixture was rewritten in this unit.

Inline literals still needed in existing sources:

- node_options_widening.a: the second direct call, because the binding was used
  by an earlier call.
- node_fs_directory_entries.a and node_fs_directory_union.a: readdirSync forwards
  its options object, so it has no consumption exemption.
- node_fs_file_read.a: its encoding option remains inline because the host reader
  supports constant UTF-8 encodings, not encoding fields read dynamically from a
  binding. This is the existing named NotYet limitation, not an absence failure.

The six host fixtures keep their original module const and need no inline rewrite.
The remaining previously green 17 fixtures need no new adaptation for this proof.

Production-source mutants use isolated Go build overlays. Each intended test
exits 1 specifically with consumes argument = true, want false; compile failures
are not counted as kills:

| Mutant | Assertion |
| --- | --- |
| Remove the const-only requirement | let_mutant |
| Ignore prior escapes | prior_escape_mutant |
| Accept a function-return initializer | returned_object_mutant |
| Skip binding-use analysis after proving a fresh literal | cast_planted_field_mutant |
| Additionally accept a structural cast initializer | cast_hidden_field_mutant |

The fourth witness starts with {recursive:false}, then plants force='yes' through
a cast before rmSync. Node reads that runtime field and prints:

    The "options.force" property must be of type boolean. Received type string ('yes')

This differs from defaulting force to false and throwing ENOENT. The direct
predicate probe prevents an independent cast refusal from masking a bad host
classification. Separate complete compiler tests also verify the let, escaped,
returned and hidden-initializer programs receive adamic/no-optional-widening.

Completed commands and output at checkpoint:

    go test ./internal/lower ./internal/ir ./internal/flow -count=1 -timeout=30m
    lower 234.982s; ir 5.852s; flow 324.431s; all ok

    go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel=2 -timeout=30m -args -update-counts
    ok internal/oracle 223.492s

Linux counts add the new row 5/5/1/7/5/0. Directory-system enumeration also grows
by the new source file: 2932/2932/5206/4374/1722/0. No macOS counts were used.
Focused native/JavaScript/WASI comparisons pass in 3.274s. The four required
mutants plus the extra assertion-cast mutant are all observed failures.
The complete configured oracle command is:

    ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -parallel=2 -timeout=30m

WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot and
GOPROXY='https://proxy.golang.org|direct' are set. Test output is written to logs.

An environment restart interrupted earlier final gates and the first combined
compiler build; those incomplete runs are not passes. The first six native
builds then failed only because the root filesystem was full (29 GB Go build
cache). Clearing that disposable cache recovered 27 GB; the complete six-fixture
retry passes. Sources and pinned dependencies were preserved.

Only the owned branch is pushed. No main push, force push, rebase or PR is used.
