from run import run,raw,root
import os,json
assert not (raw/'checkpoint-remote.log').read_text().strip(),'checkpoint branch already exists; inspect before pushing'
index=raw/'checkpoint.index';env=dict(os.environ,GIT_INDEX_FILE=str(index))
base='16429a30452d027809f0ea85bc395479423e493e'
run('checkpoint-read-tree',['git','read-tree',base],env=env)
run('checkpoint-add',['git','add','stage3/adapt/99-correctness-tracing-length'],env=env)
run('checkpoint-write-tree',['git','write-tree'],env=env);tree=(raw/'checkpoint-write-tree.log').read_text().strip()
run('checkpoint-commit',['git','commit-tree',tree,'-p',base,'-m','Checkpoint tracing length adaptation; validation in progress'],env=env)
commit=(raw/'checkpoint-commit.log').read_text().strip();assert len(commit)==40
run('checkpoint-push',['git','push','origin',commit+':refs/heads/typescript/tracing-length-checkpoint'])
(raw/'checkpoint.json').write_text(json.dumps({'commit':commit,'branch':'typescript/tracing-length-checkpoint','finalFeatureBranchUncommitted':True},indent=2)+'\n')
