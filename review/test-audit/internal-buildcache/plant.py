import json,subprocess,difflib,sys
from pathlib import Path
P=Path('review/test-audit/internal-buildcache'); S=Path('/tmp/u015'); F='internal/buildcache/buildcache.go';BASE=subprocess.check_output(['git','show','origin/main:'+F],text=True)
def loop(prefix,entry):return '\tfor _, '+prefix+' {\n\t\t'+entry+'\n\t}'
plan=[
('M01','\tfield("name", inputs.Name)','', 'drop statement',0),
('M02',loop('flag := range inputs.Flags','field("flag", flag)'),'','drop whole loop',0),
('M03',loop('tool := range inputs.Toolchain','field("tool", tool)'),'','drop whole loop',0),
('M04','field("directory", relative)','','drop statement',0),
('M05','sha256.Sum256(content)','sha256.Sum256(content[:0])','change bound',0),
('M06','info.Mode()&0o111 != 0','false','change constant',0),
('M07','field("buildcache", "v1")','field("buildcache", "v2")','change constant',0),
('M08','if filepath.IsAbs(name) || strings.HasPrefix(filepath.Clean(name), "..") {','if false && (filepath.IsAbs(name) || strings.HasPrefix(filepath.Clean(name), "..")) {','flip condition',0),
('M09','if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {','if false && os.Getenv("ADAMIC_BUILD_CACHE") == "off" {','flip condition',0),
('M10','if _, err = os.Stat(product); err == nil {','if _, err = os.Stat(product); false && err == nil {','flip condition',0),
('M11','if _, err = os.Stat(product); err == nil {','if _, err = os.Stat(product); false && err == nil {','flip condition',1),
('M12','syscall.LOCK_EX','syscall.LOCK_SH','change option',0),
('M13','\t\tos.RemoveAll(scratch)','','drop statement',0),
('M14','os.Rename(scratch, product)','os.Rename(product, scratch)','swap arguments',0),
('M15','file.WriteString(line + "\\n")','','drop statement',0),
('M16',loop('tool := range inputs.Toolchain','lines = append(lines, "tool "+tool)'),'','drop whole loop',0),
('M17','tools.Store(command, value)','','drop statement',0),
('M18','value := command + ": " + strings.TrimSpace(string(output))','value := command + " " + strings.TrimSpace(string(output))','change constant',0),
('P_PRODUCT','func Product(t testing.TB, inputs Inputs, build func(directory string) error) string {','func Product(t testing.TB, inputs Inputs, build func(directory string) error) string {\n if true { return "" }','empty answer',0),
('P_GET','func Get(inputs Inputs, build func(directory string) error) (string, error) {','func Get(inputs Inputs, build func(directory string) error) (string, error) {\n if true { return "", nil }','empty answer',0),
('P_KEY','func Key(root string, inputs Inputs) (string, error) {','func Key(root string, inputs Inputs) (string, error) {\n if true { return "", nil }','empty answer',0),
('P_TOOL','func Tool(name string, arguments ...string) string {','func Tool(name string, arguments ...string) string {\n if true { return "" }','empty answer',0)]
def position(source,before,index):
 start=-1
 for n in range(index+1):start=source.index(before,start+1)
 return start
manifest=[]; switches=[]
for id,before,after,menu,occurrence in plan:
 at=position(BASE,before,occurrence);new=BASE[:at]+after+BASE[at+len(before):];diff=''.join(difflib.unified_diff(BASE.splitlines(True),new.splitlines(True),fromfile='a/'+F,tofile='b/'+F));(P/(id+'.diff')).write_text(diff)
 q=S/'standalone'/id;q.mkdir(parents=True,exist_ok=True);f=q/'buildcache.go';f.write_text(new);(q/'overlay.json').write_text(json.dumps({'Replace':{str(Path(F).resolve()):str(f)}}))
 manifest.append({'id':id,'file':F,'line':BASE.count('\n',0,at)+1,'before':before,'after':after,'menu':menu,'occurrence':occurrence,'probe':id.startswith('P_'),'supplemental':False})
 select='os.Getenv("ADAMIC_MUTANT") == "'+id+'"'
 if id.startswith('P_'):replacement=before+'\n if '+select+' { '+('return "", nil' if id in ['P_GET','P_KEY'] else 'return ""')+' }'
 elif id in ['M10','M11']:replacement='if _, err = os.Stat(product); err == nil && !('+select+') {'
 elif id=='M08':replacement='if !( '+select+' ) && (filepath.IsAbs(name) || strings.HasPrefix(filepath.Clean(name), "..")) {'
 elif id=='M09':replacement='if !( '+select+' ) && os.Getenv("ADAMIC_BUILD_CACHE") == "off" {'
 elif id=='M12':replacement='func() int { if '+select+' { return syscall.LOCK_SH }; return syscall.LOCK_EX }()'
 elif id in ['M05','M06']:replacement='func() '+('[]byte' if id=='M05' else 'bool')+' { if '+select+' { return '+('content[:0]' if id=='M05' else 'false')+' }; return '+('content' if id=='M05' else before)+' }()';replacement='sha256.Sum256('+replacement+')' if id=='M05' else replacement
 elif id=='M14':replacement='func() error { if '+select+' { return '+after+' }; return '+before+' }()'
 elif id=='M18':replacement=before+'\n if '+select+' { value = command + " " + strings.TrimSpace(string(output)) }'
 elif not after:replacement='if !('+select+') {\n'+before+'\n}'
 else:replacement='if '+select+' { '+after+' } else { '+before+' }'
 switches.append((at,len(before),replacement))
(P/'plan.json').write_text(json.dumps(manifest,indent=2))
if sys.argv[-1]=='install':
 source=BASE
 for at,length,replacement in sorted(switches,reverse=True):source=source[:at]+replacement+source[at+length:]
 Path(F).write_text(source);subprocess.run(['gofmt','-w',F],check=True);(P/'switch.diff').write_text(subprocess.check_output(['git','diff','--',F],text=True))
print('planned',len(manifest),'including four probes')
