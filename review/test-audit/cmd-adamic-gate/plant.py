from pathlib import Path
import json,difflib,subprocess
p=Path('review/test-audit/cmd-adamic-gate');scope=json.loads((p/'scope.json').read_text());base=scope['base'];main='cmd/adamic-gate/main.go';resume='cmd/adamic-gate/resume.go';test='cmd/adamic-gate/main_test.go';orig={f:subprocess.check_output(["git","show",base+":"+f],text=True) for f in [main,resume,test]}
plan=[
('M01',main,'names[i] = regexp.QuoteMeta(names[i])','names[i] = names[i]','drop leaf escaping'),
('M02',main,'regexp.QuoteMeta(parts[i])','parts[i]','drop prefix escaping'),
('M03',main,'last := "^(" + strings.Join(names, "|") + ")$"','last := "^(" + strings.Join(names, "|") + ")"','change leaf end-anchor constant'),
('M04',main,'parts[i] = "^" + regexp.QuoteMeta(parts[i]) + "$"','parts[i] = "^" + regexp.QuoteMeta(parts[i])','change parent end-anchor constant'),
('M05',main,'seenTest[r.key()] = index','','drop seen-test update'),
('M06',main,'if u.Shard != index {','if u.Shard == index {','flip shard mismatch condition'),
('M07',main,'(r.Test == u.Test || (u.Test != "" && strings.HasPrefix(r.Test, u.Test+"/")))','(r.Test == u.Test)','drop descendant ownership condition'),
('M08',main,'strings.HasPrefix(u.Test, r.Test+"/")','false','change ancestor acceptance to false'),
('M09',main,'outputs[key] += e.Output','','drop output accumulation'),
('M10',main,'if action != "skip" {','if action == "skip" {','flip skip reason condition'),
('M11',main,'raw++','raw += 2','off-by-one terminal counter'),
('M12',resume,'}{s, pkg, patterns, []string{','}{s, "", patterns, []string{','change checkpoint package input to empty'),
('M13',resume,'}{env, tools, external})','}{nil, tools, external})','drop environment input'),
('M14',resume,'uint32(info.Mode()), stat.Uid, stat.Gid, data','uint32(0), stat.Uid, stat.Gid, data','change file mode input to zero'),
('M15',resume,'if packageTerminals != len(e.Invocations) {','if packageTerminals == len(e.Invocations) {','flip package terminal count condition'),
('M16',resume,'if log != e.LogDigest || stderr != e.StderrDigest {','if log != e.LogDigest && stderr != e.StderrDigest {','flip checksum OR to AND'),
('M17',resume,'if residual > 0 {','if residual < 0 {','flip positive parent residual condition'),
('M18',main,'if !ok || f.Name.Name != parent {','if !ok || f.Name.Name == parent {','flip AST parent-name condition'),
('M19',resume,'!strings.HasPrefix(filepath.Base(absolute), "package-")','strings.HasPrefix(filepath.Base(absolute), "package-")','flip owned scratch-name condition'),
('S01',test,'func TestAnchoredSelectors(t *testing.T) {\n\tt.Parallel()','func TestAnchoredSelectors(t *testing.T) {','drop Parallel statement from suite construction')]
# Self-assignment standalone M01 would fail vet: omit the whole loop instead.
plan[0]=('M01',main,'\t\tfor i := range names {\n\t\t\tnames[i] = regexp.QuoteMeta(names[i])\n\t\t}','','drop entire leaf escaping loop')
mutations=[];switched=dict(orig)
for id,f,b,a,description in plan:
 assert b in orig[f],id
 changed=orig[f].replace(b,a,1)
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(orig[f].splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 mutations.append({'id':id,'file':f,'line':orig[f][:orig[f].index(b)].count('\n')+1,'before':b,'after':a,'description':description,'kind':'setup' if id=='S01' else 'production'})
 # Expression mutants compose where later targets nest within earlier ones.
 if id=='S01':replacement=b
 elif id=='M01':replacement='if os.Getenv("ADAMIC_MUTANT") != "M01" {\n'+b+'\n}'
 elif id=='M02':replacement='auditString("M02", regexp.QuoteMeta(parts[i]), parts[i])'
 elif id=='M04':
  # Earlier M02 has already rewritten the shared escaping expression.
  bsw='parts[i] = "^" + auditString("M02", regexp.QuoteMeta(parts[i]), parts[i]) + "$"'
  switched[f]=switched[f].replace(bsw,'parts[i] = "^" + auditString("M02", regexp.QuoteMeta(parts[i]), parts[i]) + auditString("M04", "$", "")',1);continue
 elif id=='M03':replacement='last := "^(" + strings.Join(names, "|") + auditString("M03", ")$", ")")'
 elif id=='M05' or id=='M09':replacement='if os.Getenv("ADAMIC_MUTANT") != "'+id+'" { '+b+' }'
 elif id=='M06':replacement='if auditBool("M06", u.Shard != index, u.Shard == index) {'
 elif id=='M07':replacement='(r.Test == u.Test || (os.Getenv("ADAMIC_MUTANT") != "M07" && u.Test != "" && strings.HasPrefix(r.Test, u.Test+"/")))'
 elif id=='M08':replacement='(os.Getenv("ADAMIC_MUTANT") != "M08" && '+b+')'
 elif id=='M10':replacement='if auditBool("M10", action != "skip", action == "skip") {'
 elif id=='M11':replacement='raw++; if os.Getenv("ADAMIC_MUTANT") == "M11" { raw++ }'
 elif id=='M12':replacement='}{s, auditString("M12", pkg, ""), patterns, []string{'
 elif id=='M13':replacement='}{auditEnv(env), tools, external})'
 elif id=='M14':replacement='auditMode(uint32(info.Mode())), stat.Uid, stat.Gid, data'
 elif id=='M15':replacement='if auditBool("M15", packageTerminals != len(e.Invocations), packageTerminals == len(e.Invocations)) {'
 elif id=='M16':replacement='if auditBool("M16", log != e.LogDigest || stderr != e.StderrDigest, log != e.LogDigest && stderr != e.StderrDigest) {'
 elif id=='M17':replacement='if auditBool("M17", residual > 0, residual < 0) {'
 elif id=='M18':replacement='if !ok || auditBool("M18", f.Name.Name != parent, f.Name.Name == parent) {'
 elif id=='M19':replacement='auditBool("M19", !strings.HasPrefix(filepath.Base(absolute), "package-"), strings.HasPrefix(filepath.Base(absolute), "package-"))'
 switched[f]=switched[f].replace(b,replacement,1)
probes=[
('P01',main,'patterns','return nil'),('P02',main,'validateResults','return nil'),('P03',main,'readLog','return nil, nil, 0, nil'),('P04',resume,'checkpointKey','return ""'),('P05',resume,'pathDigest','return "", nil'),('P06',resume,'executionKey','return ""'),('P07',resume,'stableGoEnvironment','return ""'),('P08',resume,'loadPackage','return packageEvidence{}, false, nil'),('P09',main,'literalChildren','return nil, nil'),('P10',resume,'predictions','return nil'),('P11',resume,'planDigest','return ""'),('P12',resume,'cleanupPackageScratch','return nil')]
# Locate complete functions with Go AST so one-line entries and nested literals are safe.
helper=Path('/tmp/u008-locations.go');helper.write_text('''package main
import("go/parser";"go/token";"go/ast";"fmt";"os")
func main(){s:=token.NewFileSet();f,e:=parser.ParseFile(s,os.Args[1],nil,0);if e!=nil{panic(e)};for _,d:=range f.Decls{if fn,ok:=d.(*ast.FuncDecl);ok&&fn.Body!=nil{fmt.Printf("%s %d %d\\n",fn.Name.Name,s.Position(fn.Body.Lbrace).Offset,s.Position(fn.Body.Rbrace).Offset)}}}
''');subprocess.run(['go','build','-o','/tmp/u008-locations',str(helper)],check=True)
for id,f,name,empty in probes:
 Path('/tmp/u008-location-source.go').write_text(orig[f])
 loc={a.split()[0]:(int(a.split()[1]),int(a.split()[2])) for a in subprocess.check_output(['/tmp/u008-locations','/tmp/u008-location-source.go'],text=True).splitlines()};start,end=loc[name]
 changed=orig[f][:start+1]+'\n'+empty+'\n'+orig[f][end:]
 if id == 'P07': changed = changed.replace('\t\"regexp\"\n', '')
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(orig[f].splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 # Locate switched function body using declaration prefix, then brace after signature.
 prefix=orig[f][orig[f].rfind('func ',0,start):start+1]; assert prefix in switched[f],name
 switched[f]=switched[f].replace(prefix,prefix+'\n if os.Getenv("ADAMIC_MUTANT") == "'+id+'" { '+empty+' }\n',1)
 mutations.append({'id':id,'file':f,'line':orig[f][:start].count('\n')+1,'description':'empty '+name+' entry','kind':'probe','entry':name})
for f,s in switched.items():Path(f).write_text(s)
Path('cmd/adamic-gate/audit_switch.go').write_text('''package main
import "os"
func auditString(id,a,b string)string{if os.Getenv("ADAMIC_MUTANT")==id{return b};return a}
func auditBool(id string,a,b bool)bool{if os.Getenv("ADAMIC_MUTANT")==id{return b};return a}
func auditMode(a uint32)uint32{if os.Getenv("ADAMIC_MUTANT")=="M14"{return 0};return a}
func auditEnv(a []string)[]string{if os.Getenv("ADAMIC_MUTANT")=="M13"{return nil};return a}
''')
(p/'mutation-plan.json').write_text(json.dumps({'base':base,'mutations':mutations},indent=2))
(p/'switch.diff').write_text(subprocess.check_output(['git','diff','--',main,resume,test],text=True))
(p/'switch-helper.go.fixture').write_text(Path('cmd/adamic-gate/audit_switch.go').read_text())
