import pathlib,subprocess,json,difflib,time,os
repo=pathlib.Path('/workspace/adamic');p=repo/'review/test-audit/stage1-cohere-typeaware-inspect_shards';tmp=pathlib.Path('/tmp/u145'); runs=[]
controls=[('S3','stage1/cohere/typeaware/suite_test.go','const testSixRuleAgreementAndMutantsShards = 95','const testSixRuleAgreementAndMutantsShards = 94','^TestSixRuleAgreementAndMutants_Setup$'),('W5','stage1/cohere/typeaware/profile_test.go','source = strings.Replace(source, from, to, 1)','source = strings.Replace(source, from, to, 1)\n source = strings.ReplaceAll(source, "?? panic(\'missing binding index\')", "?? []")','^TestShadowIndexMissingBinding_000$')]
for id,file,old,new,pattern in controls:
 original=(repo/file).read_text();assert original.count(old)==1;modified=original.replace(old,new,1);(p/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],cwd=repo,check=True)
 side=tmp/(id+'.go');side.write_text(modified);overlay=tmp/(id+'.json');overlay.write_text(json.dumps({'Replace':{str(repo/file):str(side)}}))
 with (p/(id+'-vet.log')).open('w') as f:r=subprocess.run(['go','vet','-overlay',str(overlay),'./stage1/cohere/typeaware/'],cwd=repo,stdout=f,stderr=subprocess.STDOUT)
 assert r.returncode==0
 cmd=['timeout','120','go','test','-overlay',str(overlay),'-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',pattern];start=time.monotonic()
 with (p/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/u145/typescript',ADAMIC_TYPEAWARE_BENCH='1'),stdout=f,stderr=subprocess.STDOUT)
 runs.append({'id':id,'file':file,'line':original[:original.index(old)].count('\n')+1,'change':new,'command':cmd,'seconds':time.monotonic()-start,'exit':r.returncode});(p/'extra-control-runs.json').write_text(json.dumps(runs,indent=2))
print('extra controls complete')
