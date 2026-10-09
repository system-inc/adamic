import pathlib,json,subprocess,difflib
root=pathlib.Path('.');p=root/'review/test-audit/stage1-cohere-formatfiles-formatfiles_products';base='stage1/cohere/formatfiles/'
changes=[
('M01','golang.ts',"character === 'İ' ? 'i' : character.toLowerCase()","character === 'İ' ? 'x' : character.toLowerCase()",'change constant'),
('M02','golang.ts','return leftPoint < rightPoint ? -1 : 1;','return leftPoint < rightPoint ? 1 : -1;','change constants'),
('M03','golang.ts','return path.slice(index);','return path.slice(index + 1);','off-by-one bound'),
('M04','enumerate.ts','walked = 0;','walked = 1;','change constant'),
('M05','enumerate.ts','enumeration.files.push(path);','/* dropped file push */','drop statement'),
('M06','disk.ts',"return fileStatus(path).kind === 'Ok';","return fileStatus(path).kind !== 'Ok';",'flip condition'),
('M07','main.ts',"const hexDigits = '0123456789abcdef';","const hexDigits = '0123456789ABCDEF';",'change constant'),
('M08','disk.ts','symbolicLink: status.symbolicLink };','symbolicLink: false };','change constant'),
('S01','formatfiles_prepare_test.go','prepared.binary = formatfilesBinary(t, prepared.program, applied)','/* dropped prepared binary assignment */','drop statement'),
('S03','formatfiles_shards_test.go','"./internal/format/formatfiles")','"./internal/format/formatfiles-missing")','change option')]
original={f:subprocess.check_output(['git','show','HEAD:'+base+f],text=True) for f in {x[1] for x in changes}|{'formatfiles_prepare_test.go','formatfiles_shards_test.go'}};switched=original.copy();plan=[]
def save(mid,file,old,new,kind='production',menu='return early'):
 source=original[file];assert source.count(old)==1,(mid,source.count(old));mut=source.replace(old,new,1);(p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(source.splitlines(True),mut.splitlines(True),fromfile='a/'+base+file,tofile='b/'+base+file)));line=source[:source.index(old)].count('\n')+1;plan.append(dict(id=mid,file=base+file,line=line,old=old,new=new,kind=kind,menu=menu));return source
for mid,f,old,new,menu in changes:
 save(mid,f,old,new,'setup-construction' if mid.startswith('S') else 'production',menu)
 if f.endswith('.ts'):
  if old.startswith('const hexDigits'):rep="const hexDigits = auditSelector === 'M07' ? '0123456789ABCDEF' : '0123456789abcdef';"
  elif old=='walked = 0;':rep="walked = auditSelector === 'M04' ? 1 : 0;"
  elif old.startswith('symbolicLink:'):rep="symbolicLink: auditSelector === 'M08' ? false : status.symbolicLink };"
  elif old.startswith('character ==='):rep="(auditSelector === 'M01' ? "+new+' : '+old+')'
  elif mid=='M05':rep="if (auditSelector !== 'M05') { "+old+' }'
  else:rep="if (auditSelector === '"+mid+"') { "+new+' }\n\t\t\t'+old
 else:
  if mid=='S01':rep='if os.Getenv("ADAMIC_MUTANT") != "S01" { '+old+' }'
  else:rep='auditOraclePackage())'
 switched[f]=switched[f].replace(old,rep,1)
f='formatfiles_prepare_test.go';source=original[f];a=source.index('formatfilesUnsanitizedOnce.Do(func() {');b=source.index('\n\t})',a);old=source[a:b+5];new='formatfilesUnsanitizedOnce.Do(func() {})\n';save('S02',f,old,new,'setup-construction','drop entire callback body');switched[f]=switched[f].replace(old,old.replace('func() {','func() {\n\t\tif os.Getenv("ADAMIC_MUTANT") == "S02" { return }',1),1)
f='formatfiles_shards_test.go';a=original[f].index('product := buildcache.Product(t, inputs, func(directory string) error {',original[f].index('func formatfilesOracle'));b=original[f].index('\n\t})',a);old=original[f][a:b+5];new='product := buildcache.Product(t, inputs, func(directory string) error { return nil })\n';save('S04',f,old,new,'setup-construction','drop entire build callback body');switched[f]=switched[f].replace('product := buildcache.Product(t, inputs, func(directory string) error {','product := buildcache.Product(t, inputs, func(directory string) error {\n\t\tif os.Getenv("ADAMIC_MUTANT") == "S04" { return nil }',1)
# Empty-answer probes target the construction entries actually called by the rows.
for mid,f,entry,empty in [('P02','formatfiles_shards_test.go','func formatfilesOracle(t *testing.T) (string, string) {','return "", ""'),('P03','formatfiles_prepare_test.go','func formatfilesPreparedPort(t *testing.T) formatfilesPrepared {','return formatfilesPrepared{}'),('P04','formatfiles_prepare_test.go','func formatfilesPreparedMutant(t *testing.T, index int) formatfilesPrepared {','return formatfilesPrepared{}'),('P05','formatfiles_prepare_test.go','func formatfilesUnsanitized(t *testing.T) string {','return ""')]:
 origin=original[f];a=origin.index(entry);b=origin.index('\n}',a)+2;save(mid,f,origin[a:b],entry+'\n\t'+empty+'\n}','probe');switched[f]=switched[f].replace(entry,entry+'\n\tif os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { '+empty+' }',1)
# Port execution probe: no records processed, hence no answer.
f='main.ts';old="for (const line of read.text.split('\\n')) {";new="for (const line of [] as string[]) {";save('P01',f,old,new,'probe');switched[f]=switched[f].replace(old,"for (const line of (auditSelector === 'P01' ? [] : read.text.split('\\n'))) {",1)
for f,source in switched.items():
 if f.endswith('.ts'):
  if f=='main.ts':anchor="const hexDigits =";i=source.index(anchor);prefix="const auditRead = readTextFile('/tmp/u091-selector');\nconst auditSelector = auditRead.kind === 'Ok' ? auditRead.text : '';\n\n";source=source[:i]+prefix+source[i:]
  else:source="import { readTextFile as auditReadTextFile } from 'adamic';\nconst auditRead = auditReadTextFile('/tmp/u091-selector');\nconst auditSelector = auditRead.kind === 'Ok' ? auditRead.text : '';\n"+source
 if f=='formatfiles_shards_test.go':source+='\nfunc auditOraclePackage() string { if os.Getenv("ADAMIC_MUTANT") == "S03" { return "./internal/format/formatfiles-missing" }; return "./internal/format/formatfiles" }\n'
 pathlib.Path(base+f).write_text(source)
(p/'plan.json').write_text(json.dumps(plan,indent=2));print([(x['id'],x['file'],x['line']) for x in plan])
