"""Overlay only: observe every actual helper use in all consuming upstream suites."""
import json, os, re, subprocess, tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[5]
COHERE=ROOT/'cohere'
import sys
sys.path.insert(0, str(ROOT/'stage1/cohere/lint/helpers/testdata'))
from pin import capture_pin
PIN=capture_pin(ROOT)
HELPERS=['appendNodeSignature','appendTokensBetween','bodyDefinitelyExits','consistentReturnIsGenerator','hasSameTokens','idDenylistIsDestructuringTarget','idDenylistIsImportAttributeKey','idDenylistIsImportOptionsObject','isEmptyBracketLiteral','isNullOrUndefined','isSeparateEvaluationContext','memberAccessObject','noRestrictedExportsNameText','numericLiteralSign','switchStatementExits','tokenSignature','tryStatementExits']
with tempfile.TemporaryDirectory(prefix='core-capture-') as td:
 td=Path(td); capture=td/'capture';capture.mkdir();replace={}
 for src in (COHERE/'internal/lint/rules/core').glob('*.go'):
  if src.name.endswith('_test.go'):continue
  text=src.read_text();changed=False
  for name in HELPERS:
   marker='func '+name+'('
   if marker not in text:continue
   changed=True;text=text.replace(marker,'func adamicOriginal_'+name+'(',1)
  if changed:
   dest=td/src.name;dest.write_text(text);replace[str(src)]=str(dest)
 text=(HERE/'capture.go').read_text()
 for name in HELPERS:
  if name=='numericLiteralSign':
   sig='text string';rtype='int';params='text';sf='nil';args='nil';extra='text';body=f'result:=adamicOriginal_{name}({params})'
  elif name=='appendTokensBetween':
   sig='sourceFile *ast.SourceFile, start int, end int, signature *strings.Builder';rtype='';params='sourceFile,start,end,signature';sf='sourceFile';args='nil';extra='[]int{start,end}';body=f'before:=signature.Len();adamicOriginal_{name}({params});result:=signature.String()[before:]'
  elif name=='appendNodeSignature':
   sig='sourceFile *ast.SourceFile, node *ast.Node, signature *strings.Builder';rtype='';params='sourceFile,node,signature';sf='sourceFile';args='[]*ast.Node{node}';extra='nil';body=f'before:=signature.Len();adamicOriginal_{name}({params});result:=signature.String()[before:]'
  elif name in ['hasSameTokens','tokenSignature']:
   sig='sourceFile *ast.SourceFile, '+('left *ast.Node, right *ast.Node' if name=='hasSameTokens' else 'node *ast.Node');rtype='bool' if name=='hasSameTokens' else 'string';params='sourceFile,'+('left,right' if name=='hasSameTokens' else 'node');sf='sourceFile';args='[]*ast.Node{'+('left,right' if name=='hasSameTokens' else 'node')+'}';extra='nil';body=f'result:=adamicOriginal_{name}({params})'
  elif name in ['tryStatementExits','switchStatementExits']:
   sig='statement *ast.'+('TryStatement' if name=='tryStatementExits' else 'SwitchStatement');rtype='bool';params='statement';sf='nil';args='[]*ast.Node{statement.AsNode()}';extra='nil';body=f'result:=adamicOriginal_{name}({params})'
  else:
   sig='node *ast.Node';rtype='*ast.Node' if name=='memberAccessObject' else 'string' if name=='noRestrictedExportsNameText' else 'bool';params='node';sf='nil';args='[]*ast.Node{node}';extra='nil';body=f'result:=adamicOriginal_{name}({params})'
  text+=f'\nfunc {name}({sig}) {rtype} {{ {body};adamicCapture("{name}",{sf},{args},{extra},result);'+('return result' if rtype else '')+'}\n'
 exports=td/'capture.go';exports.write_text(text)
 replace[str(COHERE/'internal/lint/rules/core/adamic_capture.go')]=str(exports)
 overlay=td/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
 pattern='ConstructorSuper|IdLength|IdDenylist|NoConstantCondition|NoDupeElseIf|NoEmptyFunction|NoRestrictedImports|NoThisBeforeSuper|NoUnreachableLoop|NoUselessCall|PreferSpread|ForDirection|GetterReturn|ConsistentReturn|NoRestrictedExports'
 subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/core','-run','^Test('+pattern+')','-count=1','-timeout=10m'],cwd=COHERE,env=os.environ|{'ADAMIC_CORE_CAPTURE':str(capture)},check=True)
 calls=[json.loads(x) for x in (capture/'calls.jsonl').read_text().splitlines()]
 used=set(c['helper'] for c in calls);assert used==set(HELPERS),set(HELPERS)-used
 frames={key:json.loads((capture/(key+'.json')).read_text()) for key in sorted(set(c['frame'] for c in calls))}
 out={'pin':PIN,'frames':frames,'calls':calls}
 import gzip
 with gzip.open(HERE/'captures.json.gz','wt') as f:json.dump(out,f,separators=(',',':'))
 print('actual Go calls:',len(calls),'frames:',len(frames))
 for h in HELPERS: print(h,sum(c['helper']==h for c in calls))

metadata=json.loads((HERE/'helpers.json').read_text())
metadata['pin']=PIN
for row in metadata['helpers']:
  row['go_calls']=sum(c['helper']==row['symbol'].split('.')[-1] for c in calls)
(HERE/'helpers.json').write_text(json.dumps(metadata,indent=2)+'\n')
