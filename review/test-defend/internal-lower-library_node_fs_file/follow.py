import pathlib,json,subprocess,os,time,difflib
r=pathlib.Path('/workspace/adamic');o=r/'review/test-defend/internal-lower-library_node_fs_file';plans=json.loads((o/'plan.json').read_text());d=plans[1];f=r/d['file'];orig=f.read_text();names=[x for x in pathlib.Path('/tmp/defend-nodefs/list.log').read_text().splitlines() if x.startswith('Test') and x!=d['target']];runs=[]
def run(mid,regex):
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-nodefs/cache/'+mid;cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',regex];t=time.monotonic()
 with (o/(mid+'.log')).open('w') as out:rc=subprocess.run(cmd,cwd=r,env=env,stdout=out,stderr=subprocess.STDOUT).returncode
 es=[]
 for line in (o/(mid+'.log')).read_text().splitlines():
  try:es.append(json.loads(line))
  except ValueError:pass
 result={'id':mid,'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'wall_seconds':time.monotonic()-t,'exit':rc,'rows_failed':[e['Test'] for e in es if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']],'rows_passed':[e['Test'] for e in es if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']],'failures':[{'test':e.get('Test'),'output':e['Output'].strip()} for e in es if e.get('OutputType')=='error'],'package_seconds':[e.get('Elapsed') for e in es if e.get('Action') in ['fail','pass'] and 'Test' not in e]};runs.append(result);(o/'follow-results.json').write_text(json.dumps(runs,indent=2));print(mid,result['rows_failed'],len(result['rows_passed']),rc,flush=True)
try:
 f.write_text(orig.replace(d['old'],d['new']));run('Q1-rest','^('+'|'.join(names)+')$');run('Q1-alone','^'+d['target']+'$')
finally:f.write_text(orig)
f=r/'internal/lower/library_node.go';orig=f.read_text();old='''\t\t\tif (parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindInterfaceDeclaration) && parent.Name() != nil {''';new='''\t\t\tif parent.Kind == ast.KindClassDeclaration && len(parent.TypeParameters()) > 0 {
\t\t\t\treturn ""
\t\t\t}
'''+old
try:
 assert orig.count(old)==1;edited=orig.replace(old,new);f.write_text(edited)
 (o/'D2.diff').write_text(''.join(difflib.unified_diff(orig.splitlines(True),edited.splitlines(True),fromfile='a/internal/lower/library_node.go',tofile='b/internal/lower/library_node.go')))
 (o/'D2-plan.json').write_text(json.dumps({'id':'D2','file':'internal/lower/library_node.go','line':orig[:orig.index(old)].count('\n')+1,'kind':'return early','change':'return no recognized member for declarations inside generic classes'},indent=2))
 with (o/'D2-vet.log').open('w') as out:rc=subprocess.run(['go','vet','./internal/lower/'],cwd=r,stdout=out,stderr=subprocess.STDOUT).returncode
 assert rc==0;run('D2','.')
finally:f.write_text(orig)
