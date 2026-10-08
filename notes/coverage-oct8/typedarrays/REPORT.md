Built nine .a probes: seven agree in four modes; two are explicit compiler refusals.
Probe commit 9197ad5b4f0716bb2f0a6fd924ae95d594fe646c; tested base 031a1259bc7973934792dc6cb1bd4074fc2204b9; history examined: 73352e87..5a2681b1.
Commands: run.py compares source Node, ASan/UBSan native, -O2 native and backend Node; seven leak runs finish cleanly.
Mutants: negative-wrap and same-kind guards caught; regexp guard survives selected native tests but oracle catches it; readiness call invalidation survives lower; extra view retain caught by leak check.
Not covered: every readiness path, unsupported element types, allocation exhaustion, resizable/shared buffers, or the full repository gate.

Observations

The requested branch was created with `git fetch origin && git checkout -b codex/coverage-oct8-typedarrays origin/main`. The named five source files and their requested `git log -p` history were read. CLAUDE.md, README.md and the language/memory documentation were consulted. No compiler changes are committed. These probes are notes, not new registered oracle fixtures; internal/oracle/counts.md is unchanged.

`source /workspace/adamic-tools/env.sh` was used for builds and tests. Setup succeeded; its complete output is in setup.log. It reported Go ready 0.131s, Node ready 0.148s, clang ready 0.517s, markdown dependencies ready 1.763s, submodules ready 25.052s, Go build ready 366.299s, cache warm 366.642s and done 366.813s. `nproc` is 5; cgroup CPU quota is 4. Go 1.27.1, clang 20.1.8, Node 24.19.0. The first package run lacked the pinned @types/node 25.3.3 dependency; `npm ci --prefix stage3/api` installed the lockfile dependencies successfully, then lowering was rerun. The first attempt to run probes preceded compiler build completion and failed with FileNotFoundError; the completed final run is recorded here.

Reproduction: build `go build -o /tmp/coverage-adamic ./cmd/adamic`, then `python3 notes/coverage-oct8/typedarrays/run.py > /tmp/coverage-runs.log 2>&1`. The runner saves each process stream in /tmp/coverage-oct8-typedarrays and records commands, streams and exit codes in results.json. Builds use the compiler's own --sanitize option (oracle flags: -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all) and default -O2. Sanitizer comparison uses detect_leaks=0, as the oracle does; successful binaries are run separately with detect_leaks=1. Each subprocess has a 60-second deadline. Source runs use oracle/node.mjs to strip types from the source, independently of Adamic lowering. Compilation failures are recorded as compilation outcomes, not invented native executions.

Coverage

| Program | Cases | Four modes | Leak check |
|---|---|---|---|
| buffer-gap.a | Offset/length view over ArrayBuffer, deliberately unsupported | compiler refusal; source exit 0 | not applicable |
| float64array.a | Float64 preservation, signed zero reciprocal, NaN/infinities and subnormals | agree, exit 0 | clean |
| int32array.a | Int32 signed wrap, NaN/infinities, huge values, subnormals, constructor vs assignment | agree, exit 0 | clean |
| readiness-callback.a | Historical uninitialized-slot syntax and callback rebinding, deliberately refused on current main | compiler refusal; source exit 0 | not applicable |
| readiness-fields.a | Constructor branch initialization; narrowed interface reads; captured reads; false overload result to boolean-or-undefined; rest with default and explicit undefined | agree, exit 0 | clean |
| readiness.a | Branches, loop assignment, caught throw/finally, field alias, direct-call field write, supported object spread | agree, exit 0 | clean |
| replacement.a | Named/unmatched captures; $$, $&, $1, $2, $<name>, prefix/suffix tokens; string and regexp patterns; function replacers; Unicode/UTF-16 empty matches; callback lastIndex writes; no match | agree, exit 0 | clean |
| rest-closure.a | Rest function value and class method, comment witness | agree, exit 0 | clean |
| uint8array.a | Uint8 modulo/truncation, NaN/infinities, huge values, subnormals, constructor vs assignment | agree, exit 0 | clean |

All three typed-array programs additionally cover nested nonzero offsets, lengths after subarray, left and right overlapping set between sibling/nested views, fractional offsets (including -0.9), fill over a view, empty views and empty set, and a returned nested view after its original bindings have gone out of scope. Existing typed_arrays_views.a already exercises basic shared views and overlap; these programs add the same nested offset and lifetime scenario for each width and much larger conversion magnitudes.

Current-main readiness differs from the historical landing: non-null assertions in .a are refused, and .ts assertions are eager. The historical assertion-based readiness witness therefore never reaches lowering. It cannot establish a readiness miscompile on this base. Accepted readiness controls were written without assertions. No accepted-program output disagreement, C compilation failure, runtime abort, or leak was observed in these nine probes.

