from pathlib import Path
import json,shutil,difflib,gzip,hashlib,re
r=Path('/workspace/adamic/stage3/drivers/scanner/evidence/land-area-next');r.mkdir(parents=True,exist_ok=True)
walk=Path('/workspace/scratch/native3-next-walk');probes=Path('/workspace/scratch/native3-next-probes');stops=json.loads((walk/'stops.json').read_text());witnesses=json.loads((probes/'results.json').read_text());byname={w['name']:w for w in witnesses};assert len(stops)==len(witnesses)==15
for s in stops:
 w=byname[Path(s['probe']).stem];assert w['node_exit']==0 and w['build_exit']==1 and s['message'] in w['build_stderr'],(s,w)
assert len({(x['file'],x['line'],x['column'],x['message']) for x in stops})==15
conflicts=Path('/tmp/scanner-next-conflicts.txt').read_text().splitlines();assert len(conflicts)==44
for name in ['fetch','worktree','setup','compiler','merge','api-install','walk','walk-ready','walk-resume1','walk-resume2','walk-resume3','walk-resume4','walk-resume5','walk-resume6','walk-resume7','walk-resume8','witnesses','witnesses-final','finalize']:
 p=Path('/tmp/scanner-next-'+name+'.log')
 if p.exists():shutil.copyfile(p,r/(name+'.log'))
shutil.copyfile('/tmp/scanner-next-conflicts.txt',r/'conflicts.txt');shutil.copyfile('/tmp/scanner-next-loader-prerequisite.stderr',r/'loader-prerequisite.stderr')
for src,name in [('/tmp/scanner-next-walk.py','run-walk.py'),('/tmp/scanner-next-witnesses.py','run-witnesses.py'),('/tmp/scanner-next-stub.cjs','stub-function.cjs'),('/tmp/scanner-next-finalize.py','finalize-stops.py')]:shutil.copyfile(src,r/name)
(r/'witnesses').mkdir(exist_ok=True)
for p in probes.iterdir():
 if p.is_file() and not p.name.endswith('.native'):shutil.copyfile(p,r/'witnesses'/p.name)
for p in walk.glob('stop-*.stderr'):shutil.copyfile(p,r/p.name)
for name in ['repeated-signature-attempts','repeated-driver-signature-attempt']:shutil.copytree(walk/name,r/name,dirs_exist_ok=True)
shutil.copyfile(walk/'driver-throw-unreachable-checker.stderr',r/'driver-throw-unreachable-checker.stderr');shutil.copyfile(walk/'stops.json',r/'stops.json')
with gzip.open(r/'discovery.c.gz','wb') as out:out.write(Path('/tmp/scanner-next-discovery.c').read_bytes())
(r/'clang-source-excerpt.txt').write_text('\n'.join(Path('/tmp/scanner-next-discovery.c').read_text().splitlines()[7890:7926])+'\n')
mode_reports={}
for mode in ['unsplit','split']:
 source=Path('/workspace/scratch/native3-next-'+mode+'-ready');out=r/mode;out.mkdir(exist_ok=True)
 report=json.loads((source/'report.json').read_text());assert report['build_exit']==1 and report['node_exit']==0 and report['comparison_control_exit']==0 and report['end_mutant_diff_exit']==1
 assert 'corePublic.ts:9:5' in (source/'build.stderr').read_text()
 a=(source/'node.stdout').read_bytes();b=(source/'node-end-mutant.stdout').read_bytes();assert len(a)==len(b) and sum(x!=y for x,y in zip(a,b))==1
 assert not (source/'full-tree-diff.stdout').read_bytes() and not (source/'full-tree-diff.stderr').read_bytes();mode_reports[mode]=report
 for name in ['report.json','files.json','build.stdout','build.stderr','node.stderr','full-tree-diff.stdout','full-tree-diff.stderr','control-diff.stdout','control-diff.stderr','mutant-diff.stdout','mutant-diff.stderr']:shutil.copyfile(source/name,out/name)
 shutil.copyfile('/tmp/scanner-next-'+mode+'-ready.log',out/'run.log')
 (out/'prerequisite').mkdir(exist_ok=True)
 for name in ['report.json','build.stderr']:shutil.copyfile(Path('/workspace/scratch/native3-next-'+mode)/name,out/'prerequisite'/name)
diffs=[];base=Path('/workspace/scratch/native3-slice')
for p in sorted((walk/'adapted').rglob('*.ts')):
 relative=p.relative_to(walk/'adapted');old=base/relative
 if old.exists() and old.read_bytes()!=p.read_bytes():diffs.extend(difflib.unified_diff(old.read_text().splitlines(True),p.read_text().splitlines(True),fromfile='original/'+str(relative),tofile='discovery/'+str(relative)))
a=Path('/workspace/scratch/native3-combined-unsplit/main.a');b=walk/'main.a';diffs.extend(difflib.unified_diff(a.read_text().splitlines(True),b.read_text().splitlines(True),fromfile='original/main.a',tofile='discovery/main.a'))
(r/'discovery-only.json').write_text(json.dumps({'label':'uncommitted throwing discovery placeholders and explicit namespace/type placeholders; no compiler/adaptation implementation','diff':''.join(diffs)},indent=2)+'\n')
error_control=json.loads((probes/'error-capture-control.json').read_text());assert error_control['node_exit']==0 and error_control['build_exit']==1
summary={'base':'28285421bf59ea46269545144c19123156c00c91','compiler_area_next':'337aa466','combined':'2648ad3362812b0f184a8cb922bc322d18729c66','merge':'44 conflicts; aborted; measurement uses landing base alone','conflicts':conflicts,'setup_seconds':63.387,'nproc':5,'cpu_quota':4,'compiler':{'exit':0,'wall':185.816,'user':464.636,'sys':35.133,'sha256':hashlib.sha256(Path('/workspace/scratch/scanner-next-adamic').read_bytes()).hexdigest()},'modes':mode_reports,'stops':{'lowering':14,'native_clang':1},'witnesses':{'node_successes':15,'matching_build_failures':15},'error_constructor_supplementary':error_control,'one_byte_mutants':{'unsplit':'Node output, caught diff exit 1','split':'Node output, caught diff exit 1'},'excluded':'loader prerequisite, six repeated diagnostics, throwing IIFE unreachable-code checker artifact, superseded compiler-area attempt','uncovered':'Native scanner execution/timing/output mutant, merged combined-compiler measurement, full gate, parser-directed rescans'}
(r/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
# Preserve exact logs, but strip trailing whitespace from readable copies for diff --check.
raw={}
for p in r.rglob('*'):
 if p.is_file() and p.suffix in ['.log','.stderr']:
  t=p.read_text()
  if any(x!=x.rstrip() for x in t.splitlines()):raw[str(p.relative_to(r))]=t;p.write_text('\n'.join(x.rstrip() for x in t.splitlines())+'\n')
(r/'raw-whitespace-logs.json').write_text(json.dumps(raw,indent=2)+'\n')
print('PASS: 44 conflicts recorded; fifteen distinct stops match Node-success/build-failure witnesses; supplementary ErrorConstructor blocker confirmed; both full-tree/control diffs pass and exactly-one-byte mutants are caught.')
