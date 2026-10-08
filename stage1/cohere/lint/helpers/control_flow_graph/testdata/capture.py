#!/usr/bin/env python3
import json,os,subprocess,tempfile,pathlib,hashlib,gzip
here=pathlib.Path(__file__).resolve().parent;repo=here.parents[5];cohere=repo/'cohere';destination=pathlib.Path(os.environ.get('ADAMIC_CFG_FIXTURE_OUTPUT',str(here)));destination.mkdir(parents=True,exist_ok=True)
with tempfile.TemporaryDirectory(prefix='cfg-capture-') as td:
 temp=pathlib.Path(td);output=temp/'records';output.mkdir();replace={}
 for file,name in [('cfg.go','Build'),('roots.go','IndexRoots'),('paths.go','AnalyzePaths')]:
  path=cohere/'internal/lint/ecmascript/control_flow_graph'/file;s=path.read_text();anchor='func '+name;assert s.count(anchor)==1
  side=temp/file;side.write_text(s.replace(anchor,'func adamicOriginal'+name,1));replace[str(path)]=str(side)
 replace[str(cohere/'internal/lint/ecmascript/control_flow_graph/adamic_capture.go')]=str(here/'capture.go.txt')
 replace[str(cohere/'internal/lint/ecmascript/control_flow_graph/adamic_controls_test.go')]=str(here/'controls.go.txt')
 overlay=temp/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}));env=dict(os.environ,ADAMIC_CFG_CAPTURE=str(output))
 for pkg,tests in [('core','Test(ArrayCallbackReturn|ConsistentReturn|NoUnreachableLoop)'),('react','TestRulesOfHooks'),('../ecmascript/control_flow_graph','Test(IndexRoots|AdamicArenaControls|Path)')]:
  log=pathlib.Path('/tmp')/('cfg-capture-'+pkg.replace('/','-')+'.jsonl')
  with log.open('w') as f:p=subprocess.run(['go','test','-json','-overlay='+str(overlay),'./internal/lint/rules/'+pkg,'-run','^'+tests,'-count=1','-timeout=15m'],cwd=cohere,env=env,stdout=f,stderr=subprocess.STDOUT)
  if p.returncode:print(log.read_text()[-6000:]);raise SystemExit(p.returncode)
  results=[json.loads(l) for l in log.read_text().splitlines() if l.startswith('{')];assert not any(x.get('Action') in ['skip','fail'] for x in results),(pkg,'missing/failed cases')
 records={};calls={'build':0,'roots':0,'paths':0}
 for file in sorted(output.glob('*.jsonl')):
  for line in file.read_text().splitlines():
   d=json.loads(line);calls[d['op']]+=1;s=json.dumps(d,sort_keys=True,separators=(',',':'));records[s]=d
 cases=[records[s] for s in sorted(records)];(destination/'cases.json.gz').write_bytes(gzip.compress((json.dumps(cases,separators=(',',':'))+'\n').encode(),mtime=0));(destination/'capture_counts.json').write_text(json.dumps({'calls':calls,'distinct':len(cases)},indent=2)+'\n');print(calls,len(cases))
