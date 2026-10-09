exec(open('/tmp/u157.py').read().split('# Timings grouped')[0])
records=json.loads((p/'runs.json').read_text())
# Construction entries are themselves the subject of these preparation rows.
probes=[('P3','stage1/typescript/parser/whole_mutants_split_test.go','wholeMutantBuildOracleProduct','TestProduct_ParserOracle',['encoding/json']),('P4','stage1/typescript/parser/compiler_expressions_shards_test.go','compilerExpressionsLower','TestProduct_CompilerExpressionsLower',['context','github.com/system-inc/adamic/internal/load','github.com/system-inc/adamic/internal/lower']),('P5','stage1/typescript/parser/compiler_expressions_shards_test.go','compilerExpressionsNative','TestProduct_CompilerExpressionsNative',[]),('P6','stage1/typescript/parser/whole_mutants_split_test.go','wholeMutantLowerProduct','TestProduct_WholeMutantsLower family',['github.com/system-inc/adamic/internal/load','github.com/system-inc/adamic/internal/lower']),('P7','stage1/typescript/parser/whole_mutants_split_test.go','wholeMutantBuildPortProduct','TestProduct_WholeMutantsNative family',[]),('P8','stage1/typescript/parser/whole_mutants_setup_test.go','wholeMutantsPrepare','TestWholeMutants_Setup',[])]
for id,f,entry,row,remove in probes:
 s=(root/f).read_text();start=s.index('func '+entry+'(');brace=s.index('{',start);end=s.index('\n}',brace)
 body='\n\treturn\n}' if id=='P8' else '\n\treturn ""\n}'
 changed=s[:brace+1]+body+s[end+2:]
 for item in remove:changed=changed.replace('\n\t"'+item+'"','')
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 try:
  (root/f).write_text(changed);run(id+'-vet','timeout 90 go vet ./stage1/typescript/parser/')
  pat='^('+'|'.join(rows[row])+')$';run(id,f"timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '{pat}'",{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u157/cache/'+id})
 finally:(root/f).write_text(s)
 run(id+'-apply',f'git read-tree {base} && git apply --cached --check {p.relative_to(root)}/{id}.diff',{'GIT_INDEX_FILE':'/tmp/u157-index'})
run('coverage','timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run "^(TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees)$"',{'NODE_V8_COVERAGE':'/tmp/u157/coverage'})
profiles=[]
for path in pathlib.Path('/tmp/u157/coverage').glob('*.json'):
 for script in json.loads(path.read_text()).get('result',[]):
  if re.search(r'/(main|nodes|parser|statements|grammar|lookahead|jsx)\.ts$',script.get('url','')):profiles.append(dict(file=script['url'],functions=[f for f in script['functions'] if any(r['count']>0 for r in f['ranges'])]))
(p/'functions-reached.json').write_text(json.dumps(profiles,indent=2))
