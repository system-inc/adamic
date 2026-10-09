exec(open('/tmp/u148-run.py').read().split('for row in rows:')[0])
# Load preceding records, keeping their measured data.
records=json.loads((out/'runs.json').read_text())
# Repeat all qualifying samples and matrices after the corrected npm install and clean restored baseline.
for row in rows:
 for i in range(3):
  name=row+'-'+str(i+1)
  (out/(name+'.log')).rename(out/(name+'-initial.log'))
  run(name,f"timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run '^{row}$'")
for id,f,a,b,kind in menu:
 (out/(id+'.log')).rename(out/(id+'-initial.log'))
 try:
  (root/f).write_text(original[f].replace(a,b,1))
  run(id,'timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run .',{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u148/cache/replay-'+id})
 finally:(root/f).write_text(original[f])
f='internal/lower/lower.go';s=(root/f).read_text();start=s.index('func Lower(');end=s.index('\ntype lowering struct',start)
p=s[:start]+'func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\treturn &ir.Program{}, nil\n}\n'+s[end:]
p=p.replace('\n\t"fmt"','').replace('\n\t"path/filepath"','')
(out/'P2.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),p.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
try:
 (root/f).write_text(p);run('P2-vet','timeout 90 go vet ./internal/lower/');run('P2','timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run .',{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u148/cache/P2'})
finally:(root/f).write_text(s)
run('gap-coverage','timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./internal/lower,./internal/native -coverprofile=review/test-audit/stage1-cohere-values/gap.cover ./stage1/cohere/values/ -run "^TestEachGapStandsWhereGapsMdSaysItDoes$"')
run('port-coverage','timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run "^TestThePortParsesAsGoCohereDoes$/^natively"',{'NODE_V8_COVERAGE':'/tmp/u148/node-coverage'})
run('coverage-functions','go tool cover -func=review/test-audit/stage1-cohere-values/gap.cover')
profiles=[]
for path in pathlib.Path('/tmp/u148/node-coverage').glob('*.json'):
 for script in json.loads(path.read_text()).get('result',[]):
  if re.search(r'/(values|parser|tokenize|nodes|main)\.ts$',script.get('url','')):
   profiles.append(dict(file=script['url'],functions=[f for f in script['functions'] if any(r['count']>0 for r in f['ranges'])]))
(out/'port-reached.json').write_text(json.dumps(profiles,indent=2))
for diff in out.glob('*.diff'):
 extra={'GIT_INDEX_FILE':'/tmp/u148-index'}
 run(diff.stem+'-apply',f'git read-tree {base} && git apply --cached --check {diff.relative_to(root)}',extra)
run('final-restored','timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run .')