The two compiler disagreements below are supported-gap observations, not silent miscompiles. `internal/lower/typed_arrays.go:52` returns NotYet for ArrayBuffer. `internal/lower/readiness.go:21` excludes checked TypeScript syntax from lazy initialization; the .a witness is refused by the non-null refusal in internal/lower/refusals.go:104 before readiness can create its historical slots.

Verbatim four-mode observations

Empty stderr is written as an empty fenced block. Native/compiler phases are explicitly distinguished. The programs themselves are adjacent .a files. For agreeing programs all four streams are identical, but each is included so the evidence is self-contained.

### buffer-gap.a

source: exit 0, phase execution.

stdout:

```text
8:255
```

stderr:

```text
```

sanitized: exit 1, phase compiler.

stdout:

```text
```

stderr:

```text
adamic: /workspace/adamic/notes/coverage-oct8/typedarrays/buffer-gap.a:1:16: stage 0 can't lower ArrayBuffer yet
```

release: exit 1, phase compiler.

stdout:

```text
```

stderr:

```text
adamic: /workspace/adamic/notes/coverage-oct8/typedarrays/buffer-gap.a:1:16: stage 0 can't lower ArrayBuffer yet
```

javascript: exit 1, phase compiler.

stdout:

```text
```

stderr:

```text
adamic: /workspace/adamic/notes/coverage-oct8/typedarrays/buffer-gap.a:1:16: stage 0 can't lower ArrayBuffer yet
```

### float64array.a

source: exit 0, phase execution.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

sanitized: exit 0, phase execution.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

release: exit 0, phase execution.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

javascript: exit 0, phase execution.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

### int32array.a

source: exit 0, phase execution.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:256:256:0.00390625
5:-256:-256:-0.00390625
6:2147483647:2147483647:4.656612875245797e-10
7:-2147483648:-2147483648:-4.656612873077393e-10
8:2147483647:2147483647:4.656612875245797e-10
9:1:1:1
10:-1:-1:-1
11:-1:-1:-1
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:256
right:0
right:0
right:256
right:-256
right:1
right:-1
right:-1
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:256
left:-256
left:1
left:-1
left:-256
left:1
left:-1
left:-1
left:0
left:0
left:0
fill:-257:-257:-1
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

sanitized: exit 0, phase execution.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:256:256:0.00390625
5:-256:-256:-0.00390625
6:2147483647:2147483647:4.656612875245797e-10
7:-2147483648:-2147483648:-4.656612873077393e-10
8:2147483647:2147483647:4.656612875245797e-10
9:1:1:1
10:-1:-1:-1
11:-1:-1:-1
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:256
right:0
right:0
right:256
right:-256
right:1
right:-1
right:-1
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:256
left:-256
left:1
left:-1
left:-256
left:1
left:-1
left:-1
left:0
left:0
left:0
fill:-257:-257:-1
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

release: exit 0, phase execution.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:256:256:0.00390625
5:-256:-256:-0.00390625
6:2147483647:2147483647:4.656612875245797e-10
7:-2147483648:-2147483648:-4.656612873077393e-10
8:2147483647:2147483647:4.656612875245797e-10
9:1:1:1
10:-1:-1:-1
11:-1:-1:-1
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:256
right:0
right:0
right:256
right:-256
right:1
right:-1
right:-1
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:256
left:-256
left:1
left:-1
left:-256
left:1
left:-1
left:-1
left:0
left:0
left:0
fill:-257:-257:-1
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

javascript: exit 0, phase execution.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:256:256:0.00390625
5:-256:-256:-0.00390625
6:2147483647:2147483647:4.656612875245797e-10
7:-2147483648:-2147483648:-4.656612873077393e-10
8:2147483647:2147483647:4.656612875245797e-10
9:1:1:1
10:-1:-1:-1
11:-1:-1:-1
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:256
right:0
right:0
right:256
right:-256
right:1
right:-1
right:-1
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:256
left:-256
left:1
left:-1
left:-256
left:1
left:-1
left:-1
left:0
left:0
left:0
fill:-257:-257:-1
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

### readiness-callback.a

source: exit 0, phase execution.

stdout:

```text
replaced
undefined
```

stderr:

```text
```

sanitized: exit 1, phase compiler.

stdout:

```text
```

stderr:

```text
adamic: /workspace/adamic/notes/coverage-oct8/typedarrays/readiness-callback.a:1:36: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case
```

release: exit 1, phase compiler.

stdout:

```text
```

stderr:

```text
adamic: /workspace/adamic/notes/coverage-oct8/typedarrays/readiness-callback.a:1:36: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case
```

javascript: exit 1, phase compiler.

stdout:

```text
```

stderr:

```text
adamic: /workspace/adamic/notes/coverage-oct8/typedarrays/readiness-callback.a:1:36: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case
```

