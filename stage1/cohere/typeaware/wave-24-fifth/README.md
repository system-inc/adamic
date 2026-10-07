Built require-await, symbol-description and valid-typeof as native Adamic .a modules with numeric handed-node listeners.
Claim e21595a11 was pushed before implementation; current-main landing refresh follows this implementation commit.
217 valid controls produce 125 findings with identical full diagnostic/fix/suggestion bytes; 287 repository and 77 compiler roots also agree, including ASan/UBSan.
Three rule mutants, three raw-question mutants and a listener metadata mutant are caught; new released questions refuse with panic 70.
Default options only; React claims remain parked, native is slower than Go, and setup cache warming failed with disk exhaustion.

## Implementation and validation

Rules receive cached parsed nodes through an owned numeric kind-indexed driver. Parser node lookup happens once in the driver; rules neither refetch parser nodes nor compare kind strings. rule.json and compiled exported listener kinds are compared against production Go listener maps. Shared harness and generator sources are untouched. The existing owned question dispatcher adds only registration and routing lines.

require-await implements contextual promise contracts, generic substitutions, callback signatures, heritage contracts and suggestion ranges. New raw checker questions live in generic_call_signature.go, type_index.go and heritage_types.go with matching Adamic decoders. They expose checker facts, with lint decisions in Adamic. Symbol provenance uses existing raw bridge support. valid-typeof implements the default option, including global undefined suggestions; requireStringLiterals:true is not implemented.

The independent Go oracle imports pinned production cohere rules. Of 227 candidate roots (67 hand controls and 160 extracted production fixture literals), Go rejects 10 syntactically invalid roots before comparison. All 217 remaining roots are compared; native refusal does not filter them. Corpus manifests are the frozen validation-coverage manifests, not a new sample. All diagnostics, ranges, fixes and ordered suggestions are serialized and compared byte for byte. Corpus populations yield zero findings, so controls provide the positive findings and suggestion checks. This evidence establishes tested parity, not exhaustive correctness on arbitrary programs.

Mutants each exit successfully with empty stderr and are caught by the byte comparison: require-await suggestion wording; Symbol declaration provenance inversion; valid-typeof undefined wording; resolved generic signature target removal; numeric index replacement with string index; and heritage root omission. A rule.json CallExpression-to-BinaryExpression mutation fails the production listener comparison. Three released raw-question calls require exact panic 70. Normal and sanitized control and corpus runs pass. Bridge tests and vet pass; the complete repository gate was not run.

Reproduction (test output goes directly to files):

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave-24-fifth/validate.py /workspace/wave24-fifth-check --stage0 /tmp/wave24-landing3-adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-fifth-check.log 2>&1
go test ./bridge/tsgo/... -count=1 > /tmp/wave24-fifth-bridge.log 2>&1
go vet ./bridge/tsgo/... > /tmp/wave24-fifth-vet.log 2>&1
```

Compiler can be rebuilt with go build -o /tmp/adamic ./cmd/adamic. Setup tools succeeded: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s. Cache warming exited 1 with no space left on device in /tmp; there are no successful cache-warm or done timing lines. nproc=5. Owned targeted gates above completed despite that failure. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

## Timing

Three alternating whole-process count runs, no build time: repository native median 285.523 ms versus Go 89.694 ms (3.183x slower); compiler native 1476.444 ms versus Go 303.951 ms (4.858x slower). Samples are in evidence/timings.json. Driver numeric kind lookup uses a map built once per file. These zero-finding corpora include load/parse/walk costs and demonstrate no native speed advantage. Earlier concurrent measurements are not used for the final timing claim.

Evidence contains complete run logs, canonical outputs compressed with gzip, setup failure, bridge/vet logs, listener comparisons, mutants and source-selection audit. No binaries or generated C archives are committed. React native HIR/SSA/capture work remains parked pending #dnv6f2c and JSX integration; require-atomic-updates also needs capture-sensitive control-flow graph support. No further eligible non-React rule remains in the last audit after these three claims.
