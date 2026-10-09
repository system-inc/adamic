from pathlib import Path
import json,re,hashlib
repo=Path('/workspace/cemit-cut'); base=repo/'stage1/cohere/typeaware'; out=Path('/tmp/cemit-inputs');out.mkdir(exist_ok=True)
rows=[];tests={}
def add(test,name,path,tsgo=True):
 rows.append(dict(Name=name,Path=str(path),TSGo=tsgo));tests.setdefault(test,[]).append(name)
def write(name,text):
 p=out/(name+'.a');p.write_text(text);return p
def fix(text):
 text=text.replace('../../../typescript',str(repo/'stage1/typescript')).replace('../../typescript',str(repo/'stage1/typescript'))
 for file in ['facts.ts','diagnostic.ts','unary_minus.ts','rules.ts','flags.ts','frames.ts','types.ts','type_fact.ts','parameters.ts']:
  text=text.replace('./'+file,str(base/file))
 return text
six='TestSixRuleAgreementAndMutants';ta='TestTypeAwareAgreementAndMutants';v='TestVolumeAgreementCompiler';g='TestVolumeConfigGuardAndMutant';css='TestCSSPrinterAgreesWithGo';co='TestCompositionMatchesGoUnion'
entry=base/'testdata/sharded_suite.ts';add(six,'six-suite',entry)
driver=entry.read_text().replace('../../../typescript',str(repo/'stage1/typescript'))
for file in ['facts.ts','diagnostic.ts','unary_minus.ts','rules.ts','flags.ts','frames.ts','types.ts','type_fact.ts','parameters.ts']:driver=driver.replace('../'+file,'./'+file)
def suite(name,frm,to,imports=[]):
 text=(base/'rules.ts').read_text();assert text.count(frm)==1;text=text.replace(frm,to,1)
 for f,t in imports:assert text.count(f)==1;text=text.replace(f,t,1)
 p=write('six-'+name+'-rules',fix(text));main=fix(driver.replace('./rules.ts',str(p),1));add(six,'six-'+name,write('six-'+name+'-main',main))
for name,f,t in [('wrong-node',"this.ask(left, 'base-type')","this.ask(index, 'base-type')"),('last-declaration','if(use && pos < earliestPos)','if(use)'),('nullable-default','if(!plain || nullable)','if(!plain && !nullable)')]:suite(name,f,t)
facts=write('six-union-types',fix((base/'types.ts').read_text().replace('return (type.flags & mask) !== 0 ? type.parts : [type.id];','return mask === 0 ? type.parts : [type.id];',1)))
decoder=write('six-union-decoder',fix((base/'facts.ts').read_text().replace("from './types.ts'","from '"+str(facts)+"'",1)))
params=write('six-union-parameters',(base/'parameters.ts').read_text().replace("from './types.ts'","from '"+str(facts)+"'",1))
suite('union-members',"from './facts.ts'","from '"+str(decoder)+"'",[("from './types.ts'","from '"+str(facts)+"'"),("from './parameters.ts'","from '"+str(params)+"'")])
# The generated released driver is a Go quoted literal in the test.
src=(base/'suite_test.go').read_text();literal=re.search(r'releasedEntry := h.write\("released-inspect.ts", ("(?:[^"\\]|\\.)*")\)',src).group(1)
add(six,'six-released',write('six-released',json.loads(literal)));add(six,'six-facts-cost',base/'testdata/fact_cost.ts')
add(ta,'typeaware-main',base/'main.ts')
text=(base/'unary_minus.ts').read_text().replace("this.parser.node(node.children[0] ?? panic('minus without operand'))","this.parser.node(index)",1).replace('../../typescript',str(repo/'stage1/typescript')).replace('../lint',str(repo/'stage1/cohere/lint'))
p=write('typeaware-wrong-node',text);main=(base/'main.ts').read_text().replace('../../typescript',str(repo/'stage1/typescript')).replace('./unary_minus.ts',str(p),1)
add(ta,'typeaware-wrong-main',write('typeaware-wrong-main',main));add(ta,'typeaware-released',base/'testdata/released.ts');add(ta,'typeaware-query-cost',base/'testdata/query_cost.ts');add(ta,'typeaware-bad-kind',write('typeaware-bad-kind',(base/'testdata/query_cost.ts').read_text().replace("'PrefixUnaryExpression'","'Identifier'",1)))
add(g,'guard-volume',base/'volume_suite.ts')
# TestVolumeAgreementCompiler is introduced after the cut. Its helper generates
# this adapter from volume_suite.ts; preserve that adapter on the cut sources.
text=(base/'volume_suite.ts').read_text();frm='const program = tsgoProgram(config, paths);';assert text.count(frm)==1
text=text.replace(frm,"const rootsManifest = args[3] ? readTextFile(args[3]) : manifest;\nif(rootsManifest.kind === 'Error') { panic(rootsManifest.message); }\nconst program = tsgoProgram(config, rootsManifest.text.split('\\n').filter((path) => path !== ''));",1)
text=re.sub(r"from '([^']+)'",lambda m:"from '"+str((base/m[1]).resolve())+"'" if m[1].startswith('.') else m[0],text)
add(v,'volume-with-program-roots',write('volume-with-program-roots',text))
cssbase=repo/'stage1/cohere/css';add(co,'composition',cssbase/'compose_main.ts',False)
add(css,'printer',cssbase/'print_main.ts',False)
mutants=[('semicolon','print.ts',"let ending = d.text(';');\n        if(n.truth('nodes')) ending","let ending = d.text('');\n        if(n.truth('nodes')) ending"),('indent','print.ts','d.indent(d.concat([d.hard(), this.sequence(path)]))','d.concat([d.hard(), this.sequence(path)])'),('width','print_doc.ts','flat = !n.broken && this.fits(command(n.parts[0] ?? 0, c.indent, true), stack, width - column, suffix.length > 0, false);','flat = !n.broken;')]
for name,file,f,t in mutants:
 root=out/('printer-'+name)
 for sl in ['css','selector','values','mediaquery','cssstrings','cssnumbers']:
  dest=root/sl;dest.mkdir(parents=True,exist_ok=True)
  for p in (repo/'stage1/cohere'/sl).glob('*.ts'):
   text=p.read_text()
   if sl=='css' and p.name==file:assert text.count(f)==1;text=text.replace(f,t,1)
   text=text.replace('.ts\'', '.a\'').replace('.ts"','.a"')
   (dest/(p.stem+'.a')).write_text(text)
 add(css,'printer-'+name,root/'css/print_main.a',False)
tests['TestUnicodeNodeShardPlantedFailure']=[]
for row in rows:row['RootSHA256']=hashlib.sha256(Path(row['Path']).read_bytes()).hexdigest()
Path('/tmp/cemit-manifest.json').write_text(json.dumps(rows,indent=2)+'\n')
evidence=Path('/workspace/adamic/review/compiler/c-emission-identity');(evidence/'manifest.json').write_text(json.dumps(dict(tests=tests,programs=rows),indent=2)+'\n')
print({k:len(v) for k,v in tests.items()});print('distinct programs',len(rows))