### readiness-fields.a

source: exit 0, phase execution.

stdout:

```text
left:right
after
false:false:true
default:0:-1
default:2:4
```

stderr:

```text
```

sanitized: exit 0, phase execution.

stdout:

```text
left:right
after
false:false:true
default:0:-1
default:2:4
```

stderr:

```text
```

release: exit 0, phase execution.

stdout:

```text
left:right
after
false:false:true
default:0:-1
default:2:4
```

stderr:

```text
```

javascript: exit 0, phase execution.

stdout:

```text
left:right
after
false:false:true
default:0:-1
default:2:4
```

stderr:

```text
```

### readiness.a

source: exit 0, phase execution.

stdout:

```text
yes01
no01
beforeboom!
ready
changed
changedcopy
```

stderr:

```text
```

sanitized: exit 0, phase execution.

stdout:

```text
yes01
no01
beforeboom!
ready
changed
changedcopy
```

stderr:

```text
```

release: exit 0, phase execution.

stdout:

```text
yes01
no01
beforeboom!
ready
changed
changedcopy
```

stderr:

```text
```

javascript: exit 0, phase execution.

stdout:

```text
yes01
no01
beforeboom!
ready
changed
changedcopy
```

stderr:

```text
```

### replacement.a

source: exit 0, phase execution.

stdout:

```text
😀[$][ab][a][b][a][][😀][-b-ab!]-[$][b][][b][][][😀ab-][-ab!]-[$][ab][a][b][a][][😀ab-b-][!]!
😀[$][ab][a][b][a][][😀][-b-ab!]-[$][b][][b][][][😀ab-][-ab!]-[$][ab][a][b][a][][😀ab-b-][!]!
😀[$][ab][$1][$2][$<letter>][$<missing>][😀][-b-ab!]-b-[$][ab][$1][$2][$<letter>][$<missing>][😀ab-b-][!]!
😀[$][ab][$1][$2][$<letter>][$<missing>][😀][-b-ab!]-b-ab!
$:😀$:
$:�$:�$:
0:a:-:ab:$$:$&1:-:b:ab:$$:$&
0😀2a3b4
a:0:ab
1:402:413:42
3:42
miss
```

stderr:

```text
```

sanitized: exit 0, phase execution.

stdout:

```text
😀[$][ab][a][b][a][][😀][-b-ab!]-[$][b][][b][][][😀ab-][-ab!]-[$][ab][a][b][a][][😀ab-b-][!]!
😀[$][ab][a][b][a][][😀][-b-ab!]-[$][b][][b][][][😀ab-][-ab!]-[$][ab][a][b][a][][😀ab-b-][!]!
😀[$][ab][$1][$2][$<letter>][$<missing>][😀][-b-ab!]-b-[$][ab][$1][$2][$<letter>][$<missing>][😀ab-b-][!]!
😀[$][ab][$1][$2][$<letter>][$<missing>][😀][-b-ab!]-b-ab!
$:😀$:
$:�$:�$:
0:a:-:ab:$$:$&1:-:b:ab:$$:$&
0😀2a3b4
a:0:ab
1:402:413:42
3:42
miss
```

stderr:

```text
```

release: exit 0, phase execution.

stdout:

```text
😀[$][ab][a][b][a][][😀][-b-ab!]-[$][b][][b][][][😀ab-][-ab!]-[$][ab][a][b][a][][😀ab-b-][!]!
😀[$][ab][a][b][a][][😀][-b-ab!]-[$][b][][b][][][😀ab-][-ab!]-[$][ab][a][b][a][][😀ab-b-][!]!
😀[$][ab][$1][$2][$<letter>][$<missing>][😀][-b-ab!]-b-[$][ab][$1][$2][$<letter>][$<missing>][😀ab-b-][!]!
😀[$][ab][$1][$2][$<letter>][$<missing>][😀][-b-ab!]-b-ab!
$:😀$:
$:�$:�$:
0:a:-:ab:$$:$&1:-:b:ab:$$:$&
0😀2a3b4
a:0:ab
1:402:413:42
3:42
miss
```

stderr:

```text
```

javascript: exit 0, phase execution.

stdout:

```text
😀[$][ab][a][b][a][][😀][-b-ab!]-[$][b][][b][][][😀ab-][-ab!]-[$][ab][a][b][a][][😀ab-b-][!]!
😀[$][ab][a][b][a][][😀][-b-ab!]-[$][b][][b][][][😀ab-][-ab!]-[$][ab][a][b][a][][😀ab-b-][!]!
😀[$][ab][$1][$2][$<letter>][$<missing>][😀][-b-ab!]-b-[$][ab][$1][$2][$<letter>][$<missing>][😀ab-b-][!]!
😀[$][ab][$1][$2][$<letter>][$<missing>][😀][-b-ab!]-b-ab!
$:😀$:
$:�$:�$:
0:a:-:ab:$$:$&1:-:b:ab:$$:$&
0😀2a3b4
a:0:ab
1:402:413:42
3:42
miss
```

