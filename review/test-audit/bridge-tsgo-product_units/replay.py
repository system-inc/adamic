from pathlib import Path
import subprocess,json,time,difflib,os
root=Path.cwd();p=root/'review/test-audit/bridge-tsgo-product_units';src=root/'internal/buildcache/buildcache.go';base=subprocess.check_output(['git','show','HEAD:internal/buildcache/buildcache.go'],text=True)
flags='\tfor _, flag := range inputs.Flags {\n\t\tfield("flag", flag)\n\t}\n'
tools='\tfor _, tool := range inputs.Toolchain {\n\t\tfield("tool", tool)\n\t}\n'
files='\tfor _, name := range inputs.Files {\n\t\tif err := hashPath(hash, root, name, field); err != nil {\n\t\t\treturn "", err\n\t\t}\n\t}\n'
changes=[('M1',84,'if _, err = os.Stat(product); err == nil {','if _, err = os.Stat(product); err != nil {','flip condition'),('M2',132,flags,'','drop whole flag loop'),('M3',135,tools,'','drop whole toolchain loop'),('M4',127,files,'','drop whole source-file loop'),('M5',142,'if filepath.IsAbs(name) ||','if !filepath.IsAbs(name) ||','flip condition'),('M6',88,'os.OpenFile(product+".lock", os.O_CREATE|os.O_RDWR,','os.OpenFile(product+".lock", os.O_RDWR,','change option'),('P1',40,'func Product(t testing.TB, inputs Inputs, build func(directory string) error) string {\n','func Product(t testing.TB, inputs Inputs, build func(directory string) error) string {\n\treturn ""\n','empty-answer probe')]
(p/'diffs').mkdir(exist_ok=True);plan=[];compiles=[]
for mid,line,old,new,kind in changes:
 assert old in base
 altered=base.replace(old,new,1)
 if mid=='P1':
  start=base.index(old); end=base.index('\n// Get is Product',start)
  altered=base[:start]+new+'}\n'+base[end:]
  kind='empty-answer probe; remove unreachable original body'
 diff=''.join(difflib.unified_diff(base.splitlines(True),altered.splitlines(True),fromfile='a/internal/buildcache/buildcache.go',tofile='b/internal/buildcache/buildcache.go'))
 path=p/'diffs'/(mid+'.diff');path.write_text(diff)
 plan.append({'id':mid,'file':'internal/buildcache/buildcache.go','line':line,'old':old,'new':new,'kind':kind})
 start=time.monotonic()
 try:
  with (p/(mid+'-vet.log')).open('w') as out:
   for cmd in [['git','apply','--check',str(path)],['git','apply',str(path)],['timeout','90','go','vet','./internal/buildcache/']]:
    r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
    if r.returncode:raise RuntimeError((mid,cmd,r.returncode))
 finally:subprocess.run(['git','restore','--source=HEAD','--','internal/buildcache/buildcache.go'],check=True)
 compiles.append({'id':mid,'seconds':time.monotonic()-start,'exit':r.returncode})
(p/'plan.json').write_text(json.dumps(plan,indent=2));(p/'vet-status.json').write_text(json.dumps(compiles,indent=2))
s=base
s=s.replace(changes[0][2],'if _, err = os.Stat(product); (err == nil) != (os.Getenv("ADAMIC_MUTANT") == "M1") {',1)
for mid,block in [('M2',flags),('M3',tools),('M4',files)]:s=s.replace(block,'\tif os.Getenv("ADAMIC_MUTANT") != "'+mid+'" {\n'+block+'\t}\n',1)
s=s.replace(changes[4][2],'if (filepath.IsAbs(name) != (os.Getenv("ADAMIC_MUTANT") == "M5")) ||',1)
s=s.replace(changes[5][2],'os.OpenFile(product+".lock", os.O_RDWR | map[bool]int{true: 0, false: os.O_CREATE}[os.Getenv("ADAMIC_MUTANT") == "M6"],',1)
s=s.replace(changes[6][2],changes[6][2]+'\tif os.Getenv("ADAMIC_MUTANT") == "P1" { return "" }\n',1)
patch=''.join(difflib.unified_diff(base.splitlines(True),s.splitlines(True),fromfile='a/internal/buildcache/buildcache.go',tofile='b/internal/buildcache/buildcache.go'));(p/'selector.diff').write_text(patch)
subprocess.run(['git','apply',str(p/'selector.diff')],check=True)
subprocess.run(['gofmt','-w','internal/buildcache/buildcache.go'],check=True)
# Retain exactly the selector that was compiled.
sw=src.read_text();(p/'selector.diff').write_text(''.join(difflib.unified_diff(base.splitlines(True),sw.splitlines(True),fromfile='a/internal/buildcache/buildcache.go',tofile='b/internal/buildcache/buildcache.go')))
statuses=[]
try:
 for mid in ['control','M1','M2','M3','M4','M5','M6','P1']:
  env=dict(os.environ);env.pop('ADAMIC_MUTANT',None)
  if mid!='control':env['ADAMIC_MUTANT']=mid
  env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u003/cache/'+mid;env['ADAMIC_BUILD_LOG']=str(p/(mid+'-build.log'))
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./bridge/tsgo/','-run','^TestProduct_']
  start=time.monotonic()
  with (p/(mid+'.log')).open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  status={'id':mid,'command':cmd,'selector':mid,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'exit':r.returncode,'seconds':time.monotonic()-start};statuses.append(status);(p/'matrix-status.json').write_text(json.dumps(statuses,indent=2));print(status,flush=True)
  if mid in ['control','M2','M3','M4']:
   with (p/('witness-'+mid+'.json')).open('w') as out, (p/('witness-'+mid+'.log')).open('w') as err:
    subprocess.run(['go','run',str(p/'key-witness.go')],env=env,stdout=out,stderr=err,check=True)
finally:subprocess.run(['git','restore','--source=HEAD','--','internal/buildcache/buildcache.go'],check=True)
