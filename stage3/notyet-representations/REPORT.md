Built string-intersection, dynamic-object, reference-bearing intersection, and primitive-constraint representations; added nine oracle fixtures.
Implementation SHAs: 82df16f1, 9813a901, 95a72c07; oracle SHAs: 2472b925, cce35809, fc002d84; non-null merge: 3cead3fd.
Focused Node/native/JavaScript oracle run passed in 2.098s; post-merge lower checks passed in 0.749s; 161 example replays completed.
All nine registered representation mutants failed their fixture; four compiler-rule overlay mutants also failed focused checks.
52 listed roots covered by completed rules; remaining kinds are individually recorded in status.csv, including six reserved clock briefs and implementation blockers.

This report supersedes the initial zero-coverage report. The compiler tested and replayed is fc002d84c7f044aa6e932da6759552cd3c26b219. The required area base was b410340dc8f889b5799c3bc519117c63def3aa24. Checked non-null commit c41c0e062e99da37820f822968d4df1b48cdaee7 was merged and pushed as 3cead3fdfeaf8dca6920d748f016a0b0ef1823f2. No existing tests were changed by this unit.

Per-kind outcomes, covered catalog counts, replay counts and exact remaining stops are in status.csv, largest first. Completed rules cover ResolvedConfigFilePath (18), ResolvedConfigFileName (12), object (16), object | undefined (4), IncludeTypeSpaceImports (1), and string | object | undefined (1). These are catalog counts covered by representation rules, not a claim that every root was separately replayed. Reference-bearing intersection rules remove additional signatures, but those kinds lack their own reduced fixtures and are not counted complete. Concrete SolutionBuilderState, generic keys, multimap, and wrapped generic values have Node-held fixtures; their abstract census cases are not claimed cancelled.

The six contribution-clock briefs were pushed at b9a5e6b40504c0e8ebb2b329cc1f9a6b07e6cbd3 as stage3/clock-briefs/representations-01.md through representations-06.md. Reserved kinds: NonNullable<T>, string | null | undefined, Children | undefined, Child, Source, string | null. NonNullable<K> and NonNullable<U> share the first reserved rule and were left untouched.

Replay evidence is under evidence/reopened-before and evidence/reopened-replays, with index.json mapping examples to outputs and stderr logs. Before uses the worker built from 2472b925; after uses the post-merge compiler. The source is the archived latent census adapter, whose 81-file byte and SHA manifest was verified. These are isolated declarations in a checker-rejected entry-root program, not successful whole-program compilation. JSON excerpts omit declaration inventories and retain units, findings and the full transcript SHA256.

Replay command:
  go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/<example> -kind NotYet -reason '<kind>'
Exit 0 means the exact signature was found; exit 1 means it was absent, not necessarily that compilation succeeded. A guarded worker built using stage3/census/latent/make_overlay.py performed the repeated replays. Next named findings are retained in each JSON excerpt.

Fixtures each have their own internal/oracle/representation_*_test.go registration and run through Node and both backends with sanitizers:
  representation_generic_state.a: replace object identity returns with UndefinedObject.
  representation_string_intersection.a: replace branded-string returns with the string 'mutant'.
  representation_object_view.a: replace dynamic-object identity returns with absence.
  representation_object_intersection.a: replace intersection identity returns with absence.
  representation_scalar_constraint.a: return false for the constrained boolean identity.
  representation_string_object_union.a: replace union returns with absence.
  representation_generic_multimap.a: remove all six SetAdd operations across three specializations.
  representation_generic_optional_key.a: replace returns with absence across four specializations.
  representation_generic_wrapped_value.a: replace twelve returns with absence across four specializations.
Each mutant exited cleanly with output that differed from Node. Compiler overlays also removed the string-intersection rule, restored the restrictive object-intersection rule, changed dynamic object storage to plain Object, and removed primitive-constraint fallback. Each failed its focused fixture or direct checker probe. Their logs are in evidence/verification. The initial string mutant used UndefinedString and crashed under sanitizers; it was replaced with the clean string-output mutant before committing.

Final focused command:
  ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/representation_.*\.a|TestRepresentation' -count=1
Result: PASS, 2.098s. Focused lower tests cover scalar constraints, object/tuple views and checked non-null handling; post-merge result PASS, 0.749s. Every fixture group refreshed counts with:
  go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
The final counts refresh passed in 20.386s (the generic group refresh passed in 20.207s). Output was written to logs, never piped. No whole package or full gate was run. See evidence/verification for exact focused commands and outputs recorded by preceding groups.

Unknown remains an implementation blocker, not a design refusal: current Union storage aliases null with undefined. The unknown-to-Union experiment silently changed typeof null to undefined. Accepting unknown needs distinct tags and integrated narrowing rather than the unsafe fallback. No runtime C helpers were added.

The false | RegExpExecArray | null experiment was withheld and its code reverted. Positive output and a stale-narrowing check initially passed; mutants proved that the false/null distinction and a stale-read guard matter. However, a valid nullable narrowing passes through another worker's enumNeverValue union guard, which constructs typeof without null classification. It can trap incorrectly on null. The minimal programs are archived under evidence/false-regex-withheld; they are not registered as supported fixtures. This needs coordination with that function's owner before admission. No unsafe production change remains.

No kind is classified as refused for a ruling: outstanding cases are implementation gaps, reserved work, or insufficient evidence. status.csv does not silently turn absent signatures into census-echo cancellations.

Setup timings: Node 0.057s, Go 0.080s, clang 0.491s, markdown 1.231s, submodules 17.582s, Go build 221.418s, cache 221.514s, total 221.547s. nproc=5, quota=4. GOPROXY used https://proxy.golang.org|direct; environment /workspace/adamic-tools/env.sh. No setup failure. No copied cohere code, no runtime edits, no PR, and pushes only to codex/notyet-representations.
