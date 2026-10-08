# Consistent-return helpers: independent subunit

Claim: `07019e9b68798b6be942b4fac54959f187527916`.
Base: `origin/area/stage1-lint` `334509eea8a49b8085187e206c495cc6aa24c5c4`.
Go pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`.
Triage: `origin/lint-helpers/triage` `d7ab0bc4`, rank 26.
No retained partial consistentreturn source was available; these bodies are new.
The already-landed `../property_name.a` supplies property.Name (Static), from
`origin/codex/lint-helpers-02` via the area branch; no duplicate is copied.

| Go symbol | File |
|---|---|
| HasValue | has_value.a |
| IsGenerator | is_generator.a |
| IsScope | is_scope.a |
| Name | name.a |
| ReportRange | report_range.a |
| Verb | verb.a |
| capitaliseFirst | capitalise_first.a |
| isExemptFromEndJudgment | is_exempt_from_end_judgment.a |
| staticName | static_name.a |

The numeric arena in nodes.a preserves kinds, decoded names, nil edges, parent
keys, static/async flags, generator tokens and caller-provided token ranges.
HasValue receives the caller's typed hook explicitly. Capture records its
answer and verifies its invocation count, including all paths that must decline
to invoke it. No checker-hook implementation or typed-rule readiness is claimed.

## Go capture and scope

`testdata/capture.py` renames the nine pinned Go bodies in an oracle-only overlay
and instruments each live invocation from both complete consuming test suites.
The remaining graph/judgment bodies run unchanged in Go to obtain their true
inputs. Both syntax and typed harnesses record every executed case before
returning, including tests with custom span/message assertions. No upstream
cases are omitted because they report nothing. counts.json separates consumer
calls from additional controls. String results compare UTF-8 bytes; booleans,
ranges and hook call counts compare exact protocol bytes. Production helpers
never load Go answers.

Captured distinct source/file/options cases: consistent-return 75 and
@typescript-eslint/consistent-return 56, total 131. Live calls 3,136; with
controls 3,918. All nine have nonzero live coverage. Controls cover constructor,
private/static/async/generator, computed literal/dynamic/numeric keys, arrow
ranges, parent property keys, hook short circuit and false/true answers, all
ASCII first-byte capitalisations and non-ASCII ES5 names.

capitaliseFirst's certified domain is every real caller plus ASCII first-byte
controls. Name always supplies an ASCII kind word at its start. Arbitrary
non-ASCII text passed directly to capitaliseFirst is not certified: Go slices
one UTF-8 byte, while the supported string representation uses UTF-16. The ES5
constructor test preserves the Go byte-slice quirk separately: every non-ASCII
leading name is exempt, even lowercase ones.

## Explicit dependency stop

`cohere/internal/lint/ecmascript/consistentreturn/judgment.go:356` calls
`control_flow_graph.Build(node, control_flow_graph.Hooks[struct{}]{})` in
canRunOffEnd. The graph package has no complete integrated port on area or its
reserved branch `origin/lint-helpers/ecmascript-control-flow-graph`; that
owner's claim lists fourteen missing graph helpers and the recursive graph
lowering probe. This unit does not duplicate that package or replace Go's graph
with a different reachability algorithm or recorded Go answer.

Stopped on canRunOffEnd and its three callers: judgeReturns, judgeScope, Judge.
No rule directory was created: a complete consistent-return port cannot be
proved before that dependency lands. Newly fully unblocked rules: zero.
Conditional future consumers: consistent-return and
@typescript-eslint/consistent-return (the latter also needs its typed hooks).
No compiler/language gap remains in the nine delivered helpers: switch case
blocks and checked array access use plain supported equivalents.

## Commands

Source `/workspace/adamic-tools/env.sh`, then:

```
python3 stage1/cohere/lint/helpers/consistentreturn/testdata/capture.py
go test ./stage1/cohere/lint/helpers -run '^TestConsistentReturnPackageAgreementAndMutants$' -count=1 -v -timeout=30m
go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=60m
go test ./stage1/cohere/lint/helpers/comments -count=1 -v -timeout=30m
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-corpus go test ./stage1/cohere/lint -count=1 -v -timeout=120m
```

The complete lint area already contains the discovery-based JSX inventory fix.
No count map, shared registration, compiler/runtime or other worker file is
changed. Evidence is saved under evidence/ after the gates finish.

Final local gates: owned proof PASS 55.412s; complete helpers PASS 728.781s;
comments PASS 306.216s; complete lint PASS 1616.427s, including all 83 registered
mutants. See evidence/summary.md and compressed full logs. No proof-rule port;
the four graph-dependent helpers remain stopped and reserved.
