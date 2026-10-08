Added two-binary --compare mode and measured the exact pinned typescript-go build.
Commit base 6f744d6d; Go build matches performance commit afb0651c by executable SHA256.
Go passes: acceptance 60/301, tiny 0/1, upstream 11,212/13,453 against unchanged Node expectations.
Mutants: 241 acceptance, one tiny and 33 upstream one-byte stdout changes caught; 17 focused tests pass.
Not covered: native tsc; 834 upstream exclusions remain; Node upstream captures were revalidated and reused.

# TypeScript-Go agreement

| Suite | Node oracle passes | Go oracle passes | Disagreements |
|---|---:|---:|---:|
| acceptance | 301/301 | 60/301 | 241 |
| tiny | 1/1 | 0/1 | 1 |
| baselines | 13453/13453 | 11212/13453 | 2241 |

The 2,483 disagreements count suite runs: tiny is repeated in acceptance. Removing the separate tiny repeat leaves 2,482 measured runs. Exact stdout, stderr and exit remain the verdict. No expected source, reference, golden, selection or status was changed. No CLI wrapper is used. Go diagnostics use its installed libraries, as the ordinary harness specifies.

## Every disagreement grouped

Primary groups are disjoint. A removed/unsupported CLI option takes precedence; identical diagnostics with changed exit count as exit status. Equal complete diagnostic block multisets permit an ordering classification. Equal header multisets with changed bodies count as wording/type-display/presentation. Equal code/category multisets with moved locations count as diagnostic locations. Remaining changed codes, categories or multiplicities count as checker diagnostic sets. The classifier describes observed diagnostics, not a proof that all changed sets are Go bugs: compiler version, parser, resolver and declaration behavior also belong to that group. Secondary changed streams remain recorded for every row.

| Cause | Acceptance | Tiny | Upstream | Total |
|---|---:|---:|---:|---:|
| unsupported CLI option/value | 0 | 0 | 1361 | 1361 |
| checker diagnostic set | 7 | 0 | 400 | 407 |
| diagnostic locations | 2 | 0 | 25 | 27 |
| message wording/presentation | 6 | 0 | 158 | 164 |
| diagnostic ordering | 0 | 0 | 0 | 0 |
| exit status | 226 | 1 | 297 | 524 |
| stderr | 0 | 0 | 0 | 0 |
| timeout | 0 | 0 | 0 | 0 |
| non-diagnostic output | 0 | 0 | 0 | 0 |

No pure diagnostic-ordering, stderr, timeout or non-diagnostic-output disagreement was observed, so those groups have no examples. Every nonzero group has three examples below; exact captures accompany them in evidence/typescript-go/examples/.

## Three examples per nonzero group

### exit status: acceptance 061_constEnumErrors 

Node exit 2; Go exit 1. Changed streams: exit.

All diagnostic bytes match; only the process exit differs.

[Exact captures](evidence/typescript-go/examples/exit-status/1/case.json)

### exit status: acceptance 062_interfaceDeclaration1 

Node exit 2; Go exit 1. Changed streams: exit.

All diagnostic bytes match; only the process exit differs.

[Exact captures](evidence/typescript-go/examples/exit-status/2/case.json)

### exit status: acceptance 063_didYouMeanSuggestionErrors 

Node exit 2; Go exit 1. Changed streams: exit.

All diagnostic bytes match; only the process exit differs.

[Exact captures](evidence/typescript-go/examples/exit-status/3/case.json)

### message wording/presentation: acceptance 064_recursiveFunctionTypes 

Node exit 2; Go exit 1. Changed streams: stdout, exit.

Node diagnostic bytes around the first difference:

```text
es.ts(43,4): error TS2769: No overload matches this call.
  Overload 1 of 4, '(a: { (): typeof f7; (a: typeof f7): () => number; (a: number): number; (a?: typeof f7 | undefined): typeof f7; }): () => number', gave the following error.
    A
```

Go:

```text
es.ts(43,4): error TS2769: No overload matches this call.
  The last overload gave the following error.
    Argument of type 'string' is not assignable to parameter of type '{ (): typeof f7; (a: typeof f7): () => number; (a: number): number
```

