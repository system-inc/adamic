"""Finalize already published pieces; never starts a census or pushes."""
import gzip,hashlib,json,pathlib,shutil,subprocess,sys
O=pathlib.Path(__file__).resolve().parent
R=O.parents[2]
subprocess.run([sys.executable,str(O/'latent_report.py')],check=True,timeout=180)
D=json.loads((O/'latent-outcome.json').read_text())
assert D['common_timed_units']==4056 and D['regressions']==0 and D['cap_rows']==0
assert D['pair_counts']['judged']['max']==1084 and D['pair_counts']['past']['max']==79
for p in O.rglob('__pycache__'):
 shutil.rmtree(p)
for p in list((O/'evidence').rglob('*.log')):
 with gzip.open(str(p)+'.gz','wb') as f:f.write(p.read_bytes())
 p.unlink()
paths=subprocess.check_output(['git','diff','--name-only','d4c6b07a','HEAD'],cwd=R,text=True).splitlines()
assert all(p.startswith('stage3/census/container-fit/') for p in paths)
(O/'validation.txt').write_text('Complete primary latent census: 26 disjoint pieces, 79 files, 4557 units exactly once, 4056 common timed units.\nSource hashes, raw record hashes, selected unit identities and exact timer coverage validated by latent_report.py.\nAll seven public head/log chunk record streams byte-identical.\nZero regressions, zero panics, zero cap rows; 36 changed first stops in already-blocked units.\nFull latent Go subtree passes; broad fixed-array mutant fails the named struct-array refusal case.\nBase/head/log latent overlays compile; selector admission audit passes for all three.\nOrdinary 79-root results retained without changes to compiler behavior.\nAll report commits are scoped to stage3/census/container-fit/.\nSupplemental whole-public status is recorded separately; duplicate work excluded from primary totals.\n')
manifest={str(p.relative_to(O)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(O.rglob('*')) if p.is_file() and p.name!='artifacts.sha256.json'}
(O/'artifacts.sha256.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('Validated',len(manifest),'report artifacts')
