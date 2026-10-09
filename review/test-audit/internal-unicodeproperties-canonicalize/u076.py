import pathlib,json,subprocess,os,time,difflib,re
p=pathlib.Path('review/test-audit/internal-unicodeproperties-canonicalize');(p/'diffs').mkdir(exist_ok=True);(p/'probes').mkdir(exist_ok=True)
c='internal/unicodeproperties/canonicalize.go';tab='internal/unicodeproperties/canonicalize_tables.go';test='internal/unicodeproperties/canonicalize_test.go';original={f:pathlib.Path(f).read_text() for f in [c,tab,test]};plan=[]
def add(id,file,old,new,kind):
 s=original[file];assert s.count(old)==1,(id,s.count(old));plan.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind))
add('M01',c,'if codePoint < 0 || codePoint > 0x10FFFF {\n\t\treturn codePoint','if codePoint >= 0 || codePoint > 0x10FFFF {\n\t\treturn codePoint','flip condition')
add('M02',c,'return unicodeFold[i][0] >= cp','return unicodeFold[i][0] > cp','flip comparison')
add('M03',c,'return rune(unicodeFold[index][1])','return rune(unicodeFold[index][0])','change constant index')
add('M04',c,'if codePoint < 0 || codePoint > 0x10FFFF {\n\t\treturn nil','if codePoint < -1 || codePoint > 0x10FFFF {\n\t\treturn nil','off-by-one bound')
add('M05',c,'\t\tfor i, member := range members {\n\t\t\tout[i] = rune(member)\n\t\t}','', 'drop whole copy loop')
add('M06',c,'return []rune{codePoint}','return []rune{codePoint + 1}','off-by-one value')
add('M07',c,'return legacyFold[i][0] >= codeUnit','return legacyFold[i][0] > codeUnit','flip comparison')
add('M08',c,'\t\tcopy(out, members)','', 'drop statement')
add('M09',c,'return []uint16{codeUnit}','return []uint16{codeUnit + 1}','off-by-one value')
block=original[tab].split('var unicodeFold = [][2]uint32{',1)[1].split('\n}',1)[0];last=block.splitlines()[-1]+'\n}\n\n// unicodeClassCanon[i]';numbers=re.findall(r'0x[0-9A-F]+',last);changed=last.replace(numbers[1],f'0x{int(numbers[1],16)+1:04X}');add('M10',tab,last,changed,'change last Unicode fold target constant')
# Limit replacements to unique full declarations so no duplicate table row is altered.
old='var unicodeClass = [][]uint32{\n\t{0x0041, 0x0061},';add('M11',tab,old,old.replace('0x0041','0x0042'),'change first Unicode class member constant')
old='var legacyFold = [][2]uint16{\n\t{0x0061, 0x0041},';add('M12',tab,old,old.replace('0x0041','0x0042'),'change first legacy fold target constant')
(p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
(p/'code-under-test.txt').write_text('Production functions reached: CanonicalizeUnicode, UnicodeEquivalents, CanonicalizeLegacy, LegacyEquivalents, in canonicalize.go. sort.Search is standard library. Generated data: unicodeFold, unicodeClassCanon, unicodeClass, legacyFold, legacyClassCanon, legacyClass. Unicode Node family reads unicodeClassCanon/unicodeClass directly, never calls the four production functions. Legacy Node value pass reads legacyFold directly, and its scan calls LegacyEquivalents -> CanonicalizeLegacy. Setup row checks unicodeNodeBatches in canonicalize_test.go, not production. No oracle script or assertion is mutated.\n')
def run(id,regex,env=None):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run',regex];start=time.monotonic()
 with (p/'logs'/(id+'.log')).open('w') as out:r=subprocess.run(cmd,env=dict(os.environ,**(env or {})),stdout=out,stderr=subprocess.STDOUT)
 es=[]
 for l in (p/'logs'/(id+'.log')).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return dict(id=id,command=' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-start,events=es,kills=sorted(set(e['Test'].split('/')[0] for e in es if e['Action']=='fail' and e.get('Test'))),cooked=any('test timed out' in e.get('Output','') for e in es))
# Record clean individual timings first. Full Unicode family is capped at 90 seconds per attempt.
names=['TestCanonicalizeExamples','TestEquivalentsAreClosed','TestCanonicalizeLegacyNode','TestUnicodeNodeBatchOrderAndLimit'];timings=[]
for name in names:
 for i in range(3):
  q=run('timing-'+name+'-'+str(i),'^'+name+'$');assert q['exit']==0;timings.append(q);(p/'timings.json').write_text(json.dumps(timings,indent=2))
for i in range(3):
 q=run('timing-family-'+str(i),'^TestCanonicalizeUnicodeNodeRange[0-9]+$');timings.append(q);(p/'timings.json').write_text(json.dumps(timings,indent=2));print('family timing',i,q['exit'],q['cooked'],flush=True)
# Whole package baseline cooked. Bound scans to one complete 16-line first shard;
# all quick unit rows and the full legacy differential scan still run.
regex='^(TestCanonicalizeExamples|TestEquivalentsAreClosed|TestCanonicalizeLegacyNode|TestUnicodeNodeBatchOrderAndLimit|TestCanonicalizeUnicodeNodeRange0)$/^0000-0016-'
q=run('bounded-baseline',regex);assert q['exit']==0;(p/'bounded-baseline.json').write_text(json.dumps(q,indent=2))
# Unicode 17 source value checked independently, plus Node fold-equivalence witness.
with (p/'logs'/'authority.log').open('w') as out:subprocess.run(['node','-e',"console.log(JSON.stringify({legacy_a:'a'.toUpperCase(),unicode_kelvin:/^k$/iu.test('K'),unicode_version:process.versions.unicode}))"],stdout=out,stderr=subprocess.STDOUT,check=True)
def diff(q,after,folder):
 (p/folder/(q['id']+'.diff')).write_text(''.join(difflib.unified_diff(original[q['file']].splitlines(True),after.splitlines(True),fromfile='a/'+q['file'],tofile='b/'+q['file'])))
def validate(q,after):
 file=pathlib.Path(q['file']);start=time.monotonic()
 try:
  file.write_text(after)
  with (p/'logs'/('validate-'+q['id']+'.log')).open('w') as out:subprocess.run(['go','vet','./internal/unicodeproperties/'],stdout=out,stderr=subprocess.STDOUT,check=True)
  q['validation_seconds']=time.monotonic()-start;q['validation_command']='go vet ./internal/unicodeproperties/'
 finally:file.write_text(original[q['file']])
try:
 for q in plan:
  after=original[q['file']].replace(q['old'],q['new']);diff(q,after,'diffs');validate(q,after)
 (p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
 # All menu edits behind one selector in scratch source.
 switched=dict(original)
 for q in plan:
  old,new=q['old'],q['new']
  if q['file']==tab:
   # Table mutations controlled once at package initialization.
   continue
  if q['id']=='M05':replacement='\t\tif auditMutant != "M05" {\n'+old+'\n\t\t}'
  elif q['id']=='M08':replacement='\t\tif auditMutant != "M08" { copy(out, members) }'
  elif q['id']=='M01':replacement='if (auditMutant == "M01" && codePoint >= 0) || codePoint < 0 || codePoint > 0x10FFFF {\n\t\treturn codePoint'
  elif q['id']=='M04':replacement='if (codePoint < 0 && !(auditMutant == "M04" && codePoint == -1)) || codePoint > 0x10FFFF {\n\t\treturn nil'
  elif q['id'] in ['M02','M07']:
   replacement='if auditMutant == "'+q['id']+'" { '+new+' }; '+old
  else:replacement='if auditMutant == "'+q['id']+'" { '+new+' }; '+old
  switched[c]=switched[c].replace(old,replacement)
 probeplan=[]
 for i,(signature,value) in enumerate([('func CanonicalizeUnicode(codePoint rune) rune {','0'),('func UnicodeEquivalents(codePoint rune) []rune {','nil'),('func CanonicalizeLegacy(codeUnit uint16) uint16 {','0'),('func LegacyEquivalents(codeUnit uint16) []uint16 {','nil')],1):
  q=dict(id='P0'+str(i),file=c,old=signature,new=signature+'\n\tif true { return '+value+' }',line=original[c][:original[c].index(signature)].count('\n')+1);after=original[c].replace(q['old'],q['new']);diff(q,after,'probes');validate(q,after);probeplan.append(q);switched[c]=switched[c].replace(signature,signature+'\n\tif auditMutant == "'+q['id']+'" {return '+value+'}')
 (p/'probe-plan.json').write_text(json.dumps(probeplan,indent=2))
 pathlib.Path(c).write_text(switched[c]);pathlib.Path('internal/unicodeproperties/audit_switch.go').write_text('package unicodeproperties\nimport "os"\nvar auditMutant = os.Getenv("ADAMIC_MUTANT")\nfunc init(){switch auditMutant {case "M10":unicodeFold[len(unicodeFold)-1][1]++;case "M11":unicodeClass[0][0]=0x42;case "M12":legacyFold[0][1]=0x42}}\n')
 start=time.monotonic()
 with (p/'logs'/'compile-switch.log').open('w') as out:subprocess.run(['go','test','-c','-o','/tmp/u076.test','./internal/unicodeproperties/'],stdout=out,stderr=subprocess.STDOUT,check=True)
 (p/'build.json').write_text(json.dumps(dict(seconds=time.monotonic()-start)))
 q=run('switch-baseline',regex);assert q['exit']==0
 matrix=[]
 for m in plan:
  q=run(m['id'],regex,dict(ADAMIC_MUTANT=m['id']));assert not q['cooked'];matrix.append(q);(p/'matrix.json').write_text(json.dumps(matrix,indent=2));print(m['id'],q['kills'],flush=True)
 probes=[]
 for m in probeplan:
  q=run(m['id'],regex,dict(ADAMIC_MUTANT=m['id']));assert not q['cooked'];probes.append(q);(p/'probe-results.json').write_text(json.dumps(probes,indent=2));print(m['id'],q['kills'],flush=True)
finally:
 for f,s in original.items():pathlib.Path(f).write_text(s)
 pathlib.Path('internal/unicodeproperties/audit_switch.go').unlink(missing_ok=True)
# Setup-only edit allowed by brief: drop cap assignment, retaining all comparisons.
q=dict(id='S01',file=test,old='\t\tworkers = limit',new='',line=original[test][:original[test].index('\t\tworkers = limit')].count('\n')+1)
after=original[test].replace(q['old'],q['new']);diff(q,after,'probes');validate(q,after)
try:
 pathlib.Path(test).write_text(after);result=run('S01','^TestUnicodeNodeBatchOrderAndLimit$');(p/'setup-result.json').write_text(json.dumps(result,indent=2));print('S01',result['exit'],result['kills'],flush=True)
finally:pathlib.Path(test).write_text(original[test])
