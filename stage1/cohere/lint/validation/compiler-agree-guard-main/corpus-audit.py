import pathlib,subprocess,json
roots=[('/workspace/adamic','stage1'),('/workspace/scratch/typescript-6.0.3','src/compiler')]
report=[]
for repo,root in roots:
 walk={str(p.resolve()) for p in (pathlib.Path(repo)/root).rglob('*') if p.is_file() and p.suffix in ('.ts','.a')}
 tracked={str((pathlib.Path(repo)/p).resolve()) for p in subprocess.check_output(['git','-C',repo,'ls-files','--',root],text=True).splitlines() if pathlib.Path(p).suffix.lower() in ('.ts','.a')}
 report.append(dict(repository=repo,root=root,walk_count=len(walk),tracked_count=len(tracked),walk_only=sorted(walk-tracked),tracked_only=sorted(tracked-walk)))
print(json.dumps(report,indent=2))
