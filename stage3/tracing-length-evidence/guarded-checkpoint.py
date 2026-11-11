from run import run,raw
import os,json
index=raw/'checkpoint.index';env=dict(os.environ,GIT_INDEX_FILE=str(index));base='16429a30452d027809f0ea85bc395479423e493e';parent=json.loads((raw/'checkpoint.json').read_text())['commit']
run('guarded-checkpoint-read-tree',['git','read-tree',base],env=env)
run('guarded-checkpoint-add',['git','add','stage3/adapt/99-correctness-tracing-length','stage3/upstream/LEDGER.md'],env=env)
run('guarded-checkpoint-write-tree',['git','write-tree'],env=env);tree=(raw/'guarded-checkpoint-write-tree.log').read_text().strip()
run('guarded-checkpoint-commit',['git','commit-tree',tree,'-p',parent,'-m','Checkpoint guarded tracing truncations and reentrant RangeError finding'],env=env)
commit=(raw/'guarded-checkpoint-commit.log').read_text().strip();assert len(commit)==40
run('guarded-checkpoint-push',['git','push','origin',commit+':refs/heads/typescript/tracing-length-checkpoint'])
(raw/'guarded-checkpoint.json').write_text(json.dumps({'commit':commit,'parent':parent,'branch':'typescript/tracing-length-checkpoint','finalFeatureBranchUncommitted':True},indent=2)+'\n')
