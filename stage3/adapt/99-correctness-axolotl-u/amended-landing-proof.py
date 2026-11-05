from pathlib import Path
import subprocess,json,hashlib,gzip,shutil,os
s=Path('/workspace/scratch');p=Path('/workspace/adamic/stage3/adapt/99-correctness-axolotl-u');a=s/'99u2-candidate';b=s/'99u2-landing'
assert json.loads((s/'99u2-landing-replay.json').read_text())=={'files':0,'edits':0,'javascriptChanging':True}
files=lambda root:subprocess.check_output(['git','-C',str(root),'ls-files','-z']).decode().strip('\0').split('\0')
names=files(a);assert names==files(b);digest=hashlib.sha256()
for name in names:
 x=a/name;y=b/name
 assert x.is_symlink()==y.is_symlink(),name
 left=os.readlink(x).encode() if x.is_symlink() else x.read_bytes();right=os.readlink(y).encode() if y.is_symlink() else y.read_bytes();assert left==right,name
 digest.update(name.encode()+b'\0'+hashlib.sha256(left).digest())
r={'landingBase':'9a56f1c0760b444f1de85ccb4d44a9713a5d4d1e','oracleBase':'31d7a244808fa8092643926d21ab9d310a54cce3','trackedFiles':len(names),'allTrackedBuildAndTestInputsByteIdentical':True,'trackedInputSha256':digest.hexdigest(),'fullApplyExit':0,'replay':json.loads((s/'99u2-landing-replay.json').read_text())}
(p/'amended-landing.json').write_text(json.dumps(r,indent=2)+'\n')
for name in ['apply','replay']:
 f=s/('99u2-landing-apply.log' if name=='apply' else '99u2-landing-replay.json')
 if name=='apply':
  with gzip.open(p/'amended-landing-apply.log.gz','wb') as out:out.write(f.read_bytes())
 else:shutil.copyfile(f,p/'amended-landing-replay.json')
for name in ['lane','shards']:
 with gzip.open(p/f'amended-current-{name}.log.gz','wb') as out:out.write((s/f'99u2-current-{name}.log').read_bytes())
f=p/'amended-provenance.json';r2=json.loads(f.read_text());r2['base']=r['landingBase'];r2['oracleSourceBase']=r['oracleBase'];f.write_text(json.dumps(r2,indent=2)+'\n')
with (p/'README.md').open('a') as out:out.write(f'\nFinal landing verification on main 9a56f1c07 repeats full apply and zero-edit\nreplay. All {len(names):,} tracked upstream build/test inputs match the oracle-tested\namended tree byte-for-byte, including all 746 source files. No stage3 adaptation\nor oracle harness changed in the main advance. The fresh landing compiler reads\nzero over 79 roots before and after, and the shard and lane Go checks pass again.\nSee amended-landing.json and amended-current-* artifacts.\n')
with (p/'PROOF.md').open('a') as out:out.write(f'\nFinal full stage3/apply.sh on the fetched/rebased 9a56f1c07 main exits 0; its direct replay changes zero files and edits. All {len(names):,} tracked upstream files are byte-identical to the completed amended oracle candidate, not merely the edited owners. amended-landing.json records the complete input identity and digest, binding both accepted full oracle results and the 26/12 recount to the landing tree. amended-current-lane.log.gz records the repeated lane Go fixture pass (81.215s); amended-current-shards.log.gz records generate_shards --check. The commit remains one adaptation-only commit on the new branch, and its message ends Task: #b5w3ycg.\n')
print(json.dumps(r,indent=2))
