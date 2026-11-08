import json,re,collections,hashlib,tarfile
from pathlib import Path
from runner import out,raw,repo
prov=json.loads((out/'provenance.json').read_text())
def readside(side):
 rows=sum([json.loads((out/f'{side}-group-{g}.json').read_text()) for g in range(1,5)],[])
 records=[json.loads(l) for l in (raw/(side+'-core.jsonl')).read_text().splitlines()]
 record=next(r for r in records if r.get('file','').endswith('/core.ts'))
 sites={}
 for f in record['findings']:
  if f['kind']=='Boundary' or '/core.ts:' not in f.get('where',''):continue
  key=tuple(f.get(k,'') for k in ['where','phase','kind','reason'])
  sites[key]=f
 return rows,sorted(sites.values(),key=lambda f:(tuple(map(int,re.search(r':(\d+):(\d+)$',f['where']).groups())),f['phase']))
base,bs=readside('base');head,hs=readside('head')
assert len(base)==len(head)==79
assert len({r['roots'][0] for r in base})==79
def norm(where):return where.replace('/workspace/require-row-adapted/src/compiler/','')
def key(r):return (norm(r.get('where','-')),r['kind'],r.get('reason',''))
def esc(s):return str(s).replace('|','\\|').replace('\n',' ')
groups=collections.defaultdict(list)
comparison=[]
for b,h in zip(base,head):
 assert b['roots']==h['roots']
 root=norm(b['roots'][0]);groups[(key(b),key(h))].append(root)
 comparison.append(dict(root=root,base=b,head=h,identical=key(b)==key(h)))
