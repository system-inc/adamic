"""Publish one report commit on the pinned base; never force or push main."""
import importlib.util,json,re,sys
from pathlib import Path
sys.dont_write_bytecode=True
here=Path(__file__).resolve().parent;root=here.parents[2];scratch=Path('/workspace/scratch/jawqwv7')
spec=importlib.util.spec_from_file_location('supervise',here/'supervise.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
base='bcbe5d1244ee8a218d860415dea19cca114bae46';scope=str(here.relative_to(root))+'/'
branch='typescript/graph-boundary-census'
def must(name,args):
 result=m.run('publish-'+name,args,cwd=str(root),limit=180)
 if result['exit_code']!=0:raise RuntimeError(result)
 return (scratch/('publish-'+name+'.log')).read_text().strip()
# All report generation, recounting, snapshots and checksum verification precede
# this final publication step. Confirm their concrete reviewable artifacts.
validation=json.loads((here/'validation.json').read_text());assert validation['raw_recount_passed'] and validation['source_hashes_passed'] and validation['scope_passed']
summary=json.loads((here/'summary.json').read_text());assert validation['units']==summary['observed_unit_results']
changed=must('scope',['git','diff','--name-only',base]).splitlines();assert all(p.startswith(scope) for p in changed),changed
must('add',['git','add','--',scope])
staged=must('staged-scope',['git','diff','--cached','--name-only',base]).splitlines();assert staged and all(p.startswith(scope) for p in staged),staged
must('checkpoint-commit',['git','commit','-m','census checkpoint: final validation'])
must('checkpoint-push',['git','push','origin','HEAD:refs/heads/typescript/graph-boundary-census-checkpoint'])
tree=must('tree',['git','rev-parse','HEAD^{tree}']);assert re.fullmatch('[0-9a-f]{40}',tree)
sha=must('commit',['git','commit-tree',tree,'-p',base,'-m','stage3/census: graph adoption boundary counts on bcbe5d12 (jawqwv7)']);assert re.fullmatch('[0-9a-f]{40}',sha)
assert must('parent',['git','show','-s','--format=%P',sha])==base
assert must('count',['git','rev-list','--count',base+'..'+sha])=='1'
files=must('final-scope',['git','diff-tree','--no-commit-id','--name-only','-r',sha]).splitlines();assert files and all(p.startswith(scope) for p in files),files
must('branch',['git','branch',branch,sha]);must('switch',['git','switch',branch])
must('push',['git','push','origin',branch+':refs/heads/'+branch])
remote=must('remote',['git','ls-remote','--heads','origin',branch]);assert remote.split()[0]==sha,remote
result=dict(sha=sha,branch=branch,counts=summary['counts'],observed_unit_results=summary['observed_unit_results'],observed_statuses=summary['observed_statuses'],unobserved_stock_statuses=summary['unobserved_stock_statuses'],remote_verified=True)
(scratch/'final-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
