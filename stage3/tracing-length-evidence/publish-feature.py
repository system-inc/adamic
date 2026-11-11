from run import run,raw,root
import json,hashlib
base='16429a30452d027809f0ea85bc395479423e493e'
assert json.loads((raw/'final-validation-report.json').read_text())['readyForSingleFeatureCommit']
run('precommit-head',['git','rev-parse','HEAD']);assert (raw/'precommit-head.log').read_text().strip()==base
run('precommit-branch',['git','branch','--show-current']);assert (raw/'precommit-branch.log').read_text().strip()=='typescript/tracing-length'
run('precommit-unstaged',['git','diff','--name-only','--ignore-submodules']);assert not (raw/'precommit-unstaged.log').read_text()
for name,digest in json.loads((raw/'final-content.json').read_text())['files'].items():assert hashlib.sha256((root/'stage3/adapt/99-correctness-tracing-length'/name).read_bytes()).hexdigest()==digest,name
run('feature-commit',['git','commit','-m','Adapt tracing stack truncations and preserve reentrant RangeError'])
run('feature-head',['git','rev-parse','HEAD']);commit=(raw/'feature-head.log').read_text().strip()
run('feature-parent',['git','rev-parse','HEAD^']);assert (raw/'feature-parent.log').read_text().strip()==base
run('feature-commit-count',['git','rev-list','--count',base+'..HEAD']);assert (raw/'feature-commit-count.log').read_text().strip()=='1'
run('feature-changed-files',['git','diff','--name-only',base,'HEAD']);names=(raw/'feature-changed-files.log').read_text().splitlines();assert len(names)==4 and all(n=='stage3/upstream/LEDGER.md' or n.startswith('stage3/adapt/99-correctness-tracing-length/') for n in names)
run('feature-push',['git','push','origin','HEAD:refs/heads/typescript/tracing-length'])
(raw/'landing-report.json').write_text(json.dumps({'commit':commit,'parent':base,'commitsAboveBase':1,'changedFiles':names,'pushExit':0},indent=2)+'\n')
print('published one scoped feature commit: '+commit)