stderr:

```text
```

### rest-closure.a

source: exit 0, phase execution.

stdout:

```text
3:2
```

stderr:

```text
```

sanitized: exit 0, phase execution.

stdout:

```text
3:2
```

stderr:

```text
```

release: exit 0, phase execution.

stdout:

```text
3:2
```

stderr:

```text
```

javascript: exit 0, phase execution.

stdout:

```text
3:2
```

stderr:

```text
```

### uint8array.a

source: exit 0, phase execution.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:0:0:Infinity
5:0:0:Infinity
6:255:255:0.00392156862745098
7:0:0:Infinity
8:255:255:0.00392156862745098
9:1:1:1
10:255:255:0.00392156862745098
11:255:255:0.00392156862745098
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:1
right:255
right:255
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:0
left:0
left:1
left:255
left:0
left:1
left:255
left:255
left:0
left:0
left:0
fill:255:255:255
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

sanitized: exit 0, phase execution.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:0:0:Infinity
5:0:0:Infinity
6:255:255:0.00392156862745098
7:0:0:Infinity
8:255:255:0.00392156862745098
9:1:1:1
10:255:255:0.00392156862745098
11:255:255:0.00392156862745098
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:1
right:255
right:255
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:0
left:0
left:1
left:255
left:0
left:1
left:255
left:255
left:0
left:0
left:0
fill:255:255:255
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

release: exit 0, phase execution.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:0:0:Infinity
5:0:0:Infinity
6:255:255:0.00392156862745098
7:0:0:Infinity
8:255:255:0.00392156862745098
9:1:1:1
10:255:255:0.00392156862745098
11:255:255:0.00392156862745098
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:1
right:255
right:255
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:0
left:0
left:1
left:255
left:0
left:1
left:255
left:255
left:0
left:0
left:0
fill:255:255:255
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

javascript: exit 0, phase execution.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:0:0:Infinity
5:0:0:Infinity
6:255:255:0.00392156862745098
7:0:0:Infinity
8:255:255:0.00392156862745098
9:1:1:1
10:255:255:0.00392156862745098
11:255:255:0.00392156862745098
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:1
right:255
right:255
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:0
left:0
left:1
left:255
left:0
left:1
left:255
left:255
left:0
left:0
left:0
fill:255:255:255
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

Comment and documentation audit

Every comment in the five named files was compared to its adjacent implementation. This is a source audit, not a claim that every specification sentence was experimentally proven.

| Location | Observation |
|---|---|
| typed_arrays.go:30,216,290-291 | Unsupported library facilities are rejected while identity is available; optional offsets are fitted to MaybeNumber; set accepts only same-kind storage, not an object shape. Comments match these paths. |
| typed_array.c:1,34,44,67,136 | Counted views retain an owner; empty allocation has a non-NULL byte base; -0.5 truncates to -0; integer stores use finite/trunc/fmod and explicit signed wrapping; same-kind set uses memmove. Comments match. The probes exercise numeric conversion and views; allocation failure is not tested. |
| regexp_replace.c:1-2,9-10 | Intrinsic callback replacement first collects matches, constructs actual-kind arguments and validates/converts them to closure representations. The ECMA section reference describes this intrinsic path, not arbitrary overridden RegExp execution. No mismatch found in these comments. |
| readiness.go:18-19,380,385,399 | Initializer syntax recognition, eager .ts checks, and lazy operand storage match the functions. Historical lazy .a syntax is now refused before this machinery; this limits its reachable coverage. |
| readiness.go:52-54 | The monotonic statement concerns binding readiness between declarations. Field facts are separately invalidated by calls at lines 192 and 214. Read in that scope it matches; it would be misleading if taken to claim fields survive calls. |
| readiness.go:56-57,124-125,152,249,344 | Record field names keep representation checks; local metadata is cloned for CFG construction; must analysis starts at top and intersects predecessor states; reflection transforms this instruction without recursively transforming nested statements with graph locations; field facts key binding and name. Comments match the code. |
| census_small.go:10-11 | **Comment mismatch:** function values and methods do not remain NotYet on this base. rest-closure.a prints 3:2 in all four modes. The census rest helpers have no callers found by rg; functions.go:138-149 now handles rest parameters. |
| census_small.go:34,53-54,97 | The helper implementations recover a rest declaration, build a fresh packed array and pad fixed arguments. No callers were found, so the comments describe the helper bodies rather than the currently active path. |
| census_small.go:109 | **Comment mismatch:** 'other slotless representations remain refused' excludes Union in the actual condition. slotless (expression.go:1071) recognizes only MaybeBoolean and Union; the helper exempts both. |
| census_small.go:114,131,158-159,169-170,184-186 | Only body-bearing implementation declarations are recovered; the relation applies classAssignable and widened; generic binders are mapped, extra binders inferred from the signature and no-evidence candidates still checked. Comments match the code. |
| census_small.go:215,218,240,270-272 | Extra parameters ignored, absent rest gives an empty array, nullable overload results are proved or checked and calls fit the resolved result representation. Comments match; readiness-fields.a covers the present false result. |
| census_small.go:332 | Callable slots share the tagged boolean exemption and additionally exempt Union. The comment does not say boolean is the only exemption; no contradiction established. |
| census_small.go:337,348-349,360,378-379,400-401,409-410 | Defaults add undefined to the incoming contract; boolean-or-undefined truth is exactly true; checker-never branches with two booleans retain Boolean; general conditions use one toBoolean argument; never-rest marker checks rest/never and storage can compare results. Comments match their code. |

