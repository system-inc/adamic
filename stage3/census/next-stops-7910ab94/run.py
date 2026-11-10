"""Ordinary census, one file root per child, at most 20 roots per checkpoint."""
import hashlib,json,os,signal,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent
repo=Path('/workspace/adamic')
scratch=Path('/workspace/scratch')
tree=scratch/'next-stops-adapted'
def run(name,args,limit=1800):
    start=time.monotonic()
    with (out/(name+'.log')).open('wb') as log:
        child=subprocess.Popen(args,cwd=repo,stdout=log,stderr=log,start_new_session=True)
        stopped=None
        try:
            try: rc=child.wait(timeout=limit)
            except subprocess.TimeoutExpired:
                stopped='elapsed bound'
                os.killpg(child.pid,signal.SIGTERM)
                try: rc=child.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    os.killpg(child.pid,signal.SIGKILL);rc=child.wait()
        finally:
            try: os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError: pass
            child.wait()
    status=dict(command=args,exit_code=rc,seconds=time.monotonic()-start,stopping_reason=stopped,process_group_cleaned=True)
    (out/(name+'-status.json')).write_text(json.dumps(status,indent=2)+'\n')
    print(name+' child_exit='+str(rc),flush=True)
    if rc: raise SystemExit(rc if rc>0 else 128-rc)
    return status
def checkpoint(label):
    run('checkpoint-add-'+label,['git','add','--',str(out)])
    treeid=subprocess.check_output(['git','write-tree'],cwd=repo,text=True).strip()
    state=out/'checkpoint-state.json'
    parent=json.loads(state.read_text())['commit'] if state.exists() else subprocess.check_output(['git','rev-parse','7910ab94'],cwd=repo,text=True).strip()
    commit=subprocess.check_output(['git','commit-tree',treeid,'-p',parent,'-m','Next stops checkpoint: '+label],cwd=repo,text=True).strip()
    run('checkpoint-push-'+label,['git','push','origin',commit+':refs/heads/typescript/next-stops-7910ab94-checkpoint'])
    state.write_text(json.dumps(dict(commit=commit,label=label),indent=2)+'\n')
if __name__=='__main__':
    if sys.argv[1]=='groups':
        roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79,len(roots)
        (out/'roots.json').write_text(json.dumps(list(map(str,roots)),indent=2)+'\n')
        for begin in range(0,79,20):
            group=begin//20+1;rows=[]
            with (out/f'group-{group}.jsonl').open('wb') as stream:
                for index,root in enumerate(roots[begin:begin+20],start=begin+1):
                    dest=scratch/'next-stops-root.jsonl'
                    run(f'root-{index:02}',[str(scratch/'next-stops-census'),str(root),str(dest)])
                    data=dest.read_bytes();observations=[json.loads(line) for line in data.splitlines()]
                    assert len(observations)==2
                    assert {k:v for k,v in observations[0].items() if k!='seconds'}=={k:v for k,v in observations[1].items() if k!='seconds'}
                    stream.write(data);stream.flush();rows.append(observations[0])
                    (out/f'group-{group}-rows.json').write_text(json.dumps(rows,indent=2)+'\n')
                    print(f'group {group}: {index-begin}/{len(roots[begin:begin+20])} {root.name} {rows[-1]["kind"]}',flush=True)
            checkpoint('group-'+str(group))
    elif sys.argv[1]=='checkpoint':checkpoint(sys.argv[2])
    else:run(sys.argv[1],sys.argv[2:])
