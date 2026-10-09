import pathlib,subprocess,json,time,difflib,os,statistics
root=pathlib.Path('/workspace/adamic'); ev=pathlib.Path('/tmp/u116/evidence'); pkg='./stage1/cohere/lint/inventory/'; engine='stage1/cohere/lint/inventory/testdata/engine.go'; shard='stage1/cohere/lint/inventory/engine_shards_test.go'
base=(root/engine).read_text(); sb=(root/shard).read_text(); records=[]
def run(id,pattern='.',extra=[],env={}):
 c=['timeout','120','go','test','-json','-count=1','-timeout','90s',*extra,pkg,'-run',pattern]; start=time.monotonic()
 with (ev/(id+'.log')).open('w') as f: p=subprocess.run(c,cwd=root,env={**os.environ,**env},stdout=f,stderr=subprocess.STDOUT)
 rec={'id':id,'command':c,'environment':env,'wall':time.monotonic()-start,'exit':p.returncode}; records.append(rec); (ev/'runs.json').write_text(json.dumps(records,indent=2)); print(id,p.returncode,round(rec['wall'],3),flush=True)
 return rec
def diff(id,file,original,changed):
 (ev/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
rows={'family':'^TestInventoryEngine_[0-9]+$','union':'^TestInventoryEngineUnion$','assignment':'^TestInventoryEngineShardAssignment$','witness':'^TestInventoryEngineShardMutant$'}
for row,pat in rows.items():
 for i in range(3): run(f'timing-{row}-{i+1}',pat)
mut=[
 ('M1','Direct: depth == 0','Direct: depth != 0','Direct: (depth == 0 && os.Getenv("ADAMIC_MUTANT") != "M1") || (depth != 0 && os.Getenv("ADAMIC_MUTANT") == "M1")'),
 ('M2','Count: len(names)','Count: len(names) + 1','Count: len(names) + auditDelta("M2")'),
 ('M3','*report.Count++','*report.Count += 2','if os.Getenv("ADAMIC_MUTANT") == "M3" { *report.Count += 2 } else { *report.Count++ }'),
 ('M4','(?:enabled\\(','(?:disabled\\(',None),
 ('M5','if strings.HasPrefix(line, "ok  \\t"+rulesPrefix+family+"\\t") {','if !strings.HasPrefix(line, "ok  \\t"+rulesPrefix+family+"\\t") {',None),
 ('M6','string(file.FileName())+": "+fmt.Sprint(value)','filepath.Base(string(file.FileName()))+": "+fmt.Sprint(value)','auditPath(string(file.FileName()))+": "+fmt.Sprint(value)'),
 ('M7','if next != nil {\n\t\t\t\twalk(next, depth+1)','if false && next != nil {\n\t\t\t\twalk(next, depth+1)','if next != nil && os.Getenv("ADAMIC_MUTANT") != "M7" {\n\t\t\t\twalk(next, depth+1)')]
# Each target is selected from production functions, not test expected values.
switch=base; catalog=[]
for id,old,new,sw in mut:
 assert base.count(old)==1,(id,base.count(old))
 changed=base.replace(old,new,1); diff(id,engine,base,changed)
 catalog.append({'id':id,'file':engine,'line':base[:base.index(old)].count('\n')+1,'before':old,'after':new,'kind':'production'})
 if id=='M4':
  oldline=next(x for x in base.splitlines() if 'return regexp.MustCompile' in x); newline=oldline.replace(old,new)
  switch=switch.replace(oldline,'\tif os.Getenv("ADAMIC_MUTANT") == "M4" {\n'+newline+'\n\t}\n'+oldline)
 elif id=='M5': switch=switch.replace(old,'if strings.HasPrefix(line, "ok  \\t"+rulesPrefix+family+"\\t") != (os.Getenv("ADAMIC_MUTANT") == "M5") {')
 else: switch=switch.replace(old,sw,1)
switch+='\nfunc auditDelta(id string) int { if os.Getenv("ADAMIC_MUTANT") == id { return 1 }; return 0 }\nfunc auditPath(path string) string { if os.Getenv("ADAMIC_MUTANT") == "M6" { return filepath.Base(path) }; return path }\n'
probes=[('P1','func (a *analyzer) trace(root *declaration) ([]dependency, []string, bool, bool) {','return nil, nil, false, false'),('P2','func rankings(items []entry) []ranking {','return nil'),('P3','func countText(f frequency) string {','return ""'),('P4','func hasSelector(text, name string) bool {','return false'),('P5','func familyPassed(log []byte, family string) bool {','return false'),('P6','func measured(subject rule.Registration, file *tsast.SourceFile, graph *program.Graph, report *frequency) {','return'),('P7','func measureCorpus(root, output, label string, paths []string, items []entry) []string {','return nil')]
for id,anchor,ret in probes:
 added=anchor+'\n\tif os.Getenv("ADAMIC_MUTANT") == "'+id+'" { '+ret+' }'
 switch=switch.replace(anchor,added,1)
 diff(id,engine,base,base.replace(anchor,anchor+'\n\tif os.Getenv("ADAMIC_AUDIT_PROBE") == "'+id+'" { '+ret+' }',1))
 catalog.append({'id':id,'file':engine,'line':base[:base.index(anchor)].count('\n')+1,'before':anchor,'after':ret+' at entry (conditional probe)','kind':'probe'})
# countText probe precedes the witness build anchor. Restore that anchor only for witness's private built-in mutant, without weakening its comparison.
# Therefore witness failures for P3 are precondition failures, excluded from verdict.
(ev/'catalog.json').write_text(json.dumps(catalog,indent=2)); (ev/'switch.go.txt').write_text(switch)
(root/engine).write_text(switch)
try:
 for id,*_ in mut: run(id,env={'ADAMIC_MUTANT':id})
 for id,*_ in probes: run(id,env={'ADAMIC_MUTANT':id})
finally: (root/engine).write_text(base)
# Compile each standalone Go engine diff using its actual cohere overlay boundary.
for item in catalog:
 id=item['id']; target=ev/(id+'.go.txt'); changed=base
 if item['kind']=='production': changed=base.replace(item['before'],item['after'],1)
 else:
  anchor=item['before']; ret=next(p[2] for p in probes if p[0]==id); changed=base.replace(anchor,anchor+'\n\tif os.Getenv("ADAMIC_AUDIT_PROBE") == "'+id+'" { '+ret+' }',1)
 target.write_text(changed); overlay=ev/(id+'.overlay.json'); overlay.write_text(json.dumps({'Replace':{str(root/'cohere/adamic_inventory.go'):str(target),str(root/'cohere/adamic_inventory_test.go'):str(root/'stage1/cohere/lint/inventory/testdata/engine_test.go')}}))
 start=time.monotonic()
 with (ev/(id+'-vet.log')).open('w') as f: p=subprocess.run(['go','vet','-overlay='+str(overlay),str(root/'cohere/adamic_inventory.go'),str(root/'cohere/adamic_inventory_test.go')],cwd=root/'cohere',stdout=f,stderr=subprocess.STDOUT)
 records.append({'id':id+'-vet','exit':p.returncode,'wall':time.monotonic()-start}); (ev/'runs.json').write_text(json.dumps(records,indent=2)); print(id+' vet',p.returncode,flush=True)
# Allowed construction and weakened-check edits use private test overlays.
checks=[('S1','if seen[name] {','if !seen[name] {','setup'),('S2','\tTestInventoryEngine_007,\n','','setup'),('W1','return command.CombinedOutput()','output, _ := command.CombinedOutput()\n\treturn output, nil','witness')]
for id,old,new,kind in checks:
 assert sb.count(old)==1
 changed=sb.replace(old,new,1); diff(id,shard,sb,changed); catalog.append({'id':id,'file':shard,'line':sb[:sb.index(old)].count('\n')+1,'before':old,'after':new,'kind':kind})
 target=ev/(id+'.go.txt'); target.write_text(changed); overlay=ev/(id+'.overlay.json'); overlay.write_text(json.dumps({'Replace':{str(root/shard):str(target)}})); run(id,extra=['-overlay='+str(overlay)])
 with (ev/(id+'-vet.log')).open('w') as f: p=subprocess.run(['go','vet','-overlay='+str(overlay),pkg],cwd=root,stdout=f,stderr=subprocess.STDOUT)
 records.append({'id':id+'-vet','exit':p.returncode}); (ev/'runs.json').write_text(json.dumps(records,indent=2))
(ev/'catalog.json').write_text(json.dumps(catalog,indent=2))
