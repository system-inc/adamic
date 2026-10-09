exec(open('/tmp/u017-audit.py').read().split('metadata=[]; switch=base')[0])
# Reconstruct from the recorded switch, repair return syntax and overlapping match substitutions.
switch=(e/'switch.go.txt').read_text().replace('auditBool("M03", return true, return false)','return os.Getenv("ADAMIC_MUTANT") != "M03"')
a='path.Match(strings.ToLower(pattern), auditString("M02", strings.ToLower(path.Base(name)), path.Base(name)))'
b='path.Match(auditString("M15", strings.ToLower(pattern), strings.ToLower(path.Base(name))), auditString("M15", auditString("M02", strings.ToLower(path.Base(name)), path.Base(name)), strings.ToLower(pattern)))'
assert a in switch
switch=switch.replace(a,b)
env=os.environ.copy();env.update(GOWORK='off',ADAMIC_CSS_FIXTURES='/tmp/u017-prettier')
results=json.loads((e/'runs.json').read_text()); results=[x for x in results if not x['log'].endswith('.log') or x['log'].startswith('timing') or '-vet' in x['log']]
def run(cmd,log,extra={}):
 start=time.monotonic()
 with (e/log).open('w') as f:q=subprocess.run(cmd,cwd=r,env=env|extra,stdout=f,stderr=subprocess.STDOUT)
 result=dict(command=cmd,exit=q.returncode,wall=time.monotonic()-start,log=log);results.append(result);return q.returncode
try:
 p.write_text(switch)
 assert run(['gofmt','-w','internal/corpusfiles/files.go'],'switch-format.log')==0
 (e/'switch.go.txt').write_text(p.read_text())
 assert run(['go','test','-c','-o','/tmp/u017-test','./internal/corpusfiles/'],'switch-build.log')==0
 for mid in [m[0] for m in mutants]+[m[0] for m in probes]:
  run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/corpusfiles/','-run','.'],mid+'.log',dict(ADAMIC_MUTANT=mid))
finally:p.write_text(base);(e/'runs.json').write_text(json.dumps(results,indent=2))
for mid in [m[0] for m in mutants]+[m[0] for m in probes]:
 events=[json.loads(l) for l in (e/(mid+'.log')).read_text().splitlines() if l.startswith('{')]
 print(mid,[x.get('Test') for x in events if x.get('Action')=='fail' and x.get('Test') and '/' not in x['Test']])
