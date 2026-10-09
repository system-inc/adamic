import pathlib,subprocess,time,json,os
root=pathlib.Path.cwd(); out=root/'review/test-audit/internal-load-pin'
files=['internal/load/regexp_library.go','internal/load/load.go']
original={f:subprocess.check_output(['git','show','HEAD:'+f],text=True) for f in files}
lines=original[files[0]].splitlines(True)
menu={f'M{i}':(files[0],lines[18+i],'') for i in range(1,4)}
menu['M4']=(files[1],'Strict:                     core.TSTrue,','Strict:                     core.TSFalse,')
menu['M5']=(files[1],'"error TS%d: %s", diagnostic.Code(), message','"error TS%d: %s", 0, message')
menu['M6']=(files[1],'len(utf16.Encode([]rune(prefix))) + 1','len(utf16.Encode([]rune(prefix))) + 0')
menu['P1']=(files[1],'return load(paths, nil)','return nil, nil')
def command(args,name):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log: p=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT)
 return {'exit':p.returncode,'wall':round(time.monotonic()-start,3)}
measure={}
for test in ['TestTypeScriptIsThePinnedCommit','TestRegExpCaptureTypes']:
 for i in range(1,4): measure[f'{test}-{i}']=command(['timeout','120','go','test','-count=1','-timeout','90s','./internal/load/','-run','^'+test+'$'],f'{test}-{i}')
for mid,(f,old,new) in menu.items():
 assert old in original[f],(mid,old)
 pathlib.Path(f).write_text(original[f].replace(old,new,1))
 diff=subprocess.check_output(['git','diff','--',f],text=True)
 (out/(mid+'.diff')).write_text(diff)
 measure[mid+'-vet']=command(['timeout','120','go','vet','./internal/load/'],mid+'-vet')
 pathlib.Path(f).write_text(original[f])
# All changes behind one selector, scaffolding is excluded from standalone diffs.
s=original[files[0]].replace('"strings"','"strings"\n "os"',1)
for mid in ['M1','M2','M3']:
 _,old,_=menu[mid]
 s=s.replace(old,'\t\tif os.Getenv("ADAMIC_MUTANT") != "'+mid+'" {\n'+old+'\t\t}\n',1)
pathlib.Path(files[0]).write_text(s)
s=original[files[1]]
s=s.replace('return load(paths, nil)','if os.Getenv("ADAMIC_MUTANT") == "P1" { return nil, nil }; return load(paths, nil)',1)
s=s.replace('Strict:                     core.TSTrue,','Strict: func() core.Tristate { if os.Getenv("ADAMIC_MUTANT") == "M4" { return core.TSFalse }; return core.TSTrue }(),',1)
s=s.replace('"error TS%d: %s", diagnostic.Code(), message','"error TS%d: %s", func() int32 { if os.Getenv("ADAMIC_MUTANT") == "M5" { return 0 }; return diagnostic.Code() }(), message',1)
s=s.replace('len(utf16.Encode([]rune(prefix))) + 1','len(utf16.Encode([]rune(prefix))) + func() int { if os.Getenv("ADAMIC_MUTANT") == "M6" { return 0 }; return 1 }()',1)
pathlib.Path(files[1]).write_text(s)
measure['switch-build']=command(['go','test','-c','-o',str(out/'load.test'),'./internal/load/'],'switch-build')
if measure['switch-build']['exit']==0:
 for mid in menu:
  os.environ['ADAMIC_MUTANT']=mid
  measure[mid]=command(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/load/','-run','.'],mid)
for f in files: pathlib.Path(f).write_text(original[f])
(out/'timings.json').write_text(json.dumps(measure,indent=2)+'\n')
print(json.dumps(measure,indent=2))
