"""Push an apply checkpoint without adding a commit to the final branch history."""
import pathlib,subprocess,json,gzip,hashlib
repo=pathlib.Path('/workspace/adamic');out=repo/'stage3/census/sys-timer-slot';scratch=pathlib.Path('/workspace/scratch/sys-timer-slot')
record={'status':'apply complete, canonical verification pending','main':subprocess.check_output(['git','rev-parse','origin/main'],cwd=repo,text=True).strip(),'citedReportCommit':subprocess.check_output(['git','rev-parse','FETCH_HEAD'],cwd=repo,text=True).strip(),'setupExit':0,'apiInstallExit':0,'preliminaryApplyExit':0}
record['preliminarySourceHashes']={str(p.relative_to(scratch/'main')):hashlib.sha256(p.read_bytes()).hexdigest() for p in (scratch/'main/src').rglob('*') if p.is_file()}
(out/'checkpoint.json').write_text(json.dumps(record,indent=2)+'\n')
for name in ['setup.log','npm-api.log','apply.log','source-clone.log','fetch-core-report.log']:
    with gzip.open(out/(name+'.gz'),'wb') as f:f.write((scratch/name).read_bytes())
(out/'REPORT.md').write_text('| Measurement | Status |\n| --- | --- |\n| Full apply | Exit 0 |\n| Canonical verification, references and candidates | In progress |\n\nCheckpoint only. No declaration change lands.\n')
(out/'REPORT.json').write_text(json.dumps(record,indent=2)+'\n')
def run(label,args):
    p=subprocess.run(args,cwd=repo,capture_output=True,text=True);(scratch/(label+'.log')).write_text(p.stdout+p.stderr);print(label+' child_exit='+str(p.returncode),flush=True);p.check_returncode();return p.stdout.strip()
run('checkpoint-add',['git','add','--','stage3/census/sys-timer-slot'])
names=run('checkpoint-scope',['git','diff','--cached','--name-only']).splitlines();assert names and all(n.startswith('stage3/census/sys-timer-slot/') for n in names)
tree=run('checkpoint-tree',['git','write-tree']);sha=run('checkpoint-commit',['git','commit-tree',tree,'-p',record['main'],'-m','census: checkpoint System timer-slot measurement after apply'])
run('checkpoint-push',['git','push','origin',sha+':refs/heads/typescript/sys-timer-slot-checkpoint-b5w3ycg'])
(scratch/'checkpoint-sha.txt').write_text(sha+'\n');print(sha,flush=True)
