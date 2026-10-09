import pathlib,json,re,subprocess
D=pathlib.Path('/tmp/u103/evidence');plan=[]
def add(id,file,old,new,kind,rows):
 s=pathlib.Path(file).read_text();assert s.count(old)==1,(id,s.count(old));plan.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind,rows=rows))
ports=['TestPortMatchesGoCohere family','TestAdditionalJSONBoundaries','TestSingleFileStdoutDriver']
add('M01','stage1/cohere/json/formatter.ts',"base === 'package-lock.json'","base === 'lock.json'",'production',ports)
add('M02','stage1/cohere/json/parser.ts',"this.text.slice(0, this.position).split('\\n').length", "this.text.slice(0, this.position).split('\\n').length + 1",'production',ports)
add('M03','stage1/cohere/json/doc.ts','command.indent + 2','command.indent + 1','production',ports)
add('M04','stage1/cohere/json/width.ts','return text.length;','return text.length + 1;','production',ports)
add('M05','internal/childguard/childguard.go','if len(p) > 0 {','if len(p) == 0 {','production',['TestProgressGuard'])
add('H01','stage1/cohere/json/port_matches_prepared_split_test.go','portMatchesShared.ready = true','portMatchesShared.ready = false','setup',['TestPortMatchesGoCohereSplit_Setup'])
add('H02','stage1/cohere/json/port_matches_prepared_split_test.go','os.WriteFile(filepath.Join(dir, "main.c"),','os.WriteFile(filepath.Join(dir, "missing-main.c"),','setup',['TestProduct_JSON family'])
add('H03','stage1/cohere/json/repository_test.go','return pin, len(tracked), corpusPinError(expected, pin)','return pin, len(tracked) % 3, corpusPinError(expected, pin)','setup',['TestRepositoryLandingWithoutPinEdit'])
add('H04','stage1/cohere/json/repository_test.go','JSON corpus requires usable Git metadata: git %s: %w: %s','JSON corpus unavailable: git %s: %w: %s','setup',['TestRepositoryRequiresGit'])
add('W01','stage1/cohere/json/port_test.go','if string(result.stdout) == expected {','if true {','witness',['TestPortMatchesGoCohereSplitUnion'])
add('W02','stage1/cohere/json/port_test.go','if string(side.result.stdout) == expected {','if true {','witness',['TestThreePortMutantsAreCaught'])
add('W03','stage1/cohere/json/repository_test.go','return pin, len(tracked), corpusPinError(expected, pin)','return pin, len(tracked), nil','witness',['TestRepositoryCorpusMutants'])
(D/'plan.json').write_text(json.dumps(plan,indent=2))
lines=['Code under test: seven port source files (no Go cohere mutations), childguard Run/CombinedOutput/watched.Write, and separately suite constructions/witnessed checks.','Oracle: Go cohere executed independently for formatter output; self-written guard expectations, setup invariants and planted-check rejection counts.','Inventory lists every declaration in the reachable port modules and guard; all test-file helper declarations are separately included as a conservative construction call graph, not production mutants.']
for p in [*pathlib.Path('stage1/cohere/json').glob('*.ts'),pathlib.Path('internal/childguard/childguard.go'),*pathlib.Path('stage1/cohere/json').glob('*_test.go')]:
 for n,line in enumerate(p.read_text().splitlines(),1):
  if re.search(r'^func |function |^    \w+\(',line) and not re.search(r'^func Test',line):lines.append(f'{p}:{n}: {line.strip()}')
(D/'inventory.txt').write_text('\n'.join(lines)+'\n')
