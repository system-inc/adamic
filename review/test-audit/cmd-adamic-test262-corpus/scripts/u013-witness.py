import pathlib,json,time,subprocess,os
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/cmd-adamic-test262-corpus';P=R/'cmd/adamic-test262';base=json.load(open(E/'base.json'));plans=json.load(open(E/'plan.json'))
while not (E/'probes.done').exists():time.sleep(1)
fixture='package main\nimport "testing"\nfunc TestAuditSurvivorEvidence(t *testing.T) { feature, rejected := unsupportedFeature("Proxy"); t.Logf("unsupportedFeature(Proxy) = %q, %v", feature, rejected); got:=classify("probe.js", "/*---\\nfeatures: [Proxy]\\n---*/\\nconst x = 1;", false);t.Logf("classify feature-only Proxy skip=%q",got.Skip) }\n'
(E/'survivor-witness.go.fixture').write_text(fixture);target=P/'audit_survivor_test.go';target.write_text(fixture)
try:
 for variant in ['clean','M02']:
  p=next(p for p in plans if p['id']=='M02');f=p['file'].split('/')[-1];s=base[f]
  if variant=='M02':(P/f).write_text(s[:p['pos']]+p['new']+s[p['pos']+len(p['old']):])
  with (E/('survivor-'+variant+'.log')).open('w') as log:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','^TestAuditSurvivorEvidence$'],cwd=R,stdout=log,stderr=subprocess.STDOUT)
  (P/f).write_text(s)
finally:target.unlink(missing_ok=True)
(E/'witness.done').write_text('done')
