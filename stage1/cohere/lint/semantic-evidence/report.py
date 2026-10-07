"""Semantic edit timings, observed failures and exact sample provenance."""
import json,pathlib,re
p=pathlib.Path(__file__).resolve().parent
rows=json.loads((p/'measurements.json').read_text());assert len(rows)==27
meta=json.loads((p/'build-flags.json').read_text())
lines=['Built: semantic message-ID token measurements and a short author command guide.\\',
 'Commits: implementation and measurement '+meta['commit']+'.\\',
 'Commands and outputs: nine semantic samples rebuilt objects; Node, emitted JavaScript and native agreed and were rejected by unchanged Go.\\',
 'Mutants: all nine message-ID changes caught by Go; a native-only emitted-C output mutant caught by semantic Node/native parity.\\',
 'Not covered: passing behavior-changing fixes, full repository gate or fifty-rule fleet; these samples stop at expected comparison failure.', '',
 'Each rule gets a separate empty cache for each of three rounds. The unedited author check primes both correct and owned-mutant builds and all observations. The whitespace check runs next, then the semantic token edit. Both use the same cache, toolchain and commit. The edit changes exactly one literal token in the private rule entry, with an anchor-count check; no rule directory is edited.', '',
 '| rule | entry | literal before | literal after |', '|---|---|---|---|',
 '| no-var | rule.a | unexpectedVar | unexpectedVaz |',
 '| no-empty | rule.ts | unexpectedBlock | unexpectedBlocx |',
 '| eqeqeq | rule.ts | unexpected | unexpectee |', '',
 'The semantic sample builds both the changed correct port and the changed owned-mutant port concurrently, with four jobs per split build. It runs the complete normal corpus: owned witnesses, captured upstream cases and inherited corner cases/options. All three port runtimes must match byte for byte, must differ from Go, and must fail certification. Execution remains under ASan/UBSan. Timing ends at that comparison failure, as an author check on a broken rule does. The later owned-mutant execution phase is not reached; its native build cost is included.', '',
 '| loop | before warm whitespace, seconds | after warm semantic, seconds | objects recompiled in semantic rounds 1/2/3 | instrument |', '|---|---:|---:|---|---|']
best={}
for slug in ['no-var','no-empty','eqeqeq']:
 for mode in ['warm-whitespace','warm-semantic']:
  best[slug,mode]=min((r for r in rows if r['slug']==slug and r['mode']==mode),key=lambda r:r['seconds'])
 a=best[slug,'warm-whitespace'];b=best[slug,'warm-semantic'];counts='/'.join(str(r['compiled_objects']) for r in rows if r['slug']==slug and r['mode']=='warm-semantic')
 command=f"go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule {slug} -rule-semantic-change"
 lines.append(f'| {slug} | {a["seconds"]:.3f} | {b["seconds"]:.3f} | {counts} | `{command}`; before substitutes `-rule-byte-change` |')
lines += ['', 'Best of three, wall-clock including `go test` startup/build overhead. Before is a successful whitespace check; after is an expected failure on incorrect diagnostic IDs, not a successful rule certification. Every semantic run has actual clang object compilation, so none is an emitted-C cache hit.', '',
 'Build-flags line for every sample: '+'; '.join(f'{k}={meta[k]}' for k in ['commit','nproc','cpu.max','go','clang','node','GOMAXPROCS','jobs','flags','GOCACHE'])+'. Native object commands also add `-Wno-gnu-line-marker -fdebug-prefix-map=<snapshot>=/adamic-units -x cpp-output -c`. The exact argument lists appear in compiler traces.', '',
 '| winning sample | round | cache | load before | load after | log |', '|---|---:|---|---|---|---|']
for slug in ['no-var','no-empty','eqeqeq']:
 for mode in ['warm-whitespace','warm-semantic']:
  r=best[slug,mode];lines.append(f'| {slug} {mode} | {r["round"]} | cached | `{r["load_before"]}` | `{r["load_after"]}` | [{r["log"]}]({r["log"]}) |')
