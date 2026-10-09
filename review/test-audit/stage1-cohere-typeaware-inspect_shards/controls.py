import pathlib,json,subprocess,difflib,time,os
repo=pathlib.Path('/workspace/adamic');p=repo/'review/test-audit/stage1-cohere-typeaware-inspect_shards';tmp=pathlib.Path('/tmp/u145'); base='stage1/cohere/typeaware/'
controls=[
('S1',base+'shared_test.go','products.LoadOrStore(key, &product{})','products.LoadOrStore("one-key", &product{})','products.LoadOrStore(func() string { if os.Getenv("ADAMIC_AUDIT_CONTROL") == "S1" { return "one-key" }; return key }(), &product{})','construction: change cache key'),
('S2',base+'six_builds_test.go','args[i+1] = filepath.Join(dir, file)','args[i+1] = command.Args[i+2]','if os.Getenv("ADAMIC_AUDIT_CONTROL") != "S2" { args[i+1] = filepath.Join(dir, file) }','construction: drop output redirection'),
('W1',base+'six_shards_test.go','func sixUnion(expected []string, shards []sixShard) error {','func sixUnion(expected []string, shards []sixShard) error {\n if len(expected) >= 0 { return nil }','func sixUnion(expected []string, shards []sixShard) error {\n if os.Getenv("ADAMIC_AUDIT_CONTROL") == "W1" { return nil }','witness: disable union rejection'),
('W2',base+'six_shards_test.go','if !bytes.Equal(want, got) {','if false {','if os.Getenv("ADAMIC_AUDIT_CONTROL") != "W2" && !bytes.Equal(want, got) {','witness: disable byte comparison'),
('W3',base+'typeaware_budget_test.go','\t\ttypeAwareKillCommandGroups()','\t\t// Command groups retained.','\t\tif os.Getenv("ADAMIC_AUDIT_CONTROL") != "W3" { typeAwareKillCommandGroups() }','witness: drop descendant kill'),
('W4',base+'typeaware_commands_test.go','syscall.Kill(-command.Process.Pid, syscall.SIGKILL)','syscall.Kill(command.Process.Pid, syscall.SIGKILL)','syscall.Kill(func() int { if os.Getenv("ADAMIC_AUDIT_CONTROL") == "W4" { return command.Process.Pid }; return -command.Process.Pid }(), syscall.SIGKILL)','witness: weaken group kill to leader kill')]
manifest=[];switched={}
for id,file,old,new,switch,description in controls:
 original=(repo/file).read_text();assert original.count(old)==1,(id,original.count(old));modified=original.replace(old,new,1)
 diff=''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+file,tofile='b/'+file))
 (p/(id+'.diff')).write_text(diff);subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],cwd=repo,check=True)
 standalone=tmp/(id+'.go');standalone.write_text(modified);overlay=tmp/(id+'.json');overlay.write_text(json.dumps({'Replace':{str(repo/file):str(standalone)}}))
 with (p/(id+'-vet.log')).open('w') as f:r=subprocess.run(['go','vet','-overlay',str(overlay),'./stage1/cohere/typeaware/'],cwd=repo,stdout=f,stderr=subprocess.STDOUT)
 assert r.returncode==0,id
 switched[file]=switched.get(file,original).replace(old,switch,1)
 manifest.append({'id':id,'file':file,'line':original[:original.index(old)].count('\n')+1,'change':new,'description':description})
replace={}
for file,s in switched.items():
 target=tmp/('switch-'+pathlib.Path(file).name);target.write_text(s);replace[str(repo/file)]=str(target)
overlay=tmp/'controls-overlay.json';overlay.write_text(json.dumps({'Replace':replace}));(p/'controls.json').write_text(json.dumps(manifest,indent=2))
rows=['TestSharedProductPublication','TestSixBuildCallbackUsesProductDirectory','TestSixShardUnionRejectsLossAndDuplication','TestSixShardPlantedDisagreement','TestTypeAwareUnitDeadline','TestTypeAwareContextKillsCompilerGroup','TestSixPinnedFlags','TestPinnedTypeFlags','TestTypeAwareNativeBuildChild'];pattern='^('+'|'.join(rows)+')$';runs=[]
for id,*_ in controls:
 cmd=['timeout','120','go','test','-overlay',str(overlay),'-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run',pattern];start=time.monotonic()
 with (p/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=repo,env=dict(os.environ,ADAMIC_AUDIT_CONTROL=id),stdout=f,stderr=subprocess.STDOUT)
 runs.append({'id':id,'command':cmd,'seconds':time.monotonic()-start,'exit':r.returncode,'matrix_rows':rows})
(p/'control-runs.json').write_text(json.dumps(runs,indent=2)); print('controls complete')
