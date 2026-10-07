Reviewed the new regex contract; runtime pattern construction blocks the claimed rule.
Source remains f228409a on current main f8013f0b; this update changes evidence only.
Setup completed in 25 s, nproc 5; the owned declaration check PASS 0.008 s.
All nine wrong-kind declaration mutants were caught; prior semantic gates remain recorded.
No new claims: dynamic RegExp, compatible errors, JSX, labels and numeric dispatch remain gaps.

The landing-first fetch confirmed origin/main is still
f8013f0baac41ddc340d76f83bddde38536a8f07 and is an ancestor of the published
f228409aabecf67b5f3882911b02de76bc9a9c64. This is the only branch published
by this unit. The preceding all-owned byte oracle, sanitizers, released-handle,
bridge and filtered Node results are in WAVE_14_PATTERN_REPORT.md. No semantic
source changed here, so those full gates were not repeated. The owned listener
check was rerun and caught all nine [999999] declaration mutants.

The effective regex instruction prevents expanding the custom native pattern
validator as the path to completing no-invalid-regexp. Existing validation is
partial and is not claimed to satisfy the new contract. No new matcher or
regex translation was written. Fetching all origin branches found no
origin/codex/lint-regex ref, so its shared translation row could not be read.

The production Go rule does not itself contain a Go regexp literal to translate:
invalidPatternMessage calls the shared ECMAScript regexp compiler with a pattern
read from the linted string literal. A native equivalent needs runtime pattern
construction and recoverable validation errors. The minimal .a probe declares
compilePattern(pattern: string), constructs new RegExp(pattern, 'u') and tests
'a'. With a string console argument, Adamic refuses:

    adamic: /workspace/wave-14-regex-contract/dynamic.a:2:35: stage 0 can't lower RegExp with a nonconstant pattern yet

The build exits 1. The equivalent JavaScript source on Node exits 0 and prints
true. The initial probe was corrected for the CLI's required -o argument and
Adamic's string-only console signature before recording this isolated refusal.
Compiler lowering in internal/lower/regexp.go explicitly requires
constantPattern; editing that shared compiler implementation is outside this
unit's authorized rule scope.

Constructing a JS RegExp alone also does not supply byte-identical production Go
messages. The existing Go-positive '(' witness reports:

    Invalid regular expression: /(/: missing closing ) in `(`

Node new RegExp('(', 'u') reports:

    Invalid regular expression: /(/u: Unterminated group

Both exact streams are retained. A shared runtime validator/error interface or
approved error compatibility layer is required before this incomplete native
rule can be declared complete. This is not an analysis-dependent React parked
claim, and not a claim that the rule is finished. Leaked-number-render's actual
remaining dependency is JSX parsing, not high-level IR, single assignment or
capture analysis. Undefined labels and handed-node numeric dispatch likewise
remain shared parser/driver gaps. No shared files were edited or reverted.

Commands, each writing directly to retained logs:

    bash cloud/setup.sh
    source /workspace/adamic-tools/env.sh
    go run ./cmd/adamic build /workspace/wave-14-regex-contract/dynamic.a -o /workspace/wave-14-regex-contract/dynamic
    node --input-type=module -e "function compilePattern(pattern) { const expression = new RegExp(pattern, 'u'); return expression.test('a'); } console.log(compilePattern('a'));"
    node --input-type=module -e "try { new RegExp('(', 'u'); } catch (error) { console.log(error.message); }"
    go test ./stage1/cohere/typeaware -run '^TestWave14NumericListenerDeclarations$' -count=1 -v

Setup printed Go/clang/Node/submodules ready at 0 s, cache warm at 25 s and done
at 25 s on 5 processors, cpu.max 400000 100000, 17.6 GB. No timing benchmark
was repeated; the last native/Go compiler medians remain 1.708461/0.322694 s,
5.29x, with no numeric dispatch speedup claimed. No batch-8 Diagnostic SHA
was supplied. No developer-tools leak-helper diff appeared in this fetch.