Mutant observations

Run `python3 notes/coverage-oct8/typedarrays/mutants.py > /tmp/coverage-mutants.log 2>&1` with the setup environment sourced. Each mutation starts from the same restored source and the finally block restores it. All edits stay in named territory. Exact replacements, commands and exit codes are in mutants.json; complete catcher logs are beside this report. No killed mutant is claimed on the basis of clang warnings alone. The negative-wrap mutant fails with a release output mismatch (-2147483648 instead of 2147483647) and UBSan float-to-unsigned-char overflow. The same-kind guard mutant fails because a mismatched-kind set returns normally instead of panic 70. The callback argument guard survives the selected native tests, but its oracle fixture fails because backend exits 70 while native exits 0 and prints absent. Disabling both readiness call invalidation branches at readiness.go:190 and 212 survives the complete lower package (exit 0): category branch no package test caught. No inference that the branch is removable follows from this survival.

- `typed-negative-wrap` in `internal/native/runtime/typed_array.c:70`: `if (modulo < 0) { modulo += modulus; }` becomes `if (false && modulo < 0) { modulo += modulus; }`. Package command `go test ./internal/native -run ^TestTypedArrayRuntime$ -count=1 -timeout 10m` exits 1.
- `regexp-argument-guard` in `internal/native/runtime/regexp_replace.c:17`: `if (!accepted) {` becomes `if (false && !accepted) {`. Package command `go test ./internal/native -run RegexPrograms|ClosureConvention -count=1 -timeout 10m` exits 0. Additional uncached oracle command `go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/regexp_replace/argument_guard -count=1 -timeout 10m` exits 1.
- `readiness-call-invalidation` in `internal/lower/readiness.go:190`: `if readinessCalls(instruction) {` becomes `if false && readinessCalls(instruction) {`. Package command `go test ./internal/lower -count=1 -timeout 10m` exits 0.
- `typed-kind-guard` in `internal/native/runtime/typed_array.c:127`: `if (array->kind != source->kind) {` becomes `if (false && array->kind != source->kind) {`. Package command `go test ./internal/native -run ^TestTypedArrayRuntime$ -count=1 -timeout 10m` exits 1.

New-probe experiment `new-probes-negative-wrap` at typed_array.c:70: compiler build exit 0, run.py on uint8array.a exits 1. Exact edit: `if (modulo < 0) { modulo += modulus; }` to `if (false && modulo < 0) { modulo += modulus; }`. Full four-mode and leak streams are in new-probes-negative-wrap.json.

source: exit 0.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:0:0:Infinity
5:0:0:Infinity
6:255:255:0.00392156862745098
7:0:0:Infinity
8:255:255:0.00392156862745098
9:1:1:1
10:255:255:0.00392156862745098
11:255:255:0.00392156862745098
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:1
right:255
right:255
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:0
left:0
left:1
left:255
left:0
left:1
left:255
left:255
left:0
left:0
left:0
fill:255:255:255
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

sanitized: exit 1.

stdout:

```text
```

stderr:

```text
/home/agent/.cache/adamic/runtime/.build-26075203/typed_array.c:72:37: runtime error: -1 is outside the range of representable values of type 'unsigned char'
SUMMARY: UndefinedBehaviorSanitizer: undefined-behavior /home/agent/.cache/adamic/runtime/.build-26075203/typed_array.c:72:37 
```

release: exit 0.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:0:0:Infinity
5:0:0:Infinity
6:255:255:0.00392156862745098
7:0:0:Infinity
8:255:255:0.00392156862745098
9:1:1:1
10:255:255:0.00392156862745098
11:255:255:0.00392156862745098
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:1
right:255
right:255
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:0
left:0
left:1
left:255
left:0
left:1
left:255
left:255
left:0
left:0
left:0
fill:255:255:255
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

