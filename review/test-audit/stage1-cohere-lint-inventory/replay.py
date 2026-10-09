import pathlib
exec(pathlib.Path('/tmp/u116/audit.py').read_text().split('rows=')[0])
records=json.loads((ev/'runs.json').read_text())
for rec in records:
 if rec['id'] in ['M'+str(i) for i in range(1,8)]+['P'+str(i) for i in range(1,8)]: rec['id']='preliminary-'+rec['id']
for id in ['M'+str(i) for i in range(1,8)]+['P'+str(i) for i in range(1,8)]:
 (ev/(id+'.log')).rename(ev/('preliminary-'+id+'.log'))
switch=(ev/'switch.go.txt').read_text()
p3='\n\tif os.Getenv("ADAMIC_MUTANT") == "P3" { return "" }'
switch=switch.replace(p3,'')
switch=switch.replace('auditPath(string(file.FileName()))+": "+fmt.Sprint(value)','string(file.FileName())+auditSeparator()+fmt.Sprint(value)')
switch+='\nfunc auditSeparator() string { if os.Getenv("ADAMIC_MUTANT") == "M6" { return "; " }; return ": " }\n'
(ev/'production-switch.go.txt').write_text(switch)
catalog=json.loads((ev/'catalog.json').read_text())
item=next(x for x in catalog if x['id']=='M6'); item['before']='+": "+fmt.Sprint(value)'; item['after']='+"; "+fmt.Sprint(value)'; item['line']=base[:base.index(item['before'])].count('\n')+1
diff('M6',engine,base,base.replace(item['before'],item['after'],1))
(ev/'M6.go.txt').write_text(base.replace(item['before'],item['after'],1))
(root/engine).write_text(switch)
try:
 run('switch-clean')
 for i in range(1,8): run('M'+str(i),env={'ADAMIC_MUTANT':'M'+str(i)})
 for i in range(1,8):
  if i==3: (root/engine).write_text(switch.replace('func countText(f frequency) string {','func countText(f frequency) string {'+p3))
  run('P'+str(i),env={'ADAMIC_MUTANT':'P'+str(i)})
  if i==3: (root/engine).write_text(switch)
finally: (root/engine).write_text(base)
with (ev/'M6-vet.log').open('w') as f: p=subprocess.run(['go','vet','-overlay='+str(ev/'M6.overlay.json'),str(root/'cohere/adamic_inventory.go'),str(root/'cohere/adamic_inventory_test.go')],cwd=root/'cohere',stdout=f,stderr=subprocess.STDOUT)
assert p.returncode==0
checks=[('S1','\n\t\tif seen[name] {','\n\t\tif !seen[name] {','setup'),('S2','\tTestInventoryEngine_007,\n','','setup'),('W1','return command.CombinedOutput()','output, _ := command.CombinedOutput()\n\treturn output, nil','witness')]
for id,old,new,kind in checks:
 assert sb.count(old)==1,(id,sb.count(old))
 changed=sb.replace(old,new,1); diff(id,shard,sb,changed); catalog.append({'id':id,'file':shard,'line':sb[:sb.index(old)].count('\n')+1,'before':old,'after':new,'kind':kind})
 target=ev/(id+'.go.txt'); target.write_text(changed); overlay=ev/(id+'.overlay.json'); overlay.write_text(json.dumps({'Replace':{str(root/shard):str(target)}})); run(id,extra=['-overlay='+str(overlay)])
 with (ev/(id+'-vet.log')).open('w') as f: p=subprocess.run(['go','vet','-overlay='+str(overlay),pkg],cwd=root,stdout=f,stderr=subprocess.STDOUT)
 records.append({'id':id+'-vet','exit':p.returncode}); (ev/'runs.json').write_text(json.dumps(records,indent=2)); assert p.returncode==0
(ev/'catalog.json').write_text(json.dumps(catalog,indent=2))
