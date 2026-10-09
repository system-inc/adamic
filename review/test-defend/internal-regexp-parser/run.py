import pathlib,subprocess,os,time,json,difflib
root=pathlib.Path('/workspace/adamic'); p=pathlib.Path('/tmp/def-regexp'); src=root/'internal/regexp/parser.go'; original=src.read_text()
menu=[
('D01','TestParse',"\t\t\tif _, rangeOperand := operand.(*ClassRange); rangeOperand {\n\t\t\t\treturn nil, p.fail(\"range must be nested in Unicode set operation\")\n\t\t\t}\n",'', 'drop left range-operand rejection'),
('D02','TestParse',"\t\t\t\tif _, rangeOperand := right.(*ClassRange); rangeOperand {\n\t\t\t\t\treturn nil, p.fail(\"range must be nested in Unicode set operation\")\n\t\t\t\t}\n",'', 'drop right range-operand rejection'),
('D03','TestParse','\tif neg && classMayContainStrings(expr) {\n\t\treturn nil, p.fail("cannot negate a class containing strings")\n\t}\n','', 'drop negated string-containing class rejection'),
('D04','TestFlags','\t\tseen[c] = true','\t\tseen[c] = false','change duplicate tracking constant'),
('D05','TestFlags','\tif f.Unicode && f.UnicodeSets {\n\t\treturn f, &SyntaxError{0, "u and v flags are mutually exclusive"}\n\t}\n','','drop u/v exclusion'),
('D06','TestFlags','\t\tdefault:\n\t\t\treturn f, &SyntaxError{i, "invalid flag"}','\t\tdefault:\n\t\t\treturn f, nil','return early with success for unknown flags'),
('D07','TestQuantifierBounds','\t\t\t\tmax = n','\t\t\t\tmax = min','change finite upper bound to lower bound'),
('D08','TestQuantifierBounds','Min: min, Max: max','Min: max, Max: min','swap minimum and maximum arguments'),
('D09','TestQuantifierBounds',"greedy := !p.take('?')","greedy := p.take('?')",'flip greediness condition'),
]
# D07 drop whole assignment and bind decimal result to blank to avoid unused n.
menu[6]=('D07','TestQuantifierBounds','if n, ok := p.decimal(); ok {\n\t\t\t\tmax = n\n\t\t\t}','if _, ok := p.decimal(); ok {\n\t\t\t}', 'drop finite upper-bound assignment, discard now-unused decimal result')
(p/'menu.json').write_text(json.dumps([{'id':i,'target':t,'line':original[:original.index(a)].count('\n')+1,'change':c} for i,t,a,b,c in menu],indent=2))
results=[]
try:
 for i,t,a,b,c in menu:
  assert original.count(a)==1,(i,original.count(a))
  modified=original.replace(a,b); src.write_text(modified)
  subprocess.run(['gofmt','-w',str(src)],check=True)
  modified=src.read_text(); (p/f'{i}.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/internal/regexp/parser.go',tofile='b/internal/regexp/parser.go')))
  with (p/f'{i}-vet.log').open('w') as f: subprocess.run(['go','vet','./internal/regexp/'],cwd=root,stdout=f,stderr=subprocess.STDOUT,check=True)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=f'/tmp/def-regexp/cache/{i}'
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run','.'];start=time.monotonic()
  with (p/f'{i}.log').open('w') as f: r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for line in (p/f'{i}.log').read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  failed=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test'in e and '/'not in e['Test']]
  passed=[e['Test'] for e in events if e.get('Action')=='pass' and 'Test'in e and '/'not in e['Test']]
  errors=[e.get('Output','').strip() for e in events if e.get('Test')==t and 'parser_test.go:'in e.get('Output','')]
  row={'id':i,'target':t,'change':c,'line':original[:original.index(a)].count('\n')+1,'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+' > '+i+'.log 2>&1','exit':r.returncode,'wall_seconds':time.monotonic()-start,'failed':failed,'passed':passed,'target_evidence':errors}
  results.append(row);(p/'matrix.json').write_text(json.dumps(results,indent=2));print(i,failed,flush=True)
finally:src.write_text(original)