javascript: exit 0.

stdout:

```text
0:0:0:Infinity
1:0:0:Infinity
2:0:0:Infinity
3:0:0:Infinity
4:0:0:Infinity
5:0:0:Infinity
6:255:255:0.00392156862745098
7:0:0:Infinity
8:255:255:0.00392156862745098
9:1:1:1
10:255:255:0.00392156862745098
11:255:255:0.00392156862745098
12:0:0:Infinity
13:0:0:Infinity
14:0:0:Infinity
lengths:15:10:6
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:0
right:1
right:255
right:255
right:0
right:0
right:0
left:0
left:0
left:0
left:0
left:0
left:0
left:1
left:255
left:0
left:1
left:255
left:255
left:0
left:0
left:0
fill:255:255:255
empty:0:0
escaped:2:33:33
```

stderr:

```text
```


New-probe experiment `new-probes-extra-retain` at typed_array.c:149: compiler build exit 0, run.py on float64array.a exits 1. Exact edit: `view->owner = adamic_retain(array->owner != NULL ? array->owner : (adamic_typed_array *)array);` to `view->owner = adamic_retain(adamic_retain(array->owner != NULL ? array->owner : (adamic_typed_array *)array));`. Full four-mode and leak streams are in new-probes-extra-retain.json.

source: exit 0.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

sanitized: exit 0.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

release: exit 0.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

javascript: exit 0.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text
```

leaks: exit 1.

stdout:

```text
0:NaN:NaN:NaN
1:Infinity:Infinity:0
2:-Infinity:-Infinity:0
3:0:0:-Infinity
4:256.5:256.5:0.003898635477582846
5:-256.5:-256.5:-0.003898635477582846
6:2147483647.9:2147483647.9:4.656612873294233e-10
7:2147483648:2147483648:4.656612873077393e-10
8:-2147483649:-2147483649:-4.656612870908988e-10
9:4294967297:4294967297:2.3283064359965952e-10
10:-4294967297:-4294967297:-2.3283064359965952e-10
11:9007199254740991:9007199254740991:1.1102230246251568e-16
12:1e+100:1e+100:1e-100
13:-1e+100:-1e+100:-1e-100
14:5e-324:5e-324:Infinity
lengths:15:10:6
right:NaN
right:Infinity
right:-Infinity
right:0
right:256.5
right:-Infinity
right:0
right:256.5
right:-256.5
right:4294967297
right:-4294967297
right:9007199254740991
right:1e+100
right:-1e+100
right:5e-324
left:NaN
left:Infinity
left:-Infinity
left:0
left:256.5
left:-256.5
left:4294967297
left:-4294967297
left:-256.5
left:4294967297
left:-4294967297
left:9007199254740991
left:1e+100
left:-1e+100
left:5e-324
fill:-257.9:-257.9:-4294967297
empty:0:0
escaped:2:33:33
```

stderr:

```text

=================================================================
==14126==ERROR: LeakSanitizer: detected memory leaks

Direct leak of 96 byte(s) in 2 object(s) allocated from:
    #0 0x557ecc1e2274 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
    #1 0x557ecc238def in adamic_allocate /home/agent/.cache/adamic/runtime/.build-2623394424/heap.c:196:10
    #2 0x7ffa05a8dca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)

Indirect leak of 160 byte(s) in 2 object(s) allocated from:
    #0 0x557ecc1e2449 in calloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:74:3
    #1 0x557ecc27f087 in allocate /home/agent/.cache/adamic/runtime/.build-2623394424/typed_array.c:35:16
    #2 0x7ffa05a8dca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)

