from pathlib import Path
import json,shutil,difflib,gzip,hashlib
r=Path('/workspace/adamic/stage3/drivers/scanner/evidence/clean-records-rebase');r.mkdir(parents=True,exist_ok=True)
walk=Path('/workspace/scratch/native3-clean-walk');probes=Path('/workspace/scratch/native3-clean-probes')
stops=json.loads((walk/'stops.json').read_text()); witnesses=json.loads((probes/'results.json').read_text()); assert len(stops)==len(witnesses)==15
for s,w in zip(stops,witnesses):
 assert w['node_exit']==0 and w['build_exit']==1 and s['message'] in w['build_stderr']
for name in ['fetch','records-fetch','worktree','setup','compiler','records-merge','walk','witnesses']:
 shutil.copyfile('/tmp/scanner-clean-'+name+'.log',r/(name+'.log'))
with gzip.open(r/'library-merge.log.gz','wb') as dest:dest.write(Path('/tmp/scanner-clean-library-merge.log').read_bytes())
for src,name in [('/tmp/scanner-clean-walk.py','run-walk.py'),('/tmp/scanner-clean-witnesses.py','run-witnesses.py'),('/tmp/scanner-clean-stub.cjs','stub-function.cjs')]:shutil.copyfile(src,r/name)
(r/'witnesses').mkdir(exist_ok=True)
for p in probes.iterdir():
 if p.is_file() and not p.name.endswith('.native'):shutil.copyfile(p,r/'witnesses'/p.name)
for p in walk.glob('stop-*.stderr'):shutil.copyfile(p,r/p.name)
shutil.copyfile(walk/'stops.json',r/'stops.json')
mode_reports={}
for mode in ['unsplit','split']:
 source=Path('/workspace/scratch/native3-clean-'+mode);out=r/mode;out.mkdir(exist_ok=True)
 report=json.loads((source/'report.json').read_text());assert report['build_exit']==1 and report['node_exit']==0 and report['comparison_control_exit']==0 and report['end_mutant_diff_exit']==1
 a=(source/'node.stdout').read_bytes();b=(source/'node-end-mutant.stdout').read_bytes();assert len(a)==len(b) and sum(x!=y for x,y in zip(a,b))==1
 assert not (source/'full-tree-diff.stdout').read_bytes() and not (source/'full-tree-diff.stderr').read_bytes()
 mode_reports[mode]=report
 for name in ['report.json','files.json','build.stdout','build.stderr','node.stderr','full-tree-diff.stdout','full-tree-diff.stderr','control-diff.stdout','control-diff.stderr','mutant-diff.stdout','mutant-diff.stderr']:shutil.copyfile(source/name,out/name)
 shutil.copyfile('/tmp/scanner-clean-'+mode+'.log',out/'run.log')
diffs=[]
base=Path('/workspace/scratch/native3-slice')
for p in sorted((walk/'adapted').rglob('*.ts')):
 relative=p.relative_to(walk/'adapted');old=base/relative
 if old.exists() and old.read_bytes()!=p.read_bytes():diffs.extend(difflib.unified_diff(old.read_text().splitlines(True),p.read_text().splitlines(True),fromfile='original/'+str(relative),tofile='discovery/'+str(relative)))
(r/'discovery-only.json').write_text(json.dumps({'label':'uncommitted throwing discovery placeholders, not compiler or adaptation edits','diff':''.join(diffs)},indent=2)+'\n')
conflicts=[x.split('Merge conflict in ')[1] for x in (r/'records-merge.log').read_text().splitlines() if 'Merge conflict in ' in x];assert len(conflicts)==6
summary={'main':'ffe6efc1bd30de2539c1e1416230ca650105a0c7','library':'b05a9306dd9c49ea1697f9aea4e87fdc3a9d2795','requested_records':'fcddeb299460708f76e8436b4d887a260a6b84f1','scratch':'af708e4eb0b7ecea511f3883192f8f840e6dd831','records_merge':'conflicted, aborted and skipped','conflicts':conflicts,'setup_seconds':33.584,'nproc':5,'cpu_quota':4,'compiler':{'exit':0,'wall':141.791,'user':419.751,'sys':34.658,'sha256':hashlib.sha256(Path('/workspace/scratch/scanner-clean-adamic').read_bytes()).hexdigest()},'modes':mode_reports,'stops':15,'witnesses':{'node_successes':15,'build_refusals':15},'one_byte_mutants':{'unsplit':'Node output, caught diff exit 1','split':'Node output, caught diff exit 1'},'uncovered':'Native scanner execution, native timings, native byte mutant, clean records integration'}
(r/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print('PASS: fifteen ordered stops match fifteen Node-success/build-refusal witnesses; both reference/control diffs pass; both exactly-one-byte Node mutants fail diff.')
