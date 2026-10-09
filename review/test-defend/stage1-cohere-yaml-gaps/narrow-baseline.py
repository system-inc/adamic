exec(open('/tmp/defend-yaml/baseline.py').read().split('groups=')[0])
g=json.loads((p/'groups.json').read_text());g.pop('unist');g.pop('widths-speed');groups={'unist-control':['TestUnistMatchesGo'],'widths':['TestWidthsMatchGo']}
for sub in ['comment_prefix_lost','point_column_shifted','document_end_marker_lost']:groups['unist-'+sub]=['TestUnistMutants/'+sub]
for sub in ['stringUnitScan.ts','arenaNodeRead.ts','numericMapLookup.ts']:groups['speed-'+sub]=['TestSpeedCostProbes/'+sub]
g.update(groups);(p/'groups.json').write_text(json.dumps(g,indent=2))
results=[]
for name,rows in groups.items():
 row=rows[0];pattern='^'+row.replace('/','$/^')+'$'
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',pattern];start=time.monotonic()
 with (p/('baseline-'+name+'.log')).open('w') as f:r=subprocess.run(cmd,cwd='/workspace/adamic',env=dict(os.environ,ADAMIC_YAML_LIBRARY='/tmp/u152/library'),stdout=f,stderr=subprocess.STDOUT)
 results.append(dict(group=name,command=cmd,exit=r.returncode,wall=time.monotonic()-start));(p/'baseline-narrow-groups.json').write_text(json.dumps(results,indent=2));print(results[-1],flush=True)