SUMMARY: AddressSanitizer: 256 byte(s) leaked in 4 allocation(s).
```


Failure categories for the new-probe mutants: negative wrap causes a runtime abort under UBSan in the new Uint8 probe, while its release and JavaScript outputs still match source Node. The existing package runtime test separately catches different release output on Int32; excess owner retain is memory not freed, while all four comparison modes still agree. These are deliberate mutants, not defects observed on main.

Inferences and limits

Agreement supports these concrete programs, not arbitrary correctness of the runtime or lowering. A surviving mutant establishes only that its listed tests did not notice the change. Historical readiness guard survival may be explained by its former syntax being unreachable on current main; that is an inference, not evidence that the guard is safe to remove. Native callback unit-test survival does not establish that the oracle lacks coverage; its separately run argument-guard fixture is the relevant external behavior check.

The safe multiargument buffer constructors, unsupported integer/float element kinds, cross-kind set, memory exhaustion/overflow allocation guards, callback group/rest guard combinations, throwing callback cleanup beyond existing oracle fixtures, arbitrary RegExp overrides and every readiness CFG path were not exhaustively covered. No production fix, PR, merge or full repository gate is part of this unit.

Validation logs

Exact unmutated commands (all output redirected to logs):

```sh
go test ./internal/lower ./internal/native -count=1 -timeout 30m > /tmp/coverage-packages.log 2>&1
go test ./internal/lower -count=1 -timeout 30m > /tmp/coverage-lower-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(typed_arrays_|regexp_replace|non_null_uninitialized)' -count=1 -timeout 30m > /tmp/coverage-oracle.log 2>&1
go test ./internal/native -run 'TestTypedArrayRuntime|RegexPrograms|ClosureConvention' -count=1 -timeout 10m > /tmp/coverage-native-focused-final.log 2>&1
go vet ./internal/lower ./internal/native > /tmp/coverage-vet.log 2>&1
git diff --check > /tmp/coverage-diffcheck.log 2>&1
```

The initial combined command exits 1 because lower lacks @types/node; its entire native package passes (437.854s). After installing the dependency, the entire lower package passes (49.131s). The uncached filtered oracle passes (42.486s). The final focused native command passes (1.495s). Vet and diff-check exit 0 with empty output. Final run.py exits 0 and checks seven agreeing/leak-clean programs plus two expected refusals. The full repository gate was not run. The later staged `git diff --cached --check` reported one trailing space in typed-negative-wrap.log:7 (exit 2). That space is the verbatim UBSan output and is intentionally preserved; the earlier unstaged diff check had no output.

`setup.log`:

```text
go version go1.27.1 linux/amd64
setup: go ready (0.131s)
v24.19.0
setup: node ready (0.148s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.517s)

