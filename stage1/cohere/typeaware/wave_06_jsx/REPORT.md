Built: prepared-node kernels for react/jsx-fragments, react/jsx-no-undef and react/no-adjacent-inline-elements, plus numeric rule.json listener declarations. React HIR claims are marked parked.
Commit: claim e8a699b7062f66c4fc0147d4b670fb5fa88d4549 was pushed before implementation; this implementation is pushed to codex/typeaware-wave-06. Main remains f8013f0b.
Checks: 61 source controls, 47 findings and 22459 canonical bytes match independent Go, native, ASan/UBSan, source Node and emitted JavaScript. Listener contract, vet, formatting and Python syntax pass.
Mutants: fragment acceptance, intrinsic-tag classification and inline adjacency mutations compile and exit zero; only independent Go bytes catch each.
Not covered: source parser/driver/checker adapter, full source corpora, production harness integration and new bridge-handle checks. These are partial source ports.

# Selection and ownership

All previous complete wave 06 ports are landing-ready on current main, with full rerun evidence in ../wave_06_react_state/LANDING3_REPORT.md. The three React HIR claims are explicitly parked per Ahra: native source lowering, SSA, captures, compilation-unit selection and memoization handling remain blocked on #dnv6f2c, with JSX support landing on area/stage1-lint. Parking counts as finished for landing-first, not as source parity.

The all-heads fetch checked 529 origin refs and 33 distinct claim documents. Of 197 ranked checker-dependent rules, 154 were claimed and 25 appeared in native sources on main or the bridge base. The first eligible non-analysis entries are the three claimed here, each with zero measured compiler/repository volume. The intervening jsx-no-constructed-context-values was skipped because its memo-stability extension performs function-result and capture/escape analysis in functionEvaluation, anyEscapes and StaysHome. No further rules are claimed.

All changes are inside the owned directory and claim file. New Adamic source is .a. Shared parser, driver, harness, registration, Diagnostic, protected compiler files, bridge and submodule pins are untouched.

# What is implemented

Each handler consumes its supplied node data. Rules never fetch a parser node, never scan the whole node table, and never compare kind strings for relevance. The three rule.json kinds arrays use pinned numeric typescript-go AST values and match the unmodified Go listener tables. The upcoming shared numeric parser/driver must establish that same contract; the current native parser does not yet emit numeric kinds.

Fragments judge the named and shorthand forms separately, default/options direction, attributes, two-segment React.Fragment and four declaration/initializer shapes, including react-only imports and declaration merging. Fixes are deliberately absent in production Go and remain absent here. JSX no-undef implements intrinsic/custom/Unicode names, member-root references, this/name-space decline and local-versus-global declaration origins, with allowGlobals and .cjs behavior. Adjacent inline elements implements the complete 32-name list, JSX node separators, one-finding-per-container adjacency, createElement's array-only gate, pragma/import declarations and literal edge whitespace. Go's Unicode IsSpace differs from JS trim on U+0085 and BOM, so both are independently tested. The pragma binding helper is separate from Fragment binding because Go accepts any react import/property for the former.

Private input models flatten raw syntax and declaration fields. They are a prepared-node boundary, not a production parser or checker bridge. The source entry points explicitly panic NotYet instead of returning empty output. Their native refusal programs all exit 70. No Go finding predicate or verdict is consumed by the native kernels.

# Comparison and mutants

The independent Go oracle is copied into the owned testdata directory with only its selected registry names changed. It loads and parses each source independently, resolves real checker symbols and calls the unchanged production Run methods. Expected records include findings, ranges, all fixes and suggestions. These rules produce no repairs or suggestions in this default-config suite.

The private validator constructs prepared syntax/symbol-origin inputs independently from written fixtures. Spans come from UTF-8 source text searches, not Go diagnostics; declaration records are explicit fixture inputs. This proves the decisions over the supplied facts. It does not prove that a native parser/checker adapter supplies the right facts. Positive findings in every rule prevent zero-count agreement. All 32 inline names are covered, along with fragments, imported/assigned aliases, props, declared/undeclared and global component names, Unicode, member roots, JSX separators and createElement literals. These are 61 valid source controls with 47 findings and 22459 identical bytes. Normal and sanitized native stderr are empty. Source Node and emitted JavaScript match complete Go bytes too.

Semantic mutants alter fragment acceptance, make lowercase intrinsics referenceable, and change the previous/current inline-pair condition. Each mutant builds and exits zero with empty stderr, and independent Go bytes alone catch it. This is separate from metadata checks against the Go listener map. Options branches are implemented, but the Go comparison uses defaults; it does not claim full option-matrix coverage.

The first build refused namespace imports with “stage 0 can't lower reading K yet”. Named numeric imports fix this inside the owned files. The initial failed artifacts are preserved separately. The final validator, go vet ./..., Go formatting and Python parsing pass, with all output directly to logs. The configured toolchain is reused; previous setup took 105s, and current nproc is 5.

# Exact remaining blockers

The native parser currently exposes ParseNode.kind: string and does not produce these JSX trees. The shared kind-indexed driver and conversion into the private input model are not available here. Bare Fragment, component and createElement names also need raw declaration/origin facts wired into that adapter. Since source dispatch cannot yet reach these nodes, no production checker question or runtime handle is added in this partial continuation. Prior bridge released-handle/sanitizer checks remain green on unchanged main; no new-question guard is claimed.

The source analysis entry points therefore refuse. Native/Go byte parity over the 77 compiler and 287 repository roots, shared harness loading, JSX source parsing and end-to-end source timings remain uncovered. The zero-volume corpus ranking is not treated as proof of an empty native source implementation. These claims remain reserved and partially implemented, blocked on shared syntax integration; they are not parked on native HIR.

Prepared-control whole-process observations: native 0.003640s, production Go 0.082772s. Inputs and frontend work differ, so this is not an end-to-end speed comparison or a speedup claim. Complete command arguments, timings, outputs, fixtures and uncompressed SHA-256 hashes are in evidence/index.json.

```sh
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_jsx/validate.py --scratch /workspace/wave-06-jsx-validation2 > /tmp/wave-06-jsx-validation2.log 2>&1
go vet ./... > /tmp/wave-06-jsx-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_06_jsx/testdata > /tmp/wave-06-jsx-format.log
```
