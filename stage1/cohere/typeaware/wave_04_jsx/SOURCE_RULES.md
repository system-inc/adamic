Built: native source visitors for the reserved react/jsx-fragments and react/jsx-no-undef rules; constructed-context source analysis remains unfinished.
Commits: follows 625f0354d881fc0f4bd1b1691b6d12c2775ab358 on codex/typeaware-wave-04; current main 39638d9e2 and area d65a8f931 are ancestors and unchanged.
Commands: both source validators and validate_partial.py pass; full byte comparisons, sanitizers and released-handle checks are described below.
Mutants: each source rule's end+1 mutant compiles, exits zero with clean sanitizer stderr, and differs only at the Go byte comparison; six existing helper/refusal mutants also pass their rejection checks.
Uncovered: constructed-context source analysis, shared lint factory integration, arbitrary JSON option decoding and the full repository gate, including its 17 required-input checks.

The source visitors use the existing native Parser, Rules and raw binding-declarations bridge question. Fragment parsing reuses this unit's already tested wave_04_next/source_context.a and private checker archive. No new checker question, shared registration, harness or compiler file was edited. The driver fetches each visited root once, dispatches only the named listener kinds and hands the node to the rule. Child-kind checks determine semantics, rather than root relevance. No regex matcher was introduced.

Fragments cover qualified React.Fragment, direct/aliased imports, variable and object-binding declarations, exact require module spellings, attribute refusal and both syntax/element modes. Undefined-name analysis covers bare component names, leftmost member references, namespaced/this refusal, file-local declarations, allowGlobals and the .cjs exception. Go declarations are raw input facts; Go supplies no lint decision to the native visitor.

Each rule has 20 source controls. Fragments produce 11 syntax-mode and two element-mode findings; undefined names produce 12 local-mode and ten global-mode findings. All diagnostic fields, spans, messages and zero fix/suggestion counts match the unmodified production rule's independent Go oracle byte for byte. Unicode/trivia controls exercise byte span placement. The 77-root compiler and 287-root repository corpora match in both modes under ASan, UBSan and LeakSanitizer, with empty native stderr. Those corpora produce zero findings for these two rules; the positive source controls establish observable rule behavior.

Single normal-process observations, including program load, parsing and lint:

| Rule | Compiler native / Go | Repository native / Go |
| --- | --- | --- |
| Fragments | 1.755932 / 0.325898 s | 0.297747 / 0.134122 s |
| Undefined names | 1.436578 / 0.300191 s | 0.255645 / 0.134949 s |

These are individual observations, not interleaved repeated benchmarks. Sanitized timings and every subprocess result are in evidence/source-rules/validation.tar.gz.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_jsx/jsx_fragments/validate_source.py /workspace/wave04-fragment-source/check > /workspace/wave04-fragment-source/validate.log 2>&1
python3 stage1/cohere/typeaware/wave_04_jsx/jsx_no_undef/validate_source.py /workspace/wave04-fragment-source/undef-check > /workspace/wave04-fragment-source/undef-validate.log 2>&1
python3 stage1/cohere/typeaware/wave_04_jsx/validate_partial.py /workspace/wave04-fragment-source/partial-final --adamic /workspace/typeaware-wave-04-landing-d65/adamic > /workspace/wave04-fragment-source/partial-final.log 2>&1
```

The validators accept --adamic, --archive and --sanitized-archive overrides. Their defaults use the compiler and private checker archives from this unit's current-area landing. Existing wave_04_next/validate.py constructs those archives without shared edits. Both source runners take tsconfig, absolute-path manifest and mode (syntax/element or local/globals). Tests write output directly to files.

The first undefined-name span mutant had a malformed replacement and failed compilation; it was excluded, corrected, rebuilt and then caught only by comparison. The .cjs control initially used a TypeScript-only project and the native loader refused it. With explicit allowJs:true, both loaders execute the same source and complete findings match under sanitizers. The final validator includes that JavaScript-enabled control; its exact final block was run independently after the preceding complete validator run, with PASS commonjs retained in evidence. No input was skipped.

Both source drivers release the program before making their first source-context query in the released-handle control. Exit 70 and `invalid or released checker handle` are required and observed. Existing helper validation still checks Go/native/sanitizers/source Node/emitted JavaScript, 56 records / 4,941 bytes, three byte mutants and three removed-refusal mutants. Full native source runners use tsgo foreign calls and are not validated as JavaScript execution.

Shared lint integration has a specific remaining boundary: stage1/cohere/lint/context.ts RuleContext has neither a program handle nor a checker-query method, whereas these type-aware visitors take typeaware Rules. The supplied harness Finding/report additions do not add checker access. The owned listener manifests remain metadata, not discoverable factory descriptors; these source runners follow the existing type-aware native/Go oracle pattern. No shared files were edited to conceal that gap. Constructed-context's bounded construction, component, stability and escape source analyses remain owned unfinished work, not HIR-parked work or a fabricated shared blocker. All three claims stay reserved, and no additional claim was made.

No full repository gate was invoked in this increment, so no claim is made about the 17 required-input checks. They were neither skipped, relaxed nor removed. Setup timing on this restored environment remains readiness 0 s per tool/submodule, cache warm 124 s, total 124 s; nproc 5. Prior six completed-rule current-main/current-area re-green evidence remains LANDING_D65.md.