(out/'comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
(out/'base-sites.json').write_text(json.dumps(bs,indent=2)+'\n');(out/'head-sites.json').write_text(json.dumps(hs,indent=2)+'\n')
lines=['# Require row paired census','','Task 00qn5az. Report only.','','Base: `'+prov['base']+'`. Head: `'+prov['head']+'`. The CLI tip was checked before setup and still matched the requested head. TypeScript 6.0.3 source pin: `'+prov['typescript_pin']+'`.','','Box: `'+prov['box']+'`; '+str(prov['visible_cpus'])+' visible CPUs; CPU quota `'+prov['cpu_quota'].strip()+'`; '+next(l for l in prov['meminfo'].splitlines() if l.startswith('MemTotal:'))+'. See provenance.json for affinity, memory limit and exact tree.','','## Paired first stops','','All 79 roots, sorted with Python Path ordering, ran individually in groups of 20, 20, 20 and 19 on each side. Each invocation produced two observations. All non-timing fields agreed; only the first counts. The ordinary census uses no overlay.','','| Roots | Base location | Base kind | Base reason | Head location | Head kind | Head reason | Result | Root names |','| ---: | --- | --- | --- | --- | --- | --- | --- | --- |']
for (b,h),roots in sorted(groups.items(),key=lambda x:-len(x[1])):
 lines.append('| '+' | '.join(map(esc,[len(roots),*b,*h,'identical' if b==h else 'CHANGED',', '.join(roots)]))+' |')
lines+=['','Exact location/kind/reason comparison: '+str(sum(r['identical'] for r in comparison))+'/79 identical. Every root and both observations are retained. Accepted means the entire resolved root program lowers; no native output is measured.']
# Full union preserves every source-order phase/kind/reason site.
def sk(f):return (norm(f['where']),f['phase'],f['kind'],f['reason'])
bset={sk(f) for f in bs};hset={sk(f) for f in hs}
allkeys=sorted(bset|hset,key=lambda k:(tuple(map(int,re.search(r':(\d+):(\d+)$',k[0]).groups())),k[1:]))
lines+=['','## Full core.ts site list','','Every distinct location/phase/kind/reason finding is listed in source order, including lowering findings. Boundary markers are retained only in raw streams. Sites at line 2599 and later are marked `*`. Repeated recording events are deduplicated here and retained in the raw stream.','','| Mark | Location | Phase | Kind | Exact reason | Base | Head |','| --- | --- | --- | --- | --- | --- | --- |']
for k in allkeys:
 mark='*' if int(re.search(r':(\d+):',k[0]).group(1))>=2599 else ''
 lines.append('| '+' | '.join(map(esc,[mark,k[0],k[1],k[2],'`'+k[3]+'`','present' if k in bset else 'absent','present' if k in hset else 'absent']))+' |')
after=[f for f in hs if f['phase']=='refusal_scan' and tuple(map(int,re.search(r':(\d+):(\d+)$',f['where']).groups()))>(2599,15)]
removed=sorted(bset-hset);added=sorted(hset-bset)
lines+=['','Removed sites: '+str(len(removed))+'. Added sites: '+str(len(added))+'. The require declaration at core.ts:2599:80 is '+('gone at head' if any(k[0]=='core.ts:2599:80' and k[3]=='an ambient value declaration' for k in removed) else 'not confirmed removed')+'.']
lines+=['','The first refusal-scan site after process at head is `'+norm(after[0]['where'])+'`, '+after[0]['kind']+': '+'`'+after[0]['reason']+'`'+'. This is the next currently recorded refusal-scan site. Its reason still names the pending process row, so it is not an independent forecast after Runtime clears process. These measurements do not establish where the 73 go after that clearance; no process-cleared branch exists in this comparison.','','Head sites from line 2599 onward, in source order:']
for f in hs:
 if int(re.search(r':(\d+):(\d+)$',f['where']).group(1))>=2599:lines+=['','- `'+norm(f['where'])+'` '+f['phase']+' '+f['kind']+': '+f['reason']]
lines+=['','All four refusal-scan sites remaining in head core.ts name the pending process row. The later scan sites arise from hostBindingRefusal at internal/lower/host_bindings.go:268-269, which checks the row stop before allowing typeof at lines 275-276. Clearing the row can therefore clear these sites as well. The seven earlier lowering findings are independent unit-probe observations, including generic and rollback-dependent findings; their source order does not establish an ordinary next stop. Runtime clearance followed by another paired ordinary census is needed to satisfy a forecast beyond process.']
lines+=['','## Run labels','','Peak memory is Linux wait4 ru_maxrss in KiB for each measured child, including inherited child usage under Linux accounting. Times are UTC. Each child exit and process-group cleanup is recorded.','','| Run | Start UTC | End UTC | Seconds | Peak RSS KiB | Child exit |','| --- | --- | --- | ---: | ---: | ---: |']
for p in sorted(raw.glob('*-status.json')):
 s=json.loads(p.read_text())
 if 'peak_rss_kib' in s:lines.append('| '+' | '.join(map(esc,[p.name.removesuffix('-status.json'),s['started_utc'],s['ended_utc'],round(s['seconds'],3),s['peak_rss_kib'],s['exit_code']]))+' |')
lines+=['','## Method and evidence','','One complete apply from base supplies the identical adapted tree for both sides. Head changes no adapter files. Ordinary tools and guarded full latent tools were built from the respective exact commits. Full latent runs use LATENT_FULL=1, LATENT_ONLY_FILE set to the adapted core.ts, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_TIMES for progress. The latent tool receives all compiler roots as load context, then selects core.ts alone. No usable IR is produced. Deferred generic bodies and independent lowering findings are not evidence of successful whole-program lowering. Refusal scan precedes lowering, so source order across phases is not the ordinary stop sequence.','','The c3b22184 reference method was read directly from commit 9714ab2dfe3a5dab10c5cbc7e42fb9238f91278f. The named core-sites-af42d3dc report was not present in the fetched main or current core-sites and core-sites-checkpoint branches. The available core-sites-4186581e report and runners were inspected for the same explicitly requested full latent mode.','','Completed apply and each root group were pushed append-only to typescript/require-row-census-checkpoint. The final branch has one report commit directly on 71b126b0. No main push or force push.','','Raw streams, logs, overlays and status files are in raw-streams.tar.gz. Binaries are excluded; binary hashes and build information are retained. Verify with sha256sum -c raw-streams.tar.gz.sha256. runner.py records exact child exits and reaps all process groups, with a 30-minute no-output bound.']
(out/'REPORT.md').write_text('\n'.join(lines)+'\n')
assert not any(c in (out/'REPORT.md').read_text() for c in ['\u2013','\u2014'])
print('identical',sum(r['identical'] for r in comparison),'sites',len(bs),len(hs),'forecast',after[0])
