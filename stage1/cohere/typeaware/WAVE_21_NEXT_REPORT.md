Built: three native Nexus rules for collection misuse, discarded outcomes and discarded pure results.
Commits: claim 4fe9e445 pushed before code; implementation and evidence c565946d on codex/typeaware-wave-21.
Commands: agreement PASS 61.042s, checker PASS 0.146s, filtered uncached Node oracle PASS 12.830s; vet and gofmt clean.
Mutants: collection, outcome and pure-result judgments caught only by byte comparison; retained-handle mutants caught for both new queries.
Not covered: full repository gate, every compiler configuration or emitted-JavaScript comparison; no further claims after Ahra's correction.

## Selection and scope

All existing work was pushed first, then all origin heads were fetched. The
snapshot contained 320 origin refs. The audit reconstructed the VOLUME_REPORT
ranking from its compiler-all.counts and repository-all.counts, descending
combined counts with lexical ties, excluding the original 26 ports. It skipped
every rule previously named in a claim file, including released reservations.
This interpretation was stated before claiming. The first three fresh rules
were original remaining positions 91, 92 and 93:

- nexus/correctness-no-collection-misuse
- nexus/correctness-no-discarded-outcome
- nexus/correctness-no-discarded-pure-result

Main at ef3d907ecdc4c771b016f7d9c52372def057a340 and the bridge branch at
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6 had no matching native source name or
rule identifier. The complete pre-claim claim texts, ranking and searches are
in validation-wave-21-next/selection.json. Claim 4fe9e445 was pushed before
implementation. No later rules were claimed.

## Implementation

Each rule has its own .a file. Collection checks use constrained union types,
canonical array indexes, actual properties and default-library origins. Pure
results require every method declaration to belong to a permitted library
interface and decline callable, any or unknown arguments. Outcomes require
all arms of a named Nexus union, with exact declaration path and top-level
alias ancestry; awaited calls use the checker's awaited type.

Two checker questions have isolated Go and Adamic files: awaited-shape and
type-declaration-ancestors. They return raw type graphs and syntax ancestry;
all rule decisions remain in Adamic. facts.go has only two dispatch arms,
four gofmt lines added before Ahra's correction. No shared registration
generator, test harness, protected compiler file or Go cohere rule was edited.
The standalone .a runner compiled without needing .a module changes elsewhere.

## Observed validation

The independent Go executable invokes the pinned, unmodified production rules
with its own loader and AST walk. The comparison includes ranges, rule names,
message ids and descriptions, fixes and suggestions, preserving duplicates.
All three production rules have no fixes or suggestions; each observed finding
has zero fields for both, and those fields are compared.

The test reads fixture sources from the three Go tests, including their
preludes: six collection cases, thirty outcome cases and twenty-two pure-result
cases. Two further files cover canonical indexes, literal keys, numeric
boundaries, newer pure methods, Unicode positions, callbacks, parenthesized
union arms, namespaces, wrong origins, narrowed arms and awaited/optional calls.
Four copied Nexus declaration fixtures plus two added declaration fixtures
retain .ts filenames because the production rule matches those filenames.
These are external TypeScript inputs in scratch, not Adamic implementation files.
Every new Adamic implementation, runner and mutant source is .a.

| Population | Root files | Findings | Identical protocol bytes |
| --- | ---: | ---: | ---: |
| Controls | 60 | 71 | 34,896 |
| TypeScript compiler | 77 | 0 | 5,087 |
| Frozen repository | 287 | 0 | 18,485 |

Control findings split into 34 collection, 20 outcome and 17 pure-result
findings. Compiler commit is 050880ce59e30b356b686bd3144efe24f875ebc8.
The repository manifest is the same frozen population used by the existing
volume and coverage work; it does not add these new implementation files.
All three populations also passed ASan, UBSan and LeakSanitizer comparison,
with empty native stderr.

| Mutant | Changed decision | Observation |
| --- | --- | --- |
| Collection | value <= 0 becomes value < 0 | exit 0, byte comparison catches byte 1435 |
| Outcome | complete arm count equality becomes inequality | exit 0, byte comparison catches byte 7809 |
| Pure result | MethodSignature equality becomes inequality | exit 0, byte comparison catches byte 18010 |
| Retained checker registry, awaited-shape | omit live-handle deletion | exit 0; required released-query panic 70 catches it |
| Retained checker registry, type-declaration-ancestors | omit live-handle deletion | exit 0; required released-query panic 70 catches it |
| Existing Node oracle one-byte mutant | compiler output mutation | TestTheOracleCatchesOneByte passes by detecting it |

Both normal released-query probes exit 70 with exactly
`adamic: panic: invalid or released checker handle` on stderr. The ancestry
probe registers its type identity before release. Three judgment mutants
compile and exit normally with empty stderr; only finding bytes reject them.

The first native build refused unsupported Number(...) lowering; using the
supported Number.parseFloat(...) corrected it. The initial fixture config's
empty files array was rejected; the final config includes the declaration
inputs. An initial filtered Node invocation selected no child fixtures, so it
was corrected and rerun uncached. The final Node log shows nine child fixtures
and the one-byte mutant.

## Commands and costs

The toolchain is the already completed cloud/setup.sh run: total 81 seconds,
go/clang/node/submodules ready at 0 seconds, stage0 cache ready at 81 seconds.
nproc remains 5; the cgroup grants 4 cores. Go 1.27.1, clang 20.1.8 and Node
24.19.0 are unchanged.

```
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE21_NEXT_ARTIFACTS=/workspace/wave21-next-artifacts \
ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json \
ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest \
ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave21NextAgreement$' \
    -count=1 -v > /workspace/wave21-next-test.log 2>&1

go test ./bridge/tsgo/checker -count=1 > /workspace/wave21-next-checker.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave21-next-vet.log 2>&1

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
    -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|method_closures|generic_functions|number_parsing)\.a$' \
    -count=1 -timeout=10m -v > /workspace/wave21-next-node.log 2>&1
```

Targeted packages and the filtered external oracle were run in place of the
full multi-minute gate. Formatting checks and git diff --check were clean.
The final logs, compressed subprocess output and input/output hashes are in
validation-wave-21-next; native stderr hashes cover successful empty outputs.

One quiet process run per implementation and corpus, after correctness checks:

| Corpus | Native wall seconds | Go wall seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 2.506309 | 0.774367 | 3.24 |
| Repository | 0.299993 | 0.138314 | 2.17 |

These are whole-process observations including loading and parsing, not a
statistical benchmark. Native was slower in these runs. No claim of a speedup
or of parity outside the measured inputs and configurations is made.
