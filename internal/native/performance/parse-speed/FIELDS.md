# Read-side cache boundary and final split, October 7, 2026

Built: runtime-only method/function-field cache fast path, after the release and UTF-16 input cache fixes.
Commits: follows 3196034 and 86677c1 on codex/parse-speed; object.c read path only, no write/emission/parser/scanner changes.
Commands and outputs: Node interface control and 77-file batch8/AST parity PASS; 6,504,806,296 to 6,460,129,543 Ir.
Mutant: busy routed to count in the miss helper fails Node stdout, both exit 0 with empty stderr; production source restored.
Not covered: full repository gate, implemented per-file region, or either dated parse target.

## Exact release configuration

Every native Ir number and estimate below uses clang 20.1.8, `-O2 -g`,
sanitizers off, `-ffp-contract=off`, `-fno-optimize-sibling-calls`, no LTO,
no counted-runtime hooks. C is the unchanged runtime-parse.c generated after
the runtime-area merge. The scratch prototype's complete .c/.h runtime is
byte-identical to production after promotion. Exact measured after command:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I /workspace/scratch/parse-speed/callee-fast-runtime /workspace/scratch/parse-speed/runtime-parse.c /workspace/scratch/parse-speed/callee-fast-runtime/*.c -lm -o /workspace/scratch/parse-speed/callee-fast-parse > /tmp/parse-speed-callee-fast-clang.log 2>&1
VALGRIND_LIB=/workspace/scratch/parse-speed/valgrind/usr/libexec/valgrind /workspace/scratch/parse-speed/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/parse-speed/callee-fast.callgrind /workspace/scratch/parse-speed/callee-fast-parse --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/parse-speed/callee-fast.stdout 2> /workspace/scratch/parse-speed/callee-fast.stderr
```

## What remains of field plumbing

After the runtime-area merge, the source-aware bucket is **194,088,088 Ir**:
**113,527,190 Ir** for field/method reads, caches and lookups plus **80,560,898
Ir** for inline frozen-write guards owned by the compiler unit. The write guard,
SetProperty emission and object_write_field/check_write implementation are not
changed here. The historical bucket was 1,317,090,232 Ir; both source-aware v2
columns below consistently classify inline runtime work and input decoding.

Read lookup costs before this fix, same release flags: object_callee has
**2,628,517 calls**, **68,373,628 self Ir**, **68,437,982 inclusive Ir**.
Only **335 calls** miss the shape cache (0.01275%); the miss body performs 2,869
strcmp calls, 64,354 libc Ir. Ordinary object_find has just **93 calls and
4,942 self Ir**. No accessor_find/get or static-field lookup instructions occur
in this profile. Name search is not the large remaining cost. Nearly every
method/function-field read was saving/restoring registers needed only on misses.

The new noinline callee_cache_miss retains the original lookup semantics: own
function fields first, then shape methods, with cache invalidation on a shape
change, and the same missing-method panic. The ordinary path directly reads
cache/slot or method code. It never evaluates arguments, retains/releases a
receiver, or changes when a method is read; those remain at the original caller.
No generic lookup is removed or statically assumed safe. No emission hook is
needed; emit_*.go and the compiler worker's write path remain untouched by us.

After: object_callee self **23,655,983 Ir**, inclusive **23,761,233 Ir**;
callee_cache_miss self **40,896 Ir**, inclusive **105,250 Ir**, 335 calls.
The successful function-field path takes nine instructions, method path thirteen,
including return, with no saved-register prologue on a hit. Whole parse saves
**44,676,753 Ir (0.68683%)**, from 6,504,806,296 to 6,460,129,543. Outlining only
the name search (callee_index prototype) saved just 7,884,891 Ir and kept register
saves on the hit; it was not promoted. Moving the complete miss path avoids that.

## Final whole-process and disjoint results

| Same pinned 77 files, exact release flags above | Whole-process Ir |
| --- | ---: |
| Accepted reconstruction baseline | 9,259,261,285 |
| Runtime-area merge, 6fc42a9 tip | 7,389,521,274 |
| Last-reference release boundary | 6,570,598,637 |
| Input UTF-16 cache propagation | 6,504,806,296 |
| Read-side cache-miss boundary | 6,460,129,543 |

Total reduction from the accepted baseline is **30.2306%**, including the
runtime-area improvements. Our three isolated runtime fixes remove another
**929,391,731 Ir (12.5772%)** after that merge. No uncontended wall-time or
cross-machine speedup is inferred. Go parse alone remains **1,684,637,174 Ir**,
Go 1.27.1 ordinary optimized build, sanitizers/race off,
GOMAXPROCS=1 GODEBUG=asyncpreemptoff=1 under Callgrind. Native is **3.8347x**;
neither 2.53G by October 8 noon MDT nor 1.685G by October 9 noon MDT is met.

Source-aware v2 rules from resplit.py assign every self instruction once, with
known inlined header costs under the same mechanism and ambiguous/unattributed
work left in the remainder. They correct v1's input.c decode attribution;
String equality includes kind and other bodies in this table, while the original
separate kind probe remains in SPLIT.md. Inclusive results are not summed into
this table. Common exact release flags apply to every column.

| Self bucket | Baseline v2 | Runtime merge v2 | Final v2 | Saved from baseline |
| --- | ---: | ---: | ---: | ---: |
| Releases and child destruction | 1,562,416,966 | 1,534,667,356 | 715,744,719 | 846,672,247 |
| Object field and call plumbing, including inline | 1,346,142,886 | 194,088,088 | 149,411,339 | 1,196,731,547 |
| Retains | 363,315,534 | 354,990,651 | 354,990,651 | 8,324,883 |
| String equality bodies, kind and other | 335,614,089 | 339,705,054 | 339,705,054 | -4,090,965 |
| Character reads and UTF-16 indexing, including inline | 1,240,515,174 | 960,751,669 | 876,147,377 | 364,367,797 |
| Remainder: parser, arrays, driver, libc, startup and unattributed inline | 1,544,732,689 | 1,446,962,638 | 1,446,962,494 | 97,770,195 |
| Scanner generated control | 1,834,614,732 | 1,585,314,417 | 1,585,314,705 | 249,300,027 |
| Line table generated control | 83,445,905 | 83,445,905 | 83,445,905 | 0 |
| File reading and UTF-8 input decode self | 225,789,370 | 225,789,370 | 244,600,939 | -18,811,569 |
| Allocation and freeing | 211,140,210 | 211,140,210 | 211,140,210 | 0 |
| Other string operations, including token values | 96,126,460 | 95,031,911 | 95,031,911 | 1,094,549 |
| Node construction generated control | 286,305,012 | 220,508,211 | 220,508,211 | 65,796,801 |
| Parent map generated control | 50,564,366 | 55,891,184 | 55,891,184 | -5,326,818 |
| Substrings | 78,537,892 | 81,234,610 | 81,234,844 | -2,696,952 |
| Total | 9,259,261,285 | 7,389,521,274 | 6,460,129,543 | 2,799,131,742 |

## Mutant, parity and leak checks

The mutant routes a method name busy to count only in the newly outlined miss
helper. Both are zero-argument methods in the existing class_as_interface fixture.
Both programs exit 0 with empty stderr; Node prints `2 true` and native `2 false`
in the parameter-handler case. The intended stdout oracle catches it, not a
warning, refusal or sanitizer fault. The helper is restored in finally before
all passing controls and gates. Node's correct busy returns seen.length > 1.
This fixture also checks polymorphic shape changes, own function fields, class
method thunks, receiver/function replacement during argument evaluation and throws.

The good interface control passes. Full batch8 release output is **11,442,907
identical bytes** against typescript-go on all 77 files; full AST is **44,766,682
identical bytes** against Go/Node, all 77 files, including ASan/UBSan native.
The complete uncached native/oracle package checks include LeakSanitizer;
logs accompany this report. Sanitized correctness builds use `-O1 -g
-fsanitize=address,undefined -fno-sanitize-recover=all`, with the same strict
C/warning/contraction/sibling flags, and are not instruction measurements.
Repository formatting and vet are also logged. The full repository test gate
was not run. No allocation-count fixture rows change.

## Remaining region question

DESTRUCTION.md records the original diagnosis: all 889,146 ParseNodes die in
caller cleanup after each file, while 41.06M external releases include far more
non-final/immortal traffic. The observed 354.1M inclusive cleanup boundary is
about 23% of the original 1.54G release bucket. A structural per-file region
could remove most of that boundary; bulk teardown alone did not collapse most
of the original bucket. The 0.55 to 0.65G original estimate additionally assumes
removing generated node/numeric-array counts, not just changing the allocator.

After our fast release, structural teardown is a larger fraction of the remaining
715.7M release/child bucket. A rough post-fix model is 0.30 to 0.35G of teardown
plus about 77M of non-final node/numeric-array calls (8.60M at nine Ir each),
if a compiler proof elides them. That could exceed half of this smaller bucket,
but it is still an estimate with no implemented region or memory-peak validation.
Immortal region values that still call release take five instructions; setting
counts to zero alone is not call elision. Growing numeric-array buffers and
external/escaping strings must be accounted for before claiming one-stroke frees.
A follow-up region unit needs that ownership boundary and its small emission
hooks cleared with the compiler owner. No such hooks are made or pushed here.

Final checks: uncached internal/native PASS 139.273s; internal/oracle PASS
130.512s; gofmt and go vet ./... exit 0 with empty logs. On October 7 the user
reassigned subsequent character/index work to codex/string-views. The already
pushed input-cache fix is 86677c1; no further string changes are made here.
The parse will be re-measured when that worker lands in the runtime area.
