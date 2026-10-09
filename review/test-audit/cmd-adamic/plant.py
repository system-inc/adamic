from pathlib import Path
import subprocess,difflib,json,sys,shutil,time
root=Path(__file__).resolve().parents[3]
out=Path(__file__).resolve().parent
base_commit='1f34d0d300301faebc94d397adee1a090923d2c7'
files=['cmd/adamic/main.go','cmd/adamic/request.go','cmd/adamic/checks.go','cmd/adamic/tsgo.go']
base={f:subprocess.check_output(['git','show',base_commit+':'+f],cwd=root,text=True) for f in files}
# Fixed before observing any mutant catches. Each tuple is id, file, old, standalone, switch, menu.
mutations=[
('M01',files[0],'explain = true','explain = false','explain = os.Getenv("ADAMIC_MUTANT") != "M01"','change option'),
('M02',files[0],'fmt.Fprintln(os.Stderr, usage)\n\treturn 2','fmt.Fprintln(os.Stderr, usage)\n\treturn 1','fmt.Fprintln(os.Stderr, usage)\n\tif os.Getenv("ADAMIC_MUTANT") == "M02" { return 1 }; return 2','change constant'),
('M03',files[0],'return lowered, 0','return lowered, 1','if os.Getenv("ADAMIC_MUTANT") == "M03" { return lowered, 1 }; return lowered, 0','change constant'),
('M04',files[2],'if status == "unobservable" {','if status != "unobservable" {','if (status == "unobservable") != (os.Getenv("ADAMIC_MUTANT") == "M04") {','flip condition'),
('M05',files[2],'if site.Proven {','if !site.Proven {','if site.Proven != (os.Getenv("ADAMIC_MUTANT") == "M05") {','flip condition'),
('M06',files[3],'options.Count = true','options.Count = false','options.Count = os.Getenv("ADAMIC_MUTANT") != "M06"','change option'),
('M07',files[3],'options.Sanitize = true','options.Sanitize = false','options.Sanitize = os.Getenv("ADAMIC_MUTANT") != "M07"','change option'),
('M08',files[3],'arguments[index] != "wasm32-wasi"','arguments[index] == "wasm32-wasi"','(arguments[index] != "wasm32-wasi") != (os.Getenv("ADAMIC_MUTANT") == "M08")','flip condition'),
('M09',files[3],'if options.Target != "" && archive != "" {','if false && options.Target != "" && archive != "" {','if os.Getenv("ADAMIC_MUTANT") != "M09" && options.Target != "" && archive != "" {','flip condition'),
('M10',files[1],'if handler >= 0 {','if false && handler >= 0 {','if os.Getenv("ADAMIC_MUTANT") != "M10" && handler >= 0 {','flip condition'),
('M11',files[1],'if typeChecker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsStringLike == 0 {','if false && typeChecker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsStringLike == 0 {','if os.Getenv("ADAMIC_MUTANT") != "M11" && typeChecker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsStringLike == 0 {','flip condition'),
('M12',files[3],'options.Request = handler >= 0','options.Request = options.Count','if os.Getenv("ADAMIC_MUTANT") == "M12" { options.Request = options.Count } else { options.Request = handler >= 0 }','change option'),
]
for mid,f,signature,value in [
 ('E_RUN',files[0],'func run(arguments []string) int {','0'),
 ('E_CHECK',files[0],'func check(paths []string) (*load.Program, int) {','nil, 0'),
 ('E_COMPILE',files[0],'func compile(path string) (*ir.Program, int) {','nil, 0'),
 ('E_WASI',files[1],'func compileWASI(path string) (*ir.Program, int, int) {','nil, 0, 0'),
 ('E_SELECT',files[1],'func requestFunction(program *load.Program) (string, error) {','"", nil'),
]:
 mutations.append((mid,f,signature,signature+'\n\tif true { return '+value+' }',signature+'\n\tif os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { return '+value+' }','return early; empty-answer probe'))
manifest=[]
switched=dict(base)
for mid,f,old,new,switch,menu in mutations:
 assert base[f].count(old)==1,(mid,base[f].count(old))
 changed=base[f].replace(old,new,1)
 line=base[f][:base[f].index(old)].count('\n')+1
 if mid=='M02': line+=1
 (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(base[f].splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 manifest.append(dict(id=mid,file=f,line=line,before=old,after=new,menu=menu,probe=mid.startswith('E_'),supplemental=False))
 switched[f]=switched[f].replace(old,switch,1)
(out/'mutants.json').write_text(json.dumps(manifest,indent=2)+'\n')
(out/'switched-sources.json').write_text(json.dumps(switched,indent=2)+'\n')
if '--install' in sys.argv:
 for f in files:
  assert (root/f).read_text()==base[f],f+' is not clean'
 for f in files: (root/f).write_text(switched[f])
if '--vet' in sys.argv:
 scratch=Path('/tmp/u007-vet-overlay')
 scratch.mkdir(exist_ok=True)
 records=[]
 for mid,f,old,new,switch,menu in mutations:
  folder=scratch/mid; folder.mkdir(exist_ok=True)
  replacements={}
  for path in files:
   copy=folder/Path(path).name
   copy.write_text(base[path].replace(old,new,1) if path==f else base[path])
   replacements[str(root/path)]=str(copy)
  overlay=folder/'overlay.json'
  overlay.write_text(json.dumps({'Replace':replacements}))
  log=out/(mid+'-vet.log'); started=time.monotonic()
  cmd=['timeout','90','go','vet','-overlay='+str(overlay),'./cmd/adamic/']
  with log.open('w') as stream:
   result=subprocess.run(cmd,cwd=root,stdout=stream,stderr=subprocess.STDOUT)
  records.append(dict(mutant=mid,command=' '.join(cmd),cwd=str(root),seconds=time.monotonic()-started,exit=result.returncode,output=log.read_text()))
  (out/'standalone-vet.json').write_text(json.dumps(records,indent=2)+'\n')
  print('vet',mid,result.returncode,flush=True)
  if result.returncode: raise SystemExit('standalone diff failed vet')
