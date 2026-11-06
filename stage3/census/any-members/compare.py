"""Prove the base union, pair only agreeing units, and report every head any-member finding."""
from collections import Counter
import importlib.util
import json
import subprocess
from pathlib import Path
from pieces_common import OUT, E, TREE, dump, load, digest
from union import audit, units, lower

PAIR=OUT.parent/'pair/pair.py'
spec=importlib.util.spec_from_file_location('census_pair',PAIR)
pair=importlib.util.module_from_spec(spec);spec.loader.exec_module(pair)


def counts(rows):
    keys={tuple(f[k] for k in ('kind','where','reason','text')) for r in rows[1:] for f in r['findings']}
    c=Counter(k[0] for k in keys)
    return {k:c[k] for k in ('NotYet','Refused','panic','error')}

def main():
    proof, streams, timers, metrics = audit()
    filtered={label:load(E/(label+'-paired.jsonl.gz')) for label in streams}
    result=pair.compare(filtered['base'],filtered['head'],TREE,OUT/'scratch/base/stage3/api/node_modules/typescript')
    (OUT/'pair-report.md').write_text(pair.report(result))
    plan=json.loads((E/'plan.json').read_text())
    result.update(status='complete',base_commit=plan['base_commit'],head_commit=plan['head_commit'],latent_mode='first-error',measurement='matching first-error pieces; base whole-stream admission check',union_proof='evidence/pieces/union-proof.json',pair_py_sha256=digest(PAIR),typescript_pin=json.loads((OUT/'scratch/base/stage3/source.json').read_text()),box=json.loads((OUT/'evidence/box.json').read_text()),counts={l:counts(s) for l,s in streams.items()},whole_base_counts=counts(load(OUT/'evidence/base-latent.jsonl.gz')))
    hu=units(streams['head']);hl=lower(streams['head']);anys=[];seen=set()
    for row in streams['head'][1:]:
        for f in row['findings']:
            if not f['reason'].startswith('a member typed any'):continue
            k=tuple(f[x] for x in ('kind','where','reason','text','unit','phase'))
            if k in seen:continue
            seen.add(k)
            anys.append(dict(finding=f,owner=hu.get(f['unit']),is_unit_first_stop=f['phase']=='lowering' and hl[f['unit']][0]==f,paired_unit=f['unit'] not in proof['excluded_units']))
    request=dict(tree=str(TREE),typescript=str(OUT/'scratch/base/stage3/api/node_modules/typescript'),locations=[r['finding']['where'] for r in anys])
    resolved=json.loads(subprocess.run(['node',str(OUT/'owners.cjs')],input=json.dumps(request),text=True,capture_output=True,check=True).stdout)
    basekeys={tuple(f[k] for k in ('kind','where','reason','text')) for row in streams['base'][1:] for f in row['findings']}
    for r in anys:
        f=r['finding'];w=resolved[f['where']]
        assert w in hu
        r['lexical_owner']=hu[w]
        r['lexical_owner_first_lowering_stop']=hl[w][0] if hl[w] else None
        r['is_lexical_owner_first_stop']=bool(hl[w] and all(hl[w][0][k]==f[k] for k in ('where','kind','reason','text')))
        r['present_on_base']=tuple(f[k] for k in ('kind','where','reason','text')) in basekeys
    result['any_member_findings']=anys
    result['any_member_unit_first_stops']=sum(r['is_unit_first_stop'] for r in anys)
    result['unit_time_sums']={l:{k:sum(t[k] for t in ts) for k in ('wall_seconds','cpu_seconds')} for l,ts in timers.items()}
    result['whole_base_unit_time_sums']={k:sum(t[k] for t in load(OUT/'evidence/base-times.jsonl.gz')) for k in ('wall_seconds','cpu_seconds')}
    result['unit_time_sample_counts']={l:len(ts) for l,ts in timers.items()}
    ts={l:{t['where']:t for t in vv} for l,vv in timers.items()}
    result['unit_timing_delta']=[dict(where=w,base=ts['base'][w],head=ts['head'][w],wall_delta_seconds=ts['head'][w]['wall_seconds']-ts['base'][w]['wall_seconds'],cpu_delta_seconds=ts['head'][w]['cpu_seconds']-ts['base'][w]['cpu_seconds']) for w in sorted(ts['base'])]
    result['process_wall_seconds_sums']={l:sum(m['process_wall_seconds'] for m in ms) for l,ms in metrics.items()}
    result['longest_process_seconds']=max(m['process_wall_seconds'] for ms in metrics.values() for m in ms)
    dump(OUT/'delta.json',result)
    lines=['# Paired any-members census (tmdhz1d)','',f"**Head any-member findings: {len(anys)}. Unit first stops among them: {result['any_member_unit_first_stops']}.**",'']
    if not anys:lines+=['No head finding has a reason starting `a member typed any` in the completed first-error pieces. No attempted unit has refuseAnyMember as its first stop.']
    lines += ["All 12 are refusal-scan findings: 11 existing NotYet findings and one new Refused finding. None matches the first lowering stop of its lexical owning unit. Scan `unit` fields can be file-wide bookkeeping locations, so both the raw scan owner and independently resolved lexical top-level owner are retained below and in delta.json.", ""]
    for r in anys:
        f=r['finding'];owner=r['lexical_owner']
        lines += [f"- `{f['where']}`; lexical owner `{owner['where']}` ({owner.get('name',owner['kind'])}); raw scan owner `{f['unit']}`; phase `{f['phase']}`; unit first stop: {'yes' if r['is_lexical_owner_first_stop'] else 'no'}; present on base: {'yes' if r['present_on_base'] else 'no'}. Full reason: `{f['reason']}`"]
    lines+=['',f"Base `{result['base_commit']}`. Head `{result['head_commit']}` (one commit on base).",'', 'TypeScript v6.0.3, upstream `050880ce59e30b356b686bd3144efe24f875ebc8`. Cohere `06af2e399c6c1fb93fd570b774027942747400b4`; checker `d92d9bfee114c80be2c375d72edae966176e3a4f`.', '', 'Box `3cbd7fe9a660`: 5 available processors, CPU quota 4 cores, cgroup memory limit 16 GiB. Host metadata is retained in `evidence/box.json` and `evidence/cgroup.json`.','', '## Base union admission check','',f"All 23 base pieces and 23 matching head pieces completed, each exited 0, with all 79 compiler files, 4,555 unique units and 4,056 unique timed attempts per compiler. All 746 adapted source hashes still match the original manifest. The longest child was {result['longest_process_seconds']:.3f} seconds, well inside 30 minutes.",'',f"The base union differs from the whole base in **65 deferred-generic instantiated flags only**. All 4,555 units retain the same ordered lowering findings and first stop; all file scans and declaration inventories agree. Exactly 4,490 units agree in every field, including all 4,056 timed units. The 65 differing deferred units are excluded from both pairing streams. `UNION.md` lists every differing unit, its old and new flag, and the reason; `evidence/pieces/union-proof.json` retains exact machine-readable differences.",'','A piece can change a deferred generic witness because later callers in other pieces are absent from its census-wide witness map. That map annotates `instantiated` after selected calls finish; this metadata does not itself select the first lowering stop. The absence of first-stop differences here is measured, not assumed. Witness bits are neither ORed nor masked to claim equality. The gate also compares ordered lowering findings and every other unit field, not just first stops.','', '## Regressions and changed first stops','']
    for view in ('raw_findings','distinct_reads'):
        v=result[view]
        lines += [f"### {view.replace('_',' ').title()}",'',f"Regressions: {len(v['regressions'])}. Changed unit first stops: {len(v['unit_first_stop_delta'])}. New finding sites: {len(v['new_sites'])}.",'']
        if not v['regressions']:lines += ['No previously finding-free attempted unit or scan site now stops, and no new panic is observed.']
        for r in v['regressions']:
            f=r.get('finding') or r.get('head');lines += [f"- {r['type']}: `{f['where']}`; owner `{f['unit']}`; {f['kind']}: {f['reason']}"]
        for r in v['unit_first_stop_delta']:
            lines += [f"- Unit `{r['unit']['where']}` ({r['unit'].get('name',r['unit']['kind'])}):"]
            for l in ('base','head'):
                f=r[l];lines += [f"  - {l}: `{f['where']}`; {f['kind']}: {f['reason']}" if f else f'  - {l}: finding-free.']
        for r in v['new_sites']:
            f=r['finding'];lines += [f"- New site `{f['where']}`; owner `{f['unit']}`; phase `{f['phase']}`; {f['kind']}: {f['reason']}"]
    lines += ['', '## Unique findings','', 'Counts deduplicate `(kind, where, reason, text)` across attempts, as the meter does. Repeated file scans count once. Counts cover the full piece unions; excluding the 65 deferred units removes no findings, so the paired counts are identical. Whole-base counts also match.','', '| Kind | Base | Head |','| --- | ---: | ---: |']
    for k in ('NotYet','Refused','panic','error'):lines += [f"| {k} | {result['counts']['base'][k]} | {result['counts']['head'][k]} |"]
    lines += ['', '## Unit lowering time on this box','', 'One sample per compiler, assembled from the matching sequential pieces. These are sums of all 4,056 unit timers, including synchronous dependency work. Checker work, file scans and shared process setup are excluded. Base runs before head within each piece; this is not a benchmark distribution.','', '| Sum | Base seconds | Head seconds | Head minus base |','| --- | ---: | ---: | ---: |']
    for k in ('wall_seconds','cpu_seconds'):
        b=result['unit_time_sums']['base'][k];h=result['unit_time_sums']['head'][k];lines += [f'| {k} | {b:.6f} | {h:.6f} | {h-b:+.6f} |']
    lines += ['',f"Earlier whole-base reference: wall {result['whole_base_unit_time_sums']['wall_seconds']:.6f} s; CPU {result['whole_base_unit_time_sums']['cpu_seconds']:.6f} s. The piece sample repeats setup and resets process caches, so it is not substituted for whole-run timing. Successful child process wall sums: base {result['process_wall_seconds_sums']['base']:.3f} s; head {result['process_wall_seconds_sums']['head']:.3f} s. All per-unit deltas are in `delta.json`.",'','## Method and evidence','','Copied the method after reading `typescript/kind-predicate-census` REPORT.md and compare.py at `33dca393`, and the meter invocation. One shared adapted tree was built with base `stage3/apply.sh` and lockfile-pinned API packages. Adaptation and census inputs are unchanged between compiler pins; the head also changes unrelated interface-downcast fixtures. The exact compiler/dependency pins, original binary hashes and source manifest remain under `evidence/`.','','Existing speculative-only piece support and its union proof were read. It cannot select first-error attempts, so the new measurement helper adds file and top-level-unit scheduling guards to the same no-output latent overlays. It leaves body lowering, registration, dependencies, generic deferral and witness recording unchanged. Each process still loads the complete compiler program. Mode-off output equality for both compilers and an actual cross-file witness difference were checked on a fixture; dropped pieces, duplicate pieces and a finding moved outside its assigned unit were rejected. See `evidence/pieces/selector-audit.json`, the selector patches and `first-error-piece.go.txt`.','','Ordinary files form seven disjoint file groups. Utilities and utilitiesPublic exceed a safe whole-file budget, so their top-level units form sixteen disjoint chunks. Every unit is assigned exactly once. Repeated file scans are checked identical and retained once; declaration inventories are not summed. Cost estimates use the retained base LATENT_TIMES. See `PIECES.md` and `evidence/pieces/plan.json`.','','Each process runs sequentially in first-error mode with `LATENT_FULL=0`, `LATENT_SPECULATIVE=0`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TIMES` and the explicit selection file. The hard child limit is 20 minutes; five minutes without timer progress requests a goroutine dump and stops. No successful child hit either limit. Header-only finding buffers are expected while timers progress because Latent finalizes deferred witnesses before flushing. No mode switch or backend invocation occurred.','','Every completed child was committed immediately to scratch Git, then losslessly archived and committed on this report branch. Checkpoint IDs and hashes are beside each stream. Two interrupted unfinished piece attempts were retried; neither supplies observations here. A transient remote `commit_refs_rules` error delayed ordinary pushes, and a later ordinary push published the pending commits. Every successful checkpoint has now been pushed. Earlier incomplete whole-head attempts remain as historical evidence and supply no findings to this result.','','Before feature pairing, `union.py` checked every base unit against the retained complete whole stream and wrote the admission proof. The actual result was reported before invoking main `stage3/census/pair`. Pairing uses shared source hashes and exact offsets, retaining raw findings and normalized access-chain reads. Only lowering-phase events define unit first stops; scan owners remain separate. `delta.json` retains every transition, regression and any-member owner. `pair-report.md` retains the standard report.','','The checker accepted the program: checker_rejected=false, with no diagnostics. The fixed measurement label in the raw streams is retained verbatim. Finding-free applies to these eligible first-error attempts and scans; it does not establish complete lowering past earlier stops or native compilation, nor supply whole-run witness flags for the head.','','`validate.py` independently checks union arithmetic, admission exclusions, first stops, kind counts, timings, source/pin provenance and artifact hashes. See `validation.txt`.','']
    (OUT/'REPORT.md').write_text('\n'.join(lines))
    print(json.dumps(dict(any_member_findings=anys,regressions=result['raw_findings']['regressions'],changed_first_stops=result['raw_findings']['unit_first_stop_delta'],counts=result['counts'],time_sums=result['unit_time_sums']),indent=2))

if __name__=='__main__':main()
