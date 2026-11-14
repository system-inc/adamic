import hashlib,json,shutil,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent; repo=Path('/workspace/scratch/changed-files-undefined-1');main=out/'main-tree';after=out/'after-tree'
shutil.copytree(main/'src',after/'src')
p=after/'src/compiler/sys.ts';old='const changedFilesInLastPoll: WatchedFileWithUnchangedPolls[] = [];';new='const changedFilesInLastPoll: (WatchedFileWithUnchangedPolls | undefined)[] = [];'
a=p.read_bytes();old=old.encode();new=new.encode();assert a.count(old)==1;p.write_bytes(a.replace(old,new))
pins=json.loads((repo/'stage3/adapt/99-correctness-timer-closures/owners.json').read_text());prereq=out/'contract-input'
for f in sorted({x['file'] for x in pins}):
 p=prereq/f;p.parent.mkdir(parents=True,exist_ok=True)
 p.write_bytes(subprocess.check_output(['git','show','refs/stage3/99-correctness-timer-closures-before:'+f],cwd=main))
manifest=lambda t:{str(p.relative_to(t)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(t.joinpath('src').rglob('*')) if p.is_file()}
m1,m2=manifest(main),manifest(after);delta=[p for p in m1 if m1[p]!=m2[p]];assert delta==['src/compiler/sys.ts'],delta
(out/'source-manifests.json').write_text(json.dumps(dict(main=m1,after=m2,changed=delta),indent=2)+'\n')
print(json.dumps(dict(changed=delta,main=m1[delta[0]],after=m2[delta[0]])))
