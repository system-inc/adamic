import json,pathlib,shlex,subprocess
root=pathlib.Path('/workspace/adamic/evidence/f7277112')
work=pathlib.Path('/workspace/adamic-rule-g-2594029')
pin='f727711256b624ab4fb025e2facdedffc858e7ef'
results={}
def run(label,args):
 args=list(map(str,args));(root/(label+'.command')).write_text(shlex.join(args)+'\n')
 r=subprocess.run(args,cwd=work,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 (root/(label+'.stdout')).write_bytes(r.stdout);(root/(label+'.stderr')).write_bytes(r.stderr)
 (root/(label+'.exit')).write_text(str(r.returncode)+'\n')
 results[label]={'exit':r.returncode,'stdout':r.stdout.decode(errors='replace'),'stderr':r.stderr.decode(errors='replace')}
 (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
 return r.returncode
if run('01-checkout',['git','switch','--detach',pin]):raise RuntimeError('checkout failed')
run('02-revision',['git','rev-parse','HEAD'])
run('03-submodules',['git','submodule','status','--recursive'])
if run('04-clean',['git','diff','--exit-code',pin,'--']):raise RuntimeError('dirty source')
run('05-go-version',['go','version'])
run('06-clang-version',['clang','--version'])
run('07-node-version',['node','--version'])
(root/'rule_g_test.go.txt').write_bytes((work/'internal/lower/rule_g_test.go').read_bytes())
run('10-gofmt',['gofmt','-l','internal/lower/'])
run('11-focused',['go','test','-count=1','-run','TestAGenericTableCast|TestASetTimeout|TestAGenericMethod|TestAGenericFunction|TestAHostMethod','-v','./internal/lower/'])
path=work/'internal/lower/any_relations.go'; original=path.read_bytes()
a='func init() { refusalRules.register(3270, "refuseAnyRelation", (*lowering).refuseAnyRelation) }'
s=original.decode();assert s.count(a)==1
path.write_text(s.replace(a,'func init() {}'))
try:
 run('20-mutant-patch',['git','diff','--','internal/lower/any_relations.go'])
 run('21-mutant',['go','test','-count=1','-run','^TestAGenericTableCastThroughAWrittenAnyStaysNotYet$','-v','./internal/lower/'])
finally:
 path.write_bytes(original)
run('22-restoration',['git','diff','--exit-code',pin,'--'])
run('23-final-status',['git','status','--short'])
print(json.dumps({k:v['exit'] for k,v in results.items()},indent=2))
print(results['21-mutant']['stdout'],results['21-mutant']['stderr'])
