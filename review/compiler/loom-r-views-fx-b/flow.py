import json
from verify import run,root,out
paths=sorted((root/'internal/lower/testdata/fx6').glob('*.a'))+sorted((root/'internal/lower/testdata/scalar_union_views').glob('*.a'))
text='package flow\nimport "testing"\n'
for p in paths:
 name='TestRebasedViews_'+p.parent.name+'_'+p.stem
 text+='func '+name+'(t *testing.T) { t.Parallel(); checkAllFlowProgram(t, "../lower/testdata/'+p.parent.name+'/'+p.name+'") }\n'
p=out/'flow-extra_test.go.txt';p.write_text(text)
overlay=out/'flow-overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/'internal/flow/rebased_views_test.go'):str(p)}}))
raise SystemExit(run('flow',['go','test','-overlay',str(overlay),'./internal/flow','-run','^TestRebasedViews_|^TestFlowCorpus|^TestFlowProgram.*(view|predicate|interface|union)','-count=1','-json','-timeout','180s'],240))