[Exact captures](evidence/typescript-go/examples/message-wording-presentation/1/case.json)

### message wording/presentation: acceptance 080_builtinIterator 

Node exit 2; Go exit 1. Changed streams: stdout, exit.

Node diagnostic bytes around the first difference:

```text
ber, boolean>' is not assignable to parameter of type 'Iterator<string, unknown, undefined> | Iterable<string, unknown, undefined>'.
  Type 'Generator<string, number, boolean>' is not assignable to type 'Iterator<string, unknown, undefined>
```

Go:

```text
ber, boolean>' is not assignable to parameter of type 'Iterable<string, unknown, undefined> | Iterator<string, unknown, undefined>'.
  Type 'Generator<string, number, boolean>' is not assignable to type 'Iterator<string, unknown, undefined>
```

[Exact captures](evidence/typescript-go/examples/message-wording-presentation/2/case.json)

### message wording/presentation: acceptance 097_mappedTypeGenericWithKnownKeys 

Node exit 2; Go exit 1. Changed streams: stdout, exit.

Node diagnostic bytes around the first difference:

```text
Property 'unknownLiteralKey' does not exist on type 'Record<keyof Shape | "knownLiteralKey", number>'. Did you mean 'knownLiteralKey'?
mappedTypeGenericWithKnownKeys.ts(10,5): error TS2862: Type 'Record<keyof Shape | "knownLiteralKey", numb
```

Go:

```text
Property 'unknownLiteralKey' does not exist on type 'Record<"knownLiteralKey" | keyof Shape, number>'. Did you mean 'knownLiteralKey'?
mappedTypeGenericWithKnownKeys.ts(10,5): error TS2862: Type 'Record<"knownLiteralKey" | keyof Shape, numb
```

[Exact captures](evidence/typescript-go/examples/message-wording-presentation/3/case.json)

### diagnostic locations: acceptance 082_unusedDestructuring 

Node exit 2; Go exit 1. Changed streams: stdout, exit.

Node diagnostic bytes around the first difference:

```text
lared but its value is never read.
unusedDestructuring.ts(6,7): error TS6133: 'e' is declared but its value is never read.
unusedDestructuring.ts(7,7): error TS6133: 'g' is declared but its value is never read.
unusedDestructuring.ts(8,1): 
```

Go:

```text
lared but its value is never read.
unusedDestructuring.ts(6,9): error TS6133: 'e' is declared but its value is never read.
unusedDestructuring.ts(7,12): error TS6133: 'g' is declared but its value is never read.
unusedDestructuring.ts(8,1):
```

[Exact captures](evidence/typescript-go/examples/diagnostic-locations/1/case.json)

### diagnostic locations: acceptance 178_bitwiseCompoundAssignmentOperators 

Node exit 2; Go exit 1. Changed streams: stdout, exit.

Node diagnostic bytes around the first difference:

```text
bitwiseCompoundAssignmentOperators.ts(3,1): error TS2447: The '^=' operator is not allowed for boolean types. Consider using '!==' instead.
bitwiseCompoundAssignmentOperators.ts(7,1): error TS2362: The left-hand side of 
```

Go:

```text
bitwiseCompoundAssignmentOperators.ts(3,3): error TS2447: The '^=' operator is not allowed for boolean types. Consider using '!==' instead.
bitwiseCompoundAssignmentOperators.ts(7,1): error TS2362: The left-hand side of 
```

[Exact captures](evidence/typescript-go/examples/diagnostic-locations/2/case.json)

### diagnostic locations: baselines tests/cases/compiler/bitwiseCompoundAssignmentOperators.ts 

Node exit 2; Go exit 2. Changed streams: stdout.

Node diagnostic bytes around the first difference:

```text
bitwiseCompoundAssignmentOperators.ts(3,1): error TS2447: The '^=' operator is not allowed for boolean types. Consider using '!==' instead.
bitwiseCompoundAssignmentOperators.ts(7,1): error TS2362: The left-hand side of 
```

Go:

