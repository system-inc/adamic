from pathlib import Path
import json,shutil,difflib,gzip,hashlib
r=Path('/workspace/adamic/stage3/drivers/scanner/evidence/combined-records-library');r.mkdir(parents=True,exist_ok=True)
walk=Path('/workspace/scratch/native3-combined-walk');probes=Path('/workspace/scratch/native3-combined-probes')
stops=json.loads((walk/'stops.json').read_text());witnesses=json.loads((probes/'results.json').read_text());byname={w['name']:w for w in witnesses};assert len(stops)==15 and len(witnesses)==16
for s in stops:
 w=byname[Path(s['probe']).stem];assert w['node_exit']==0 and w['build_exit']==1 and s['message'] in w['build_stderr']
 if s['order'] in [13,15]:s['qualification']='Depends on the Debug object-of-functions throwing placeholder; do not attribute to original namespace lowering.'
assert byname['01-index']['node_exit']==byname['01-index']['build_exit']==0
control=json.loads((probes/'01-index.control.json').read_text());assert control['diff_exit']==0 and control['one_byte_mutant_diff_exit']==1
for name in ['fetch','worktree','setup','compiler','walk','witnesses','new-witness','maplike']:
 shutil.copyfile('/tmp/scanner-combined-'+name+'.log',r/(name+'.log'))
with gzip.open(r/'merge.log.gz','wb') as dest:dest.write(Path('/tmp/scanner-combined-merge.log').read_bytes())
for src,name in [('/tmp/scanner-combined-walk.py','run-walk.py'),('/tmp/scanner-combined-witnesses.py','run-witnesses.py'),('/tmp/scanner-combined-stub.cjs','stub-function.cjs'),('/tmp/scanner-combined-new-witness.py','new-witness.py'),('/tmp/scanner-combined-maplike.py','maplike-control.py')]:shutil.copyfile(src,r/name)
(r/'witnesses').mkdir(exist_ok=True)
for p in probes.iterdir():
 if p.is_file() and not p.name.endswith('.native'):shutil.copyfile(p,r/'witnesses'/p.name)
for p in walk.glob('stop-*.stderr'):shutil.copyfile(p,r/p.name)
(r/'stops.json').write_text(json.dumps(stops,indent=2)+'\n')
mode_reports={}
for mode in ['unsplit','split']:
 source=Path('/workspace/scratch/native3-combined-'+mode);out=r/mode;out.mkdir(exist_ok=True)
 report=json.loads((source/'report.json').read_text());assert report['build_exit']==1 and report['node_exit']==0 and report['comparison_control_exit']==0 and report['end_mutant_diff_exit']==1
 assert 'debug.ts:7:1' in (source/'build.stderr').read_text()
 a=(source/'node.stdout').read_bytes();b=(source/'node-end-mutant.stdout').read_bytes();assert len(a)==len(b) and sum(x!=y for x,y in zip(a,b))==1
 assert not (source/'full-tree-diff.stdout').read_bytes() and not (source/'full-tree-diff.stderr').read_bytes()
 mode_reports[mode]=report
 for name in ['report.json','files.json','build.stdout','build.stderr','node.stderr','full-tree-diff.stdout','full-tree-diff.stderr','control-diff.stdout','control-diff.stderr','mutant-diff.stdout','mutant-diff.stderr']:shutil.copyfile(source/name,out/name)
 shutil.copyfile('/tmp/scanner-combined-'+mode+'.log',out/'run.log')
diffs=[];base=Path('/workspace/scratch/native3-slice')
for p in sorted((walk/'adapted').rglob('*.ts')):
 relative=p.relative_to(walk/'adapted');old=base/relative
 if old.exists() and old.read_bytes()!=p.read_bytes():diffs.extend(difflib.unified_diff(old.read_text().splitlines(True),p.read_text().splitlines(True),fromfile='original/'+str(relative),tofile='discovery/'+str(relative)))
(r/'discovery-only.json').write_text(json.dumps({'label':'uncommitted throwing discovery placeholders, not compiler or adaptation edits','diff':''.join(diffs)},indent=2)+'\n')
summary={'main':'6998ebc24ae353193cb1495d3d51308131a4b5c7','library_base':'e40216f23caa519ad90e011dfc3a47d5c3bfeb6e','combined':'2648ad3362812b0f184a8cb922bc322d18729c66','scratch':'dbd7a7c84952b31aa1bd3d723b58d71c4dc1430f','merge':'successful, no conflicts, no hand resolutions','setup_seconds':54.226,'nproc':5,'cpu_quota':4,'compiler':{'exit':0,'wall':186.725,'user':450.874,'sys':48.701,'sha256':hashlib.sha256(Path('/workspace/scratch/scanner-combined-adamic').read_bytes()).hexdigest()},'modes':mode_reports,'stops':15,'witnesses':{'node_successes':15,'build_refusals':15},'closed_blocker':{'previous_order':1,'file':'corePublic.ts','line':9,'column':5,'message':'index signature refusal','control':control},'remaining_previous_orders':list(range(2,16)),'new_stop_15':'Debug throwing placeholder optional parameter, explicitly qualified','one_byte_mutants':{'unsplit':'Node output, caught diff exit 1','split':'Node output, caught diff exit 1','MapLike_control':'native control output, caught diff exit 1'},'uncovered':'Native scanner execution, timings and output mutant; full repository gate; parser-directed rescans'}
(r/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print('PASS: fifteen Node-success/build-refusal message pairs; both reference/control diffs pass; both exactly-one-byte Node mutants caught; former MapLike control matches native and its one-byte mutant is caught.')
