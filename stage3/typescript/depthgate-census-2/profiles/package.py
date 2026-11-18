from pathlib import Path
import json,gzip,shutil,hashlib,shlex,subprocess,re
W=Path('/workspace');root=W/'depthgate2-base/stage3/typescript/depthgate-census-2';out=root/'profiles';out.mkdir(exist_ok=True);a=W/'depthgate2-profile-analysis'
attr=json.loads((a/'attribution.json').read_text());behavior=json.loads((a/'behavior-comparison.json').read_text());assert set(behavior)=={'base','head'}
for v in behavior.values():assert all(c['byte_identical'] for c in v['comparisons'])
statuses={}
for side in ['base','head']:
 run=W/'depthgate2-profile-runs'/f'{side}-utilities-r1';dest=out/'runs'/side;dest.mkdir(parents=True,exist_ok=True)
 s=json.loads((run/'status.json').read_text());statuses[side]=s;assert s['exit_code']==0 and not s['timed_out']
 shutil.copy2(run/'cpu.pprof',out/(side+'.pprof'))
 for name in ['started.json','status.json']:shutil.copy2(run/name,dest/name)
 for name in ['latent.jsonl','stderr.log','stdout.log']:(dest/(name+'.gz')).write_bytes(gzip.compress((run/name).read_bytes(),mtime=0))
 d=out/'analysis'/side;d.mkdir(parents=True,exist_ok=True)
 for p in (a/side).iterdir():
  if p.name.endswith('.raw.txt'):(d/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
  else:shutil.copy2(p,d/p.name)
 (out/(side+'-findings-arrays.jsonl.gz')).write_bytes(gzip.compress((a/(side+'-findings-arrays.jsonl')).read_bytes(),mtime=0))
 prep=out/'build'/side;prep.mkdir(parents=True,exist_ok=True)
 for name in ['overlay.json','profile-hook.diff','build.log']:
  src=W/'depthgate2-profiles'/side/name
  if name.endswith('.diff'):(prep/(name+'.gz')).write_bytes(gzip.compress(src.read_bytes(),mtime=0))
  else:shutil.copy2(src,prep/name)
 assert (W/'depthgate2-profiles'/side/'entry_main.go').read_bytes()==(W/'depthgate2-profiles/base/entry_main.go').read_bytes()
shutil.copy2(W/'depthgate2-profiles/base/entry_main.go',out/'profile-overlay-main.go.txt')
for name in ['attribution.json','behavior-comparison.json','pprof-commands.json']:shutil.copy2(a/name,out/name)
for name in ['prepare','runner','compare','analyze','package','validate']:
 shutil.copy2(W/('depthgate2-profile-'+name+'.py'),out/(name+'.py'))
(out/'runner.log').write_bytes((W/'depthgate2-profile-runner.log').read_bytes())
pprofbuild=W/'depthgate2-pprof-build.log'
if pprofbuild.exists():shutil.copy2(pprofbuild,out/'pprof-build.log')
cmds=json.loads((a/'pprof-commands.json').read_text());(out/'commands.txt').write_text('''# Both entry mains were copied from the existing own-SHA canonical overlays.
# prepare.py asserts they are identical, adds the identical env-gated profiling hook,
# and replaces ONLY stage3/census/latent/tool/main.go in each scratch overlay map.
python3 /workspace/depthgate2-profile-prepare.py
# In each /workspace/depthgate2-{base,head}, with LABEL set accordingly:
# PATH=/workspace/adamic-tools/go/bin:/workspace/adamic-tools/bin:$PATH
# go build -buildvcs=false -overlay=/workspace/depthgate2-profiles/$LABEL/overlay.json -o /workspace/depthgate2-profiles/$LABEL/entry-census ./stage3/census/latent/tool
python3 /workspace/depthgate2-profile-runner.py > /workspace/depthgate2-profile-runner.log 2>&1
# The inherited Go distribution lacked a compiled pprof helper. AFTER BOTH runs:
# In /workspace/adamic-tools/go/src/cmd:
# /workspace/adamic-tools/go/bin/go build -o /workspace/adamic-tools/go/pkg/tool/linux_amd64/pprof cmd/pprof
python3 /workspace/depthgate2-profile-compare.py
python3 /workspace/depthgate2-profile-analyze.py

# Every actual pprof command used (full paths preserve provenance):
'''+''.join(shlex.join(c)+'\n' for c in cmds))
def table(headers,rows):return '\n'.join(['| '+' | '.join(headers)+' |','| '+' | '.join('---' for _ in headers)+' |']+['| '+' | '.join(map(str,r))+' |' for r in rows])
p=['# CPU profiles — depth-gate submission census\n']
p.append('Base: **`c56b022995e1e960bad71626fd2074bc46aee485`**. Head: **`26d208af47f81143ff1db2280a5b43292deb96ba`**. One utilities.ts full single-file run per side, sequential base then head, on the **same unchanged adapted tree** used for [census-2](REPORT.md). Both complete normally within 30 minutes.\n')
b=attr['base']['buckets'];h=attr['head']['buckets'];extra=attr['delta']['total_extra_cpu_seconds']
p.append('**The earlier cost gap is not reproduced by this profiled pair.** Wall time is 604.772 s base versus 604.656 s head; wait4 process CPU is 799.092 s versus 801.471 s. Census-2’s unprofiled wall-time means were 599.675 s versus 812.441 s (+35.5%). These profiles cannot establish where that earlier large difference was spent, and do not establish why it disappeared.\n')
p.append(f"Sampled CPU totals are **{b['total']['cpu_seconds']:.2f} s base** and **{h['total']['cpu_seconds']:.2f} s head**: **{extra:+.2f} CPU seconds**. These are sampled CPU totals, not wall time.\n")
p.append('## Capture and behavior check\n')
p.append('Both own-SHA entry tools were rebuilt with the **identical scratch-only `entry_main.go` hook**: import `runtime/pprof`; if `ADAMIC_CPU_PROFILE` is set, create that file and call `pprof.StartCPUProfile` at the start of main; defer `pprof.StopCPUProfile` and file close until main returns. This profiles loading/checking, full latent lowering and stream serialization, excluding process initialization before main. The Go default CPU sampling rate is used (100 Hz / 10 ms). No lowering implementation, adaptation or ledger is patched. [Exact hook diff](profiles/build/base/profile-hook.diff.gz), [identical full overlay main](profiles/profile-overlay-main.go.txt), [preparation script](profiles/prepare.py), [overlay maps](profiles/build/), and binary hashes in run statuses are retained.\n')
p.append('The canonical overlays are those generated from each source SHA for census-2. The base worktree now contains the earlier report commit `62eb8c2bb`; its compiler/source files are verified identical to c56b0229, with only the prior report/evidence added. The head worktree is at 26d208af. Both use the same Go toolchain as census-2. The distribution lacked a compiled `pprof` helper; its shipped `cmd/pprof` was built **after both measured runs**, so no build ran alongside either profile capture.\n')
p.append('Environment is unchanged from census-2 apart from the additional profile path: `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_ONLY_FILE=/workspace/depthgate2-adapted/src/compiler/utilities.ts`, `LATENT_TIMES=/dev/stderr`. Neither side sets `ADAMIC_PROBE_DEPTHGATE`. The command remains `entry-census /workspace/depthgate2-adapted/src/tsc/tsc.ts <run>/latent.jsonl`, using the rebuilt own-side binary and an absolute `ADAMIC_CPU_PROFILE=<run>/cpu.pprof`. Whole stdout/stderr and streams are archived; the 30-minute termination guard is unchanged and did not fire.\n')
p.append('**Findings match census-2 byte for byte, on each side, against both original repeats.** The comparison extracts each top-level serialized `findings` array directly from the original JSON text, preserves every byte within it, and concatenates the arrays in stream order with newlines. It does not re-encode, sort, strip positions, deduplicate, or alter messages. Timing/unit metadata outside these arrays is not part of this finding comparison. [Comparison method](profiles/compare.py) and [exact hashes/results](profiles/behavior-comparison.json).\n')
rows=[]
for side in ['base','head']:
 x=behavior[side];s=statuses[side];rows.append([side,x['findings'],x['serialized_array_bytes'],f"`{x['sha256']}`",'identical to repeats 1 and 2'])
p.append(table(['Side','Findings','Finding-array bytes','SHA256','Census-2 comparison'],rows)+'\n')
rows=[]
for side in ['base','head']:
 s=statuses[side];rows.append([side,f"{s['wall_seconds']:.3f}",f"{s['user_cpu_seconds']+s['system_cpu_seconds']:.3f}",f"{attr[side]['buckets']['total']['cpu_seconds']:.2f}",f"{s['peak_rss_kib']:,}",s['hostname'],s['exit_code']])
p.append(table(['Side','Wall s','wait4 user+system CPU s','Sampled CPU s','Peak RSS KiB','Machine','Exit'],rows)+'\n')
p.append('Profiles: **[base.pprof](profiles/base.pprof)** and **[head.pprof](profiles/head.pprof)**. Status metadata preserves timestamps, CPU/memory limits, selected-source hash, binary hash and exact environment. This is one profiled observation per side; profiling overhead and sampling variation apply.\n')
p.append('## Where the extra CPU sits\n')
p.append('Each row sums the **CPU nanosecond weight of every selected sample once** after `pprof -focus`/`-ignore`. Recursive cumulative rows are not added together. “Extra” is head minus base, and the final percentage divides by the total sampled CPU difference. These are stack attribution measurements, not an allocation-causality estimate.\n')
labels={
 'mentionsHeld':'mentionsHeld, cumulative (any call stack)',
 'walk_excluding_mentionsHeld':'containerWhereOtherGoes and callees, excluding mentionsHeld',
 'outside_walk_and_mentionsHeld':'All remaining stacks (disjoint remainder)',
 'typeNesting_global':'typeNesting and callees, globally (overlaps above)',
 'typeNesting_in_walk_excluding_mentionsHeld':'typeNesting within walk, excluding mentionsHeld (subset)',
 'walk_excluding_mentionsHeld_and_typeNesting':'Walk excluding mentionsHeld and typeNesting (subset)',
}
rows=[]
for name,label in labels.items():
 d=attr['delta']['buckets'][name];rows.append([label,f"{b[name]['cpu_seconds']:.2f}",f"{h[name]['cpu_seconds']:.2f}",f"{d['extra_cpu_seconds']:+.2f}",f"{d['percent_of_total_extra_cpu']:+.1f}%"])
p.append(table(['Selected stacks','Base CPU s','Head CPU s','Extra CPU s','Share of total extra CPU'],rows)+'\n')
p.append('The first three rows partition the total. The percentages use the small **2.20 s difference in this profile pair**, not census-2’s earlier gap; they are sensitive to sampling/run variation. A measured zero means **no observed samples**, not proof of zero work. **typeNesting has no observed samples inside the walk on either side** (globally it changes 0.63 → 0.73 s). The strict outer-frame and expanded recursive-pair-anchor views agree exactly: **0 sampled CPU seconds** have a pair frame but omit the outer frame after excluding mentionsHeld, so no ancestry correction is needed here. The Go runtime’s CPU stack limit is 64 PCs; stack-length diagnostics are retained.\n')
p.append('**typeNesting is overlapping evidence**, principally useful as a subset of the walk; do not add it to the walk row. The final two subset rows partition the walk excluding mentionsHeld. Background GC and other runtime work without the selected functions on its stack remain in the disjoint remainder; this analysis does not assign that work back to allocation sites. Exact integer weights and partition checks are in [attribution.json](profiles/attribution.json).\n')
p.append('Commands, repeated with each side’s binary/profile:\n\n```sh\nGO=/workspace/adamic-tools/go/bin/go\nBINARY=/workspace/depthgate2-profiles/HEAD_OR_BASE/entry-census\nPROFILE=/workspace/depthgate2-profile-runs/HEAD_OR_BASE-utilities-r1/cpu.pprof\n"$GO" tool pprof -raw "$BINARY" "$PROFILE"\n"$GO" tool pprof -raw -focus=mentionsHeld "$BINARY" "$PROFILE"\n"$GO" tool pprof -raw -focus=containerWhereOtherGoes -ignore=mentionsHeld "$BINARY" "$PROFILE"\n"$GO" tool pprof -raw -focus=typeNesting "$BINARY" "$PROFILE"\n"$GO" tool pprof -raw -focus=containerWhereOtherGoes -ignore=\'mentionsHeld|typeNesting\' "$BINARY" "$PROFILE"\n"$GO" tool pprof -raw -ignore=\'containerWhereOtherGoes|mentionsHeld\' "$BINARY" "$PROFILE"\n```\n')
p.append('Use lowercase `base` / `head` in those paths. The same filters were also run with `-top -cum` for readable focused reports. To obtain the typeNesting-in-walk intersection, the analyzer reads the walk-without-mentionsHeld raw samples and sums those whose raw location stack contains typeNesting; it does not OR the two focus expressions. [All exact commands](profiles/commands.txt), [analysis script](profiles/analyze.py), [focused/raw reports](profiles/analysis/).\n')
p.append('## pprof top reports — first 40 lines each\n\nThe excerpts below are the first 40 stdout lines from the literal commands shown, without adding a node-count or changing default thresholds. Full stdout/stderr is retained under `profiles/analysis/`.\n')
for side in ['base','head']:
 for mode,flags in [('top-cum','-top -cum'),('top','-top')]:
  p.append('### '+side+' — '+flags+'\n')
  p.append('```sh\n'+shlex.join([str(W/'adamic-tools/go/bin/go'),'tool','pprof',*flags.split(),str(W/'depthgate2-profiles'/side/'entry-census'),str(W/'depthgate2-profile-runs'/f'{side}-utilities-r1'/'cpu.pprof')])+'\n```\n')
  p.append('```text\n'+(a/side/(mode+'-first40.txt')).read_text().rstrip()+'\n```\n')
p.append('## Evidence and scope\n\n[Validation](profiles/validation.txt) verifies source hashes, identical hooks, own-SHA source provenance, run completion, byte-for-byte findings, profile readability, stack partition identities and lossless archives. [SHA256SUMS](profiles/SHA256SUMS) covers every other profile-evidence file. Only PROFILE.md and this new profile evidence are added to the existing report branch; no candidate code, adaptation, ledger or `internal/` change is committed.\n')
(root/'PROFILE.md').write_text('\n'.join(p).rstrip()+'\n')
print('Profile report and evidence packaged.')