```text
bitwiseCompoundAssignmentOperators.ts(3,3): error TS2447: The '^=' operator is not allowed for boolean types. Consider using '!==' instead.
bitwiseCompoundAssignmentOperators.ts(7,1): error TS2362: The left-hand side of 
```

[Exact captures](evidence/typescript-go/examples/diagnostic-locations/3/case.json)

### checker diagnostic set: acceptance 106_interfaceMergeWithNonGenericTypeArguments 

Node exit 2; Go exit 1. Changed streams: stdout, exit.

Node diagnostic bytes around the first difference:

```text
s(4,34): error TS2315: Type 'SomeBaseClass' is not generic.
interfaceMergeWithNonGenericTypeArguments.ts(6,3): error TS2346: Call target does not contain any signatures.

```

Go:

```text
s(4,34): error TS2315: Type 'SomeBaseClass' is not generic.

```

[Exact captures](evidence/typescript-go/examples/checker-diagnostic-set/1/case.json)

### checker diagnostic set: acceptance 116_recursivelyExpandingUnionNoStackoverflow 

Node exit 2; Go exit 1. Changed streams: stdout, exit.

Node diagnostic bytes around the first difference:

```text
recursivelyExpandingUnionNoStackoverflow.ts(3,10): error TS2589: Type instantiation is excessively deep and possibly infinite.
recursivelyExpandingUnionNoStackoverflow.ts(3,10): error TS2615: Type of property 'M' circularly references itsel
```

Go:

```text
recursivelyExpandingUnionNoStackoverflow.ts(3,10): error TS2615: Type of property 'M' circularly references itself in mapped type '{ [P in "M"]: any; }'.

```

[Exact captures](evidence/typescript-go/examples/checker-diagnostic-set/2/case.json)

### checker diagnostic set: acceptance 177_readonlyTupleAndArrayElaboration 

Node exit 2; Go exit 1. Changed streams: stdout, exit.

Node diagnostic bytes around the first difference:

```text
readonlyTupleAndArrayElaboration.ts(10,20): error TS2345: Argument of type 'readonly [3, 4]' is not assignable to parameter of type '[number, number]'.
  The type 'readonly [3, 4]' is 'readonly' and cannot be assigned to the mutable
```

Go:

```text
readonlyTupleAndArrayElaboration.ts(10,20): error TS4104: The type 'readonly [3, 4]' is 'readonly' and cannot be assigned to the mutable type '[number, number]'.
readonlyTupleAndArrayElaboration.ts(13,8): error TS4104: The type 'rea
```

[Exact captures](evidence/typescript-go/examples/checker-diagnostic-set/3/case.json)

### unsupported CLI option/value: baselines tests/cases/compiler/SystemModuleForStatementNoInitializer.ts 

Node exit 2; Go exit 2. Changed streams: stdout.

Node diagnostic bytes around the first difference:

```text
error TS5107: Option 'module=System' is deprecated and will stop functioning in TypeScript 7.0. Specify compilerOption '"ignoreDeprecations": "6.0"' to silence this error.

```

Go:

```text
error TS5095: Option 'bundler' can only be used when 'module' is set to 'preserve', 'commonjs', or 'es2015' or later.
error TS5108: Option 'module=System' has been removed. Please remove it
```

[Exact captures](evidence/typescript-go/examples/unsupported-option-value/1/case.json)

### unsupported CLI option/value: baselines tests/cases/compiler/abstractPropertyBasics.ts target=es5

Node exit 2; Go exit 2. Changed streams: stdout.

Node diagnostic bytes around the first difference:

```text
error TS5107: Option 'target=ES5' is deprecated and will stop functioning in TypeScript 7.0. Specify compilerOption '"ignoreDeprecations": "6.0"' to silence this error.

```

Go:

```text
error TS5108: Option 'target=ES5' has been removed. Please remove it from your configuration.

```

[Exact captures](evidence/typescript-go/examples/unsupported-option-value/2/case.json)

### unsupported CLI option/value: baselines tests/cases/compiler/accessorWithLineTerminator.ts target=es5

Node exit 2; Go exit 2. Changed streams: stdout.