added 3 packages in 1s
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=1.524s
setup: markdown dependencies ready (1.763s)
submodule cache: restored cohere 7945d102a6c18dd36adf9114a758ce646e8b2359
setup: submodules ready (25.052s)
setup: go build ready (366.299s)
setup: test binaries deferred (use --warm-tests) (366.638s)
setup: build cache warm (366.642s)
setup: build-flags commit=031a1259bc7973934792dc6cb1bd4074fc2204b9 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.22 0.06 0.02 1/136 717 load-after=11.75 9.38 4.23 11/200 4955
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (366.813s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.YEp2YR
```

`npm.log`:

```text

added 3 packages in 894ms
npm notice
npm notice New major version of npm available! 11.9.0 -> 12.2.0
npm notice Changelog: https://github.com/npm/cli/releases/tag/v12.2.0
npm notice To update run: npm install -g npm@12.2.0
npm notice
```

`build.log`:

```text
```

`runs.log`:

```text
buffer-gap False [('source', 0), ('sanitized', 1), ('release', 1), ('javascript', 1)]
float64array True [('source', 0), ('sanitized', 0), ('release', 0), ('javascript', 0)]
int32array True [('source', 0), ('sanitized', 0), ('release', 0), ('javascript', 0)]
readiness-callback False [('source', 0), ('sanitized', 1), ('release', 1), ('javascript', 1)]
readiness-fields True [('source', 0), ('sanitized', 0), ('release', 0), ('javascript', 0)]
readiness True [('source', 0), ('sanitized', 0), ('release', 0), ('javascript', 0)]
replacement True [('source', 0), ('sanitized', 0), ('release', 0), ('javascript', 0)]
rest-closure True [('source', 0), ('sanitized', 0), ('release', 0), ('javascript', 0)]
uint8array True [('source', 0), ('sanitized', 0), ('release', 0), ('javascript', 0)]
```

`packages-initial.log`:

```text
--- FAIL: TestNodeFSFileOptionsBorrow (0.03s)
    library_node_fs_file_test.go:14: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileDoesNotAuthorizeMutableWidening (0.03s)
    library_node_fs_file_test.go:21: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileRefusesOptionEffects (0.07s)
    library_node_fs_file_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileRefusesVoidValues (0.05s)
    library_node_fs_file_test.go:35: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileRefusesPinnedUnsupportedOverloads (0.21s)
    --- FAIL: TestNodeFSFileRefusesPinnedUnsupportedOverloads/readFileSync (0.04s)
        library_node_fs_file_test.go:50: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeFSFileRefusesPinnedUnsupportedOverloads/openSync (0.03s)
        library_node_fs_file_test.go:50: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeFSFileRefusesPinnedUnsupportedOverloads/statSync (0.03s)
        library_node_fs_file_test.go:50: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeFSFileRefusesPinnedUnsupportedOverloads/readSync (0.06s)
        library_node_fs_file_test.go:50: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeFSFileRefusesPinnedUnsupportedOverloads/writeFileSync (0.05s)
        library_node_fs_file_test.go:50: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileNamespaceImport (0.03s)
    library_node_fs_file_test.go:60: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileQualifiedErrorType (0.02s)
    library_node_fs_file_test.go:66: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileKeepsDetachedMethodRefusal (0.06s)
    library_node_fs_file_test.go:74: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileBufferBorrow (0.03s)
    library_node_fs_file_test.go:82: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileScratchOptionsBorrow (0.02s)
    library_node_fs_file_test.go:92: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeFSFileScratchOverloadsAreNamed (0.05s)
    library_node_fs_file_test.go:104: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeLibraryNamesUnimplementedMembers (0.45s)
    --- FAIL: TestNodeLibraryNamesUnimplementedMembers/node:fs.copyFileSync (0.03s)
        library_node_test.go:26: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeLibraryNamesUnimplementedMembers/node:fs.readFile (0.04s)
        library_node_test.go:26: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeLibraryNamesUnimplementedMembers/node:path.toNamespacedPath (0.04s)
        library_node_test.go:26: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeLibraryNamesUnimplementedMembers/node:os.userInfo (0.07s)
        library_node_test.go:26: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeLibraryNamesUnimplementedMembers/node:crypto.randomUUID (0.03s)
        library_node_test.go:26: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeLibraryNamesUnimplementedMembers/node:buffer.BufferConstructor.byteLength (0.10s)
        library_node_test.go:26: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeLibraryNamesUnimplementedMembers/node:path.sep (0.09s)
        library_node_test.go:26: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeLibraryNamesUnimplementedMembers/node:fs.readFile#01 (0.04s)
        library_node_test.go:26: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames (0.03s)
    library_node_test.go:40: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeLibraryDistinguishesReceiverOwners (0.03s)
    library_node_test.go:54: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeLibraryQualifiedTypeAssertion (0.03s)
    library_node_test.go:77: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNamespaceAmbientHostInitialization (0.05s)
    --- FAIL: TestNamespaceAmbientHostInitialization/realpath (0.03s)
        namespaces_ambient_test.go:22: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNamespaceAmbientHostInitialization/cwd (0.02s)
        namespaces_ambient_test.go:22: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
--- FAIL: TestNodeBufferRefusals (0.00s)
    --- FAIL: TestNodeBufferRefusals/buffer_alloc (0.04s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/encoding (0.02s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/structural_hash_view (0.01s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/typed_array_view (0.01s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/detached (0.01s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/digest (0.01s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/algorithm (0.02s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/catch (0.03s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/hash_inherited_read (0.01s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/crypto_constructor (0.07s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/hash_copy (0.03s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/buffer_read (0.03s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/crypto_random (0.10s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
    --- FAIL: TestNodeBufferRefusals/buffer_isUtf8 (0.09s)
        library_node_buffer_test.go:28: Load: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
FAIL
FAIL	github.com/system-inc/adamic/internal/lower	118.276s
ok  	github.com/system-inc/adamic/internal/native	437.854s
FAIL
```

`lower-final.log`:

```text
ok  	github.com/system-inc/adamic/internal/lower	49.131s
```

`oracle.log`:

```text
ok  	github.com/system-inc/adamic/internal/oracle	42.486s
```

`mutants.log`:

```text
{'mutant': 'typed-negative-wrap', 'file': 'internal/native/runtime/typed_array.c', 'before': 'if (modulo < 0) { modulo += modulus; }', 'after': 'if (false && modulo < 0) { modulo += modulus; }', 'command': ['go', 'test', './internal/native', '-run', '^TestTypedArrayRuntime$', '-count=1', '-timeout', '10m'], 'exit': 1}
{'mutant': 'regexp-argument-guard', 'file': 'internal/native/runtime/regexp_replace.c', 'before': 'if (!accepted) {', 'after': 'if (false && !accepted) {', 'command': ['go', 'test', './internal/native', '-run', 'RegexPrograms|ClosureConvention', '-count=1', '-timeout', '10m'], 'exit': 0, 'oracle': {'command': ['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_replace/argument_guard', '-count=1', '-timeout', '10m'], 'exit': 1}}
{'mutant': 'readiness-call-invalidation', 'file': 'internal/lower/readiness.go', 'before': 'if readinessCalls(instruction) {', 'after': 'if false && readinessCalls(instruction) {', 'command': ['go', 'test', './internal/lower', '-count=1', '-timeout', '10m'], 'exit': 0}
{'mutant': 'typed-kind-guard', 'file': 'internal/native/runtime/typed_array.c', 'before': 'if (array->kind != source->kind) {', 'after': 'if (false && array->kind != source->kind) {', 'command': ['go', 'test', './internal/native', '-run', '^TestTypedArrayRuntime$', '-count=1', '-timeout', '10m'], 'exit': 1}
```

`probe-mutants.log`:

```text
```

`native-focused-final.log`:

```text
ok  	github.com/system-inc/adamic/internal/native	1.495s
```

`vet.log`:

```text
```

`diffcheck.log`:

```text
```

