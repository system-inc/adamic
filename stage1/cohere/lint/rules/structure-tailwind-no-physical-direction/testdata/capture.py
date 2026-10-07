"""Capture pinned Go fixtures with options, preserving each original filename."""
import json, os, pathlib, subprocess, tempfile
here=pathlib.Path(__file__).resolve().parent
root=here.parents[5]
cohere=root/'cohere'
names={'structure/tailwind-no-physical-direction','@eslint-community/eslint-comments/require-description','@next/next/google-font-display'}
with tempfile.TemporaryDirectory(prefix='wave1-01-next-capture-') as temp:
 scratch=pathlib.Path(temp)
 harness=cohere/'internal/lint/testing/rule_testing.go'
 source=harness.read_text(); anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
 assert source.count(anchor)==1
 side=scratch/'harness.go';side.write_text(source.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'))
 overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(harness):str(side)}}))
 for family,pattern in [('core','RequireDescription'),('next','GoogleFontDisplay'),('tailwind','NoPhysicalDirection')]:
  with (here/('capture-'+family+'.log')).open('w') as log:
   subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+family,'-run','^Test'+pattern,'-count=1','-timeout=10m'],cwd=cohere,env=os.environ|{'COHERE_DOCS_CAPTURE':str(scratch/'capture')},stdout=log,stderr=subprocess.STDOUT,check=True)
 unique={}
 for file in (scratch/'capture').glob('*.jsonl'):
  for line in file.read_text().split('\n'):
   if not line:continue
   row=json.loads(line)
   if row['rule'] in names:
    key=(row['rule'],row['file'],row['source'],json.dumps(row.get('options'),sort_keys=True))
    unique[key]={k:row[k] for k in ['rule','file','source']}|{'options':row.get('options')}
 rows=[unique[key] for key in sorted(unique)]
 assert names=={r['rule'] for r in rows}
 (here/'cases.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
 print('captured',len(rows),'cases', {n:sum(r['rule']==n for r in rows) for n in sorted(names)})
