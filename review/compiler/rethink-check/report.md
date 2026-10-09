Built: evidence-only repros for six old compiler tasks and all nine review-lane witnesses; no compiler fixes or fixtures.
Commits: base 50654a40; delivery is the commit containing this report on compiler/rethink-check.
Commands and outputs: Node source, native release, JavaScript backend and ASan/UBSan/LSan runs; 2 FIXED, 5 REFUSED, 8 STILL WRONG across 15 programs.
Mutants: three executed JavaScript artifact mutants independently caught by stdout, stderr and exit comparisons; no compiler implementation mutants.
Not covered: full gate, generic static specializations, accessor-spread throw edges while refused, or new fixtures/count rows.

The explicitly requested base takes precedence over the generic current-main start rule. No later main is merged into this audit. Programs are committed only as .a.txt and copied to /tmp/rethink-check/*.a for execution. No cohere source was copied. This supplies prelanding evidence for the lowering chain and roadmap steps 16 (generics) and 22 (accessor spread); it lands no implementation.

Classification uses exact stdout, stderr and exit agreement with source Node, plus a clean sanitized run. REFUSED includes a sound path-bearing NotYet implementation stop, not just a permanent language refusal. A workaround below is reviewer guidance, not a fix supplied by a diagnostic unless explicitly stated. STILL WRONG includes bad C, crashes and exact output differences. In particular, an intentional engine-stderr difference is labeled and not inferred to be a new compiler bug.

| Task | Classification | Evidence |
| --- | --- | --- |
| #bxpmash | FIXED | exit 0; stdout "ok\n"; stderr "" |
| #drtmvb3 | FIXED | exit 0; stdout "false true\n"; stderr "" |
| #vk7ed2m | REFUSED | exit 1; stdout ""; stderr "adamic: vk7ed2m.a: stage 0 can't lower spreading in a program with static constructor objects yet\n" |
| #q9hz1bf | REFUSED | exit 1; stdout ""; stderr "adamic: q9hz1bf.a: stage 0 can't lower spreading an accessor literal whose getter may throw yet\n" |
| #dv99xzy | REFUSED | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/dv99xzy.a:2:8: stage 0 can't lower a function value that captures the variable its own initializer declares yet\n" |
| #eep1m5z | STILL WRONG | 0 FIXED; 2 REFUSED; 7 STILL WRONG |

| Program | Classification | Node | Native release | JavaScript | Sanitized native |
| --- | --- | --- | --- | --- | --- |
| [bxpmash](bxpmash.a.txt) | FIXED | exit 0; stdout "ok\n"; stderr "" | exit 0; stdout "ok\n"; stderr "" | exit 0; stdout "ok\n"; stderr "" | exit 0; stdout "ok\n"; stderr "" |
| [drtmvb3](drtmvb3.a.txt) | FIXED | exit 0; stdout "false true\n"; stderr "" | exit 0; stdout "false true\n"; stderr "" | exit 0; stdout "false true\n"; stderr "" | exit 0; stdout "false true\n"; stderr "" |
| [dv99xzy](dv99xzy.a.txt) | REFUSED | exit 0; stdout "0\n"; stderr "" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/dv99xzy.a:2:8: stage 0 can't lower a function value that captures the variable its own initializer declares yet\n" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/dv99xzy.a:2:8: stage 0 can't lower a function value that captures the variable its own initializer declares yet\n" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/dv99xzy.a:2:8: stage 0 can't lower a function value that captures the variable its own initializer declares yet\n" |
| [eep1m5z-93122fa_t_t5](eep1m5z-93122fa_t_t5.a.txt) | STILL WRONG | exit 0; stdout "1 2\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" | exit 0; stdout "1 2\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" |
| [eep1m5z-93122fa_t_v1](eep1m5z-93122fa_t_v1.a.txt) | STILL WRONG | exit 0; stdout "2 -1\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" | exit 0; stdout "2 -1\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" |
| [eep1m5z-93122fa_t_v2](eep1m5z-93122fa_t_v2.a.txt) | STILL WRONG | exit 0; stdout "2 3\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" | exit 0; stdout "2 3\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" |
| [eep1m5z-93122fa_t_v5](eep1m5z-93122fa_t_v5.a.txt) | STILL WRONG | exit 0; stdout "2\n2\n"; stderr "" | exit 70; stdout "2\n"; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" | exit 0; stdout "2\n2\n"; stderr "" | exit 70; stdout "2\n"; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" |
| [eep1m5z-9984394_uncaught_name](eep1m5z-9984394_uncaught_name.a.txt) | STILL WRONG | exit 1; stdout ""; stderr "file:///tmp/rethink-check/eep1m5z-9984394_uncaught_name.a:2\n    throw new Error(\`message ${1}\`);\n          ^\n\nmessage 1\n    at file:///tmp/rethink-check/eep1m5z-9984394_uncaught_name.a:2:11\n    at ModuleJob.run (node:internal/modules/esm/module_job:439:25)\n    at async node:internal/modules/esm/loader:643:26\n    at async file:///workspace/adamic/oracle/node.mjs:82:1\n\nNode.js v24.19.0\n" | exit 1; stdout ""; stderr "" | exit 1; stdout ""; stderr "" | exit 1; stdout ""; stderr "" |
| [eep1m5z-classfeat_maybe_setter](eep1m5z-classfeat_maybe_setter.a.txt) | REFUSED | exit 0; stdout "3\n"; stderr "" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/eep1m5z-classfeat_maybe_setter.a:4:5: stage 0 can't lower a setter with an optional numeric input; use a plain optional numeric field until accessor thunks support two-word arguments yet\n" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/eep1m5z-classfeat_maybe_setter.a:4:5: stage 0 can't lower a setter with an optional numeric input; use a plain optional numeric field until accessor thunks support two-word arguments yet\n" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/eep1m5z-classfeat_maybe_setter.a:4:5: stage 0 can't lower a setter with an optional numeric input; use a plain optional numeric field until accessor thunks support two-word arguments yet\n" |
| [eep1m5z-classfeat_method_view](eep1m5z-classfeat_method_view.a.txt) | STILL WRONG | exit 0; stdout "worker one1\nlazy\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" | exit 1; stdout ""; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" |
| [eep1m5z-classfeat_static_virtual](eep1m5z-classfeat_static_virtual.a.txt) | STILL WRONG | exit 1; stdout ""; stderr "file:///tmp/rethink-check/eep1m5z-classfeat_static_virtual.a:13\n        return this.label.toUpperCase();\n                          ^\n\nTypeError: Cannot read properties of undefined (reading 'toUpperCase')\n    at Derived.n (file:///tmp/rethink-check/eep1m5z-classfeat_static_virtual.a:13:27)\n    at Derived.m (file:///tmp/rethink-check/eep1m5z-classfeat_static_virtual.a:6:26)\n    at <static_initializer> (file:///tmp/rethink-check/eep1m5z-classfeat_static_virtual.a:10:25)\n    at file:///tmp/rethink-check/eep1m5z-classfeat_static_virtual.a:9:23\n    at ModuleJob.run (node:internal/modules/esm/module_job:439:25)\n    at async node:internal/modules/esm/loader:643:26\n    at async file:///workspace/adamic/oracle/node.mjs:82:1\n\nNode.js v24.19.0\n" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" | exit 1; stdout ""; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" |
| [eep1m5z-iterators_derived_symbol](eep1m5z-iterators_derived_symbol.a.txt) | REFUSED | exit 0; stdout "1,2\n"; stderr "" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/eep1m5z-iterators_derived_symbol.a:5:5: stage 0 can't lower a computed member name in a derived class; use a named member until override lookup supports computed names yet\n" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/eep1m5z-iterators_derived_symbol.a:5:5: stage 0 can't lower a computed member name in a derived class; use a named member until override lookup supports computed names yet\n" | exit 1; stdout ""; stderr "adamic: /tmp/rethink-check/eep1m5z-iterators_derived_symbol.a:5:5: stage 0 can't lower a computed member name in a derived class; use a named member until override lookup supports computed names yet\n" |
| [eep1m5z](eep1m5z.a.txt) | STILL WRONG | exit 0; stdout "1 2\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" | exit 0; stdout "1 2\n"; stderr "" | exit 70; stdout ""; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n" |
| [q9hz1bf](q9hz1bf.a.txt) | REFUSED | exit 0; stdout "getter\n"; stderr "" | exit 1; stdout ""; stderr "adamic: q9hz1bf.a: stage 0 can't lower spreading an accessor literal whose getter may throw yet\n" | exit 1; stdout ""; stderr "adamic: q9hz1bf.a: stage 0 can't lower spreading an accessor literal whose getter may throw yet\n" | exit 1; stdout ""; stderr "adamic: q9hz1bf.a: stage 0 can't lower spreading an accessor literal whose getter may throw yet\n" |
| [vk7ed2m](vk7ed2m.a.txt) | REFUSED | exit 0; stdout "1\n"; stderr "" | exit 1; stdout ""; stderr "adamic: vk7ed2m.a: stage 0 can't lower spreading in a program with static constructor objects yet\n" | exit 1; stdout ""; stderr "adamic: vk7ed2m.a: stage 0 can't lower spreading in a program with static constructor objects yet\n" | exit 1; stdout ""; stderr "adamic: vk7ed2m.a: stage 0 can't lower spreading in a program with static constructor objects yet\n" |

Observations and attribution:

- #bxpmash: Main commit 13213182: git log -S found classType = l.concrete(classType), before inner layout selection. Historical static generic specialization requirements are outside this minimal probe.
- #drtmvb3: Candidate attribution: main commit 4556d540 added the Optional exclusion in the shared checked-view reader (git log -S). No claim that this search proves it is the first fixing commit.
- #vk7ed2m: Safe NotYet stop. Workaround: construct { value: box.value } directly rather than spreading it. The diagnostic itself does not supply a fix.
- #q9hz1bf: Safe NotYet stop. Workaround: read source.item explicitly inside the try, then construct a data record. No accessor spread exception edge was exercised; no checkThrown mutant is claimed.
- #dv99xzy: Safe NotYet stop. Workaround: declare function visit(depth: number): number inside run instead of a self-capturing arrow initializer. The diagnostic itself does not supply a fix.
- #eep1m5z: Primary minimal tuple-union length probe. The nine original review witnesses follow below; the brief truncation was resolved using the repository evidence.
- eep1m5z-93122fa_t_t5: Use an array representation for dynamic lengths or discriminate the tuple variants before their individual fixed-length reads.
- eep1m5z-93122fa_t_v1: Test undefined explicitly before reading the known tuple length.
- eep1m5z-93122fa_t_v2: Use an array representation for dynamic lengths or discriminate the tuple variants.
- eep1m5z-93122fa_t_v5: Use an array representation for dynamic length or test the possibly missing tuple explicitly.
- eep1m5z-9984394_uncaught_name: Strict stderr comparison differs intentionally: current uncaught language exceptions exit 1, and native omits engine stack rendering (runtime/exceptions.c). This does not reproduce the former colon formatting bug.
- eep1m5z-classfeat_maybe_setter: Use explicit getLevel/setLevel methods, or a non-optional numeric setter.
- eep1m5z-classfeat_method_view: Use an ordinary run method, or a callable property interface with an arrow field.
- eep1m5z-classfeat_static_virtual: Initialize Derived.label before Derived.early evaluates this.m().
- eep1m5z-iterators_derived_symbol: Use a non-derived iterator class with its own limit field.

Complete per-command outputs, invocation arguments and elapsed seconds are in [results.json](results.json). Every compiler invocation and execution has separate .out.log and .err.log files. Generated JavaScript is retained in the js-compile.out.log files. The program links above contain the exact inputs. [run.py](run.py) recreates the suite with a 60-second limit per subprocess and progress saved after every program. [run-comparison-mutants.py](run-comparison-mutants.py) executes the three mutants, saving their complete artifacts as .mjs.txt.

Toolchain setup:

GOPROXY=https://proxy.golang.org|direct was set before setup; /workspace/adamic-tools/env.sh was sourced for builds and runs. nproc=5. The initial timeout 600 bash cloud/setup.sh and retry with GOFLAGS=-p=2 under timeout 300 did not complete dependency warming. The retry exited 124. The initial surrounding shell continued to nproc, so it did not preserve the setup exit code. Go, clang with its sanitizer overflow check, Node and the submodule finished installation. Initial CLI build limits 180 and 300 also expired (124). The final bounded build used timeout 600 go build -p 2 -o /tmp/rethink-adamic ./cmd/adamic; succeeded with exit 0; see build-final.log. The installed toolchain is Go 1.27.1, clang 20.1.8, Node 24.19.0.

Initial setup timing lines:

```text
setup: go ready (0.342s)
setup: node ready (0.398s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (6.791s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=10.998s
setup: markdown dependencies ready (11.882s)
setup: submodules ready (72.338s)
```

Retry setup timing lines:

```text
setup: node ready (0.954s)
setup: go ready (1.881s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.293s
setup: markdown dependencies ready (3.641s)
setup: submodules ready (3.856s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (7.805s)
```

Validation commands:

```sh
source /workspace/adamic-tools/env.sh
timeout 600 python3 review/compiler/rethink-check/run.py > review/compiler/rethink-check/run.log 2>&1
timeout 60 python3 review/compiler/rethink-check/run-comparison-mutants.py > review/compiler/rethink-check/comparison-mutants.log 2>&1
git diff --check
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

No Go tests or .a fixtures were added or touched, so counts.md regeneration and TestCallTargetReaders do not apply. No whole package tests or full gate were run. Integration lane output is saved in lane-checks.log. An automatic approval-review attempt for the evidence runner timed out; its authorized retry succeeded.
