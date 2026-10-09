import pathlib,json,subprocess,os,time,shlex,difflib
p=pathlib.Path('/tmp/u099/evidence');root=pathlib.Path.cwd();files=['stage1/cohere/graphql/printer/'+f for f in ['shards_test.go','preflight_units_test.go','gaps_test.go']];base={f:pathlib.Path(f).read_text() for f in files};rules=[
('W01',files[1],'if counts.unexpected != 0 {','if counts.unexpected < 0 {','TestPrinterPreflightPlantedDisagreement','witness weakening'),
('W02',files[0],'if string(result.stdout) == want.String() {','if result.exitCode == 0 {','TestPrinterShardPlantedDisagreement','witness weakening'),
('W03',files[0],'item != original','item.id != original.id','TestPrinterShardUnionRejectsMissingAndRepeated','construction comparison weakening'),
('S01',files[0],'selected[i] = i%boxes == index','selected[i] = i%boxes != index','TestPrinterShardSelection','construction condition flip'),
('S02',files[1],'\t\twhole = append(whole, enumeration...)\n','','TestPrinterUpstreamPreflight_Setup','construction dropped statement'),
('S03',files[2],'for _, item := range cases {\n\t\towner := whitespaceOwner(item.id)','for _, item := range cases[:len(cases)-1] {\n\t\towner := whitespaceOwner(item.id)','TestPrinterWhitespaceGap_Setup','construction off-by-one'),
('P02',files[0],'func printerShardUnion(whole []printerCase, shards []printerShard) error {','func printerShardUnion(whole []printerCase, shards []printerShard) error {\n\tif os.Getenv("ADAMIC_AUDIT_PROBE") == "P02" { return nil }','TestPrinterShardUnionRejectsMissingAndRepeated','empty construction entry probe'),
('P03',files[0],'func printerShardSelection(value string, count int) ([]bool, error) {','func printerShardSelection(value string, count int) ([]bool, error) {\n\tif os.Getenv("ADAMIC_AUDIT_PROBE") == "P03" { return nil, nil }','TestPrinterShardSelection','empty construction entry probe'),
('P04',files[0],'func enumeratePrinter(t *testing.T, mode, path, want string) []printerCase {','func enumeratePrinter(t *testing.T, mode, path, want string) []printerCase {\n\tif os.Getenv("ADAMIC_AUDIT_PROBE") == "P04" { return nil }','TestPrinterUpstreamPreflight_Setup','empty construction entry probe'),
('P05',files[2],'func whitespaceShards(cases []printerCase) []printerShard {','func whitespaceShards(cases []printerCase) []printerShard {\n\tif os.Getenv("ADAMIC_AUDIT_PROBE") == "P05" { return nil }','TestPrinterWhitespaceGap_Setup','empty construction entry probe')]
meta=[];env=os.environ.copy();env['ADAMIC_GRAPHQL_PRETTIER']='/tmp/u099-prettier'
for id,f,a,b,test,kind in rules:
 s=base[f];assert s.count(a)==1,(id,s.count(a));new=s.replace(a,b,1);source=pathlib.Path('/tmp/u099-'+id+'.go');source.write_text(new);overlay=pathlib.Path('/tmp/u099-'+id+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/f):str(source)}}));(p/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)));env['ADAMIC_AUDIT_PROBE']=id
 cmd=['timeout','120','go','test','-overlay='+str(overlay),'-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run','^'+test+'$'];start=time.monotonic()
 with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 d={'id':id,'kind':kind,'file':f,'line':s[:s.index(a)].count('\n')+1,'before':a,'after':b,'test':test,'command':'ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier ADAMIC_AUDIT_PROBE='+id+' '+shlex.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start};(p/(id+'-run.json')).write_text(json.dumps(d,indent=2));meta.append(d)
 with (p/(id+'-vet.log')).open('w') as log:v=subprocess.run(['timeout','90','go','vet','-overlay='+str(overlay),'./stage1/cohere/graphql/printer/'],env=env,stdout=log,stderr=subprocess.STDOUT)
 d['vet_exit']=v.returncode;(p/'harness-validation.json').write_text(json.dumps(meta,indent=2));print(id,'test',r.returncode,'vet',v.returncode,flush=True)