Node diagnostic bytes around the first difference:

```text
error TS5107: Option 'target=ES5' is deprecated and will stop functioning in TypeScript 7.0. Specify compilerOption '"ignoreDeprecations": "6.0"' to silence this error.

```

Go:

```text
error TS5108: Option 'target=ES5' has been removed. Please remove it from your configuration.

```

[Exact captures](evidence/typescript-go/examples/unsupported-option-value/3/case.json)

## Provenance and commands

The native-preview package is 7.0.0-dev.20260707.2, package gitHead 9977d6d38fcc78de8ae71770f3aa08256e6cc861. Its binary SHA256 is 138b9ab195ae0ad372dae2f952090e07d4c504bbccf63240172ae843116009ce, identical to performance evidence at afb0651c. The original lockfile, including npm integrities, is committed under toolchains/typescript-go/.

Installed library collections have 108 shared filenames, 6 with different bytes. The primary measurement preserves each installation; these differences can influence checking or type display. Their SHA256 maps are retained. The three checker-diagnostic examples were additionally rerun with the exact Node 6.0.3 library files next to a byte-identical Go executable. All three retained the same Go diagnostics/stderr/exit and still differed from Node, proving those examples are checker diagnostic behavior rather than bundled-library differences. This control is supplemental and does not change primary pass counts. Other grouped cases remain classified by observed diagnostic shape, not an asserted internal bug cause. [Control proof](evidence/typescript-go/checker-controls/proof.json).

Node acceptance and tiny ran fresh and passed 301/301 and 1/1. The prior unit’s 13,453 passing upstream captures were reused only after verifying the unchanged CLI artifact SHA256, source/reference/expected-summary hashes, header options, and every diagnostic/stderr/exit byte. Their original capture locations are recorded. An initially started redundant full Node baseline run was stopped after successful reuse verification; its partial log is retained and is not claimed as a completed run. Go ran all suites fresh and exited 1 for the measured differences, with no harness errors.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/verdict-go-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix /tmp/verdict-tsgo-pin --ignore-scripts --no-audit --no-fund > /tmp/verdict-tsgo-install.log 2>&1
STAGE3_VERDICT_UPSTREAM=/tmp/stage3-verdict-adapted stage3/verdict/run.sh --tsc /tmp/verdict-tsgo-pin/node_modules/@typescript/native-preview-linux-x64/lib/tsgo /tmp/verdict-compare-go > /tmp/verdict-compare-go.log 2>&1
python3 stage3/verdict/comparison.py /tmp/verdict-compare-node-reused /tmp/verdict-compare-go /tmp/verdict-node-go-final > /tmp/verdict-node-go-final.log 2>&1
stage3/verdict/run.sh --compare stage3/verdict/standins/node.sh stage3/verdict/standins/mutant.sh --baseline-limit 40 /tmp/verdict-compare-mode-mutant > /tmp/verdict-compare-mode-mutant.log 2>&1
python3 stage3/verdict/test_comparison.py > /tmp/verdict-comparison-tests.log 2>&1
python3 stage3/verdict/test_verdict.py > /tmp/verdict-go-existing-tests.log 2>&1
```

Public --compare runs two compilers and prints the agreement table; each child retains its ordinary verdict. Its real Node/mutant proof exits 1 with stdout-only differences in all three suites. The final cause classifier was rechecked against those same captures after taxonomy changes. Five focused comparison tests plus twelve existing verdict tests pass; byte, stderr, exit, ordering, location, source identity, timeout and identical-wrong-output mutants are caught. No Adamic oracle fixture was added, so counts.md is unchanged. No whole packages or full gate were run.

Setup timing lines: Node 0.024s, Go 0.024s, markdown 0.079s, submodules 0.084s, clang 0.183s, Go build 36.389s, deferred tests 36.489s, cache 36.490s, done 36.523s. nproc=5, cgroup CPU quota=4. All output went to log files.

[Comparison JSON with every disagreement](evidence/typescript-go/comparison.json) · [Go verdict](evidence/typescript-go/go-summary.json) · [Mutant proof](evidence/typescript-go/mutant-proof.json)