lines += ['', 'All individual semantic samples and their largest phase:', '', '| rule | round | wall seconds | objects | port build phase seconds | emitted JavaScript build seconds | comparison seconds |', '|---|---:|---:|---:|---:|---:|---:|']
for r in rows:
 if r['mode']!='warm-semantic':continue
 text=(p/r['log']).read_text()
 phases={name:float(value) for name,value in re.findall(r'^\| ([^|]+) \| ([\d.]+) \|$',text,re.M)}
 lines.append(f'| {r["slug"]} | {r["round"]} | {r["seconds"]:.3f} | {r["compiled_objects"]} | {phases["sanitized correct and mutant port builds"]:.3f} | {phases["emitted JavaScript build"]:.3f} | {phases["Go Node JavaScript native comparison"]:.3f} |')
 trace=json.loads((p/r['trace']).read_text());objects=[event for event in trace if '-c' in event['argv'] and 'cpp-output' in event['argv']]
 names={event['argv'][event['argv'].index('cpp-output')+2] for event in objects};assert len(names)==34
lines += ['', 'Best-of-three semantic checks are under 20 seconds for all three rules, but individual samples are not uniformly below the target. Eqeqeq round 3 took 27.750 seconds wall-clock: 21.216 seconds were in the concurrent port-build phase, versus 2.604 for emitted JavaScript building and 1.161 for comparison. Broad native rebuild work dominates that slower sample; this is a measured failure-check time, not a latency guarantee.', '',
 'All 34 unit names (32 function groups, state and main) were rebuilt. The pair performs 35 physical object compilations for no-var/no-empty and 44 for eqeqeq because shared objects compile once while differing correct/mutant objects need separate versions. A literal edit can affect shared declarations and program-wide string numbering; this prototype provides no general one-object guarantee. Broad invalidation is observed in the traces. The port-build phase, including keying, loading/lowering, C emission, preprocessing, compilation and linking, is the largest measured phase. No compiler scheduling or native implementation changes were made.', '',
 'Correctness and reproduction:', '', '```bash', 'source /workspace/adamic-tools/env.sh', 'python3 stage1/cohere/lint/semantic-evidence/measure.py', 'python3 stage1/cohere/lint/semantic-evidence/validate.py', '```', '',
 '[validate.py](validate.py) specifies exact gate commands. [normal-parity.log](normal-parity.log) verifies selected/full registration, split/fresh whole-file output, cached/uncached bytes and compiler selection. [all-rule-uncached.log](all-rule-uncached.log) passes the complete all-rule upstream/inherited comparison on ordinary whole-file compilation. [vet.log](vet.log) passes touched packages. [native-parity-mutant.log](native-parity-mutant.log) adds a `puts` call to actual emitted C only during semantic checks: compilation and execution succeed, then the new parity check fails with `semantic edit Node/native outputs differ`. The nine semantic source mutants compile/run on all three port runtimes and fail only the Go comparison with the explicit expected marker.', '',
 '[measurements.json](measurements.json) contains every command, time, exit, cache state, full flags and load readings. Every `.clang.json` contains actual compiler invocations; objects are counted from `-c -x cpp-output`, not cache directory sizes. [setup.log](setup.log) records setup timings and build flags. The earlier whitespace timings remain as whitespace measurements, not semantic edit evidence.', '',
 'Scope: normal author checks remain unchanged in behavior when the benchmark flag is absent. No new cache was introduced or key weakened. Full repository, throughput/profile and fifty-rule-fleet gates were not run. Full all-rule comparison and focused parity were run. Future behavior-preserving port corrections can have different invalidation patterns and timing; these measurements establish observable ID edits and identical failures, not every possible edit.']
(p/'REPORT.md').write_text('\n'.join(lines)+'\n')
