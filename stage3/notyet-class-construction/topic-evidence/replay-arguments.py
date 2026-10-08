import csv, json, hashlib, pathlib, subprocess, os
root=pathlib.Path('/tmp/class-construction-census-source')
manifest=json.loads(subprocess.check_output(['git','show','e8c283b5:stage3/notyet-table/rerun-0730/source-manifest.json'],text=True))
for row in manifest:
 data=(root/row['file']).read_bytes()
 assert len(data)==row['bytes'] and hashlib.sha256(data).hexdigest()==row['sha256'],row['file']
print('Frozen morning source identity:',len(manifest),'files match',flush=True)
rs=list(csv.DictReader(open('/tmp/class-topic-roots.csv')))
types=['string | NodeArray<JSDocComment> | undefined','string | number | undefined','NodeArray<ParameterDeclaration> | readonly JSDocParameterTag[]','"boolean" | "list" | "listOrElement" | "number" | "object" | "string" | Map<string, string | number>','boolean | (() => boolean) | undefined','AnyBuildOrder | undefined','false | string[] | undefined','"boolean" | "list" | "number" | "object" | "string" | Map<string, string | number>','"boolean" | "number" | "object" | "string" | Map<string, string | number>','0 | boolean | undefined','boolean | (() => boolean)','string | false','string | false | undefined']
reasons=['a field of type '+t for t in types]+["a base that isn't a declared class"]
groups=[(reason,[r for r in rs if r['reason']==reason]) for reason in reasons]
groups.sort(key=lambda g:-len(g[1]))
results=[]
for i,(reason,group) in enumerate(groups,1):
 sample=group[0];where=str(root/sample['where']);stem=f'/tmp/class-topic-replay-{i:02d}'
 with open(stem+'.json','w') as out,open(stem+'.log','w') as err:
  run=subprocess.run(['/tmp/class-topic-replay','-project',str(root/'src/tsc/tsc.ts'),'-where',where,'-kind',sample['kind'],'-reason',reason],stdout=out,stderr=err,env=os.environ|{'LATENT_FULL':'1','LATENT_TRACE_WHERE':where})
 record=json.loads(pathlib.Path(stem+'.json').read_text())
 trace=pathlib.Path(stem+'.log').read_text()
 results.append(dict(reason=reason,root_sites=len(group),example=sample['where'],exit_code=run.returncode,trace=trace,findings=record['findings'],units=record['units']))
 print(i,len(group),run.returncode,sample['where'],flush=True)
pathlib.Path('/tmp/class-topic-replay-results.json').write_text(json.dumps(results,indent=2)+'\n')
