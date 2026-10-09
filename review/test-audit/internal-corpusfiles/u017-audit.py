import pathlib,subprocess,json,time,os,difflib
r=pathlib.Path('/workspace/adamic'); e=r/'review/test-audit/internal-corpusfiles'; p=r/'internal/corpusfiles/files.go'; base=p.read_text()
mutants=[
('M01',68,'strings.TrimSuffix(string(data), "\\x00")','string(data)','drop trailing NUL trim'),
('M02',73,'strings.ToLower(path.Base(name))','path.Base(name)','drop filename case folding'),
('M03',74,'return true','return false','change matching result constant'),
('M04',98,'actual != pin','actual == pin','flip pin comparison'),
('M05',107,'err != nil {\n\t\t\treturn nil, actual, nil, fmt.Errorf("%s: missing named root','false && err != nil {\n\t\t\treturn nil, actual, nil, fmt.Errorf("%s: missing named root','disable missing-root guard'),
('M06',111,'root == "."','root != "."','flip whole-repository pathspec condition'),
('M07',119,'len(status) != 0','false && len(status) != 0','disable upstream status guard'),
('M08',141,'len(dirty) != 0','false && len(dirty) != 0','disable dirty-file guard'),
('M09',163,'len(missing) != 0','false && len(missing) != 0','disable sparse missing-file guard'),
('M10',166,'counts[i] == 0','false && counts[i] == 0','disable empty-root guard'),
('M11',174,'sort.Strings(files)','sort.Sort(sort.Reverse(sort.StringSlice(files)))','change sorting option'),
('M12',160,'selected[absolute] = true','selected[absolute] = false','change map value constant'),
('M13',161,'counts[i]++','counts[i] += 2','off-by-one count increment'),
('M14',97,'strings.TrimSpace(string(head))','string(head)','drop HEAD whitespace trim'),
('M15',73,'path.Match(strings.ToLower(pattern), strings.ToLower(path.Base(name)))','path.Match(strings.ToLower(path.Base(name)), strings.ToLower(pattern))','swap glob arguments')]
probes=[('PRepository',23,'func Repository(t testing.TB, checkout string, roots, patterns []string) []string {','return nil'),('PUpstream',31,'func Upstream(t testing.TB, checkout, pin string, roots, patterns []string) []string {','return nil'),('PSelect',80,'func selectFiles(checkout, pin string, roots, patterns []string) ([]string, string, []int, error) {','return nil, "", make([]int, len(roots)), nil')]
metadata=[]; switch=base
for mid,line,old,new,desc in mutants:
 assert old in base,(mid,old)
 changed=base.replace(old,new,1)
 (e/(mid+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/internal/corpusfiles/files.go',tofile='b/internal/corpusfiles/files.go')))
 metadata.append(dict(id=mid,line=line,old=old,new=new,change=desc))
 # expression switches preserve original standalone edit exactly
 if mid=='M13': switched='counts[i] += 1; if os.Getenv("ADAMIC_MUTANT") == "M13" { counts[i]++ }'
 elif mid=='M12': switched='selected[absolute] = os.Getenv("ADAMIC_MUTANT") != "M12"'
 elif mid=='M11': switched='if os.Getenv("ADAMIC_MUTANT") == "M11" { '+new+' } else { '+old+' }'
 elif mid=='M05': switched=old.replace('err != nil {','os.Getenv("ADAMIC_MUTANT") != "M05" && err != nil {')
 elif mid in ['M07','M08','M09','M10']: switched='os.Getenv("ADAMIC_MUTANT") != "'+mid+'" && '+old
 else:
  # inline generic typed helpers declared below
  helper='auditString' if mid in ['M01','M02','M14'] else 'auditBool'
  switched=f'{helper}("{mid}", {old}, {new})'
 switch=switch.replace(old,switched,1)
for mid,line,old,ret in probes:
 switch=switch.replace(old,old+'\n if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { '+ret+' }',1)
 changed=base.replace(old,old+'\n '+ret,1)
 (e/(mid+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/internal/corpusfiles/files.go',tofile='b/internal/corpusfiles/files.go')))
switch+='\nfunc auditString(id, original, mutated string) string { if os.Getenv("ADAMIC_MUTANT") == id { return mutated }; return original }\nfunc auditBool(id string, original, mutated bool) bool { if os.Getenv("ADAMIC_MUTANT") == id { return mutated }; return original }\n'
(e/'menu.json').write_text(json.dumps(metadata,indent=2)); (e/'switch.go.txt').write_text(switch)
env=os.environ.copy();env.update(GOWORK='off',ADAMIC_CSS_FIXTURES='/tmp/u017-prettier')
def run(cmd,log,extra={}):
 start=time.monotonic()
 with (e/log).open('w') as f: q=subprocess.run(cmd,cwd=r,env=env|extra,stdout=f,stderr=subprocess.STDOUT)
 return dict(command=cmd,exit=q.returncode,wall=time.monotonic()-start,log=log)
results=[]
rows=[x for x in (e/'list.log').read_text().splitlines() if x.startswith('Test')]
for row in rows:
 for n in range(3): results.append(run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/corpusfiles/','-run','^'+row+'$'],f'timing-{row}-{n+1}.log'))
try:
 for mid,_,old,new,_ in mutants:
  p.write_text(base.replace(old,new,1)); results.append(run(['go','vet','./internal/corpusfiles/'],mid+'-vet.log'))
 p.write_text(switch);results.append(run(['gofmt','-w','internal/corpusfiles/files.go'],'switch-format.log'));results.append(run(['go','test','-c','-o','/tmp/u017-test','./internal/corpusfiles/'],'switch-build.log'))
 for mid in [m[0] for m in mutants]+[m[0] for m in probes]:
  results.append(run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/corpusfiles/','-run','.'],mid+'.log',dict(ADAMIC_MUTANT=mid)))
finally: p.write_text(base);(e/'runs.json').write_text(json.dumps(results,indent=2))
print(json.dumps(results,indent=2))
