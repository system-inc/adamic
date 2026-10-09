import pathlib,json,subprocess,re
D=pathlib.Path('review/test-audit/stage1-cohere-estree-gaps');plan=[]
def add(id,file,old,new,menu,rows,kind='production'):
 s=pathlib.Path(file).read_text();assert s.count(old)==1,(id,s.count(old));plan.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,menu=menu,matrix_rows=rows,kind=kind))
jsx=['TestJSXAgreement family'];gap=['TestPostfixValueGap','TestMethodReplacementGap','TestInterfaceDefaultGap','TestInterfaceTypeMethodGap']
add('M01','stage1/cohere/estree/jsxConvert.ts',"'JSXIdentifier'","'JSXName'",'change constant',jsx)
add('M02','stage1/typescript/parser/parser.ts',"const members: number[] = [];\n        while(this.kind() !== 'CloseBraceToken')", "const members: number[] = [];\n        while(this.kind() === 'CloseBraceToken')",'flip condition',['TestParserRecoveryGap'])
add('M03','internal/native/runtime/utf8.c','return (double)text->length;','return (double)text->length + 1;','change constant',['TestRawInputGap']+jsx)
add('M04','stage1/cohere/estree/protocol.ts','new DumpFrame(property.value.node, level + 1, -1)','new DumpFrame(property.value.node, level + 2, -1)','change constant',jsx)
add('M05','internal/lower/diagnostics.go','return "a " + name','return "an " + name','change constant',gap)
add('M06','internal/lower/diagnostics.go','Where: l.program.Where(node), What: what','Where: l.program.Where(node) + ":0", What: what','change constant',gap)
add('M07','internal/lower/diagnostics.go','return fmt.Sprintf("%s: Adamic 0.1 refuses %s; %s", r.Where, r.What, r.Fix)','return fmt.Sprintf("%s: Adamic 0.1 refuses %s", r.Where, r.What)','drop argument and change format option',gap)
add('H01','stage1/cohere/estree/misc_shards_test.go','return nil, fmt.Errorf("%d cases cannot enumerate %d shards", len(ids), count)','return nil, nil','weaken construction rejection',['TestMiscShardUnionRejectsInvalidEnumeration'], 'setup')
add('H02','stage1/cohere/estree/misc_shards_test.go','shard := miscShard(id, count)','shard := i % count','break stable construction',['TestMiscShardGrowthKeepsAssignments'],'setup')
add('H03','stage1/cohere/estree/recovery_cache_test.go','sha256.Sum256([]byte(normalized))','sha256.Sum256([]byte(normalized[:0]))','drop mutable input content from key',['TestRecoveryCacheInputKeys'],'setup')
old='args := native.Flags(native.Options{Sanitize: true})';add('H04','stage1/cohere/estree/recovery_cache_test.go',old,'args := native.Flags(native.Options{Sanitize: false})','disable construction sanitizer option',['TestRecoveryNativeRecipe'],'setup')
add('W01','stage1/cohere/estree/misc_shards_test.go','diff := firstDifference(want, got)','diff := ""','force comparison to report agreement',['TestJSXAgreementPlantedDisagreement','TestJSXMutant','TestJSXMutantPlantedSurvivor'],'witness')
(D/'plan.json').write_text(json.dumps(plan,indent=2));(D/'origin.txt').write_text(subprocess.check_output(['git','rev-parse','HEAD']).decode())
files=list(pathlib.Path('stage1/cohere/estree').glob('*.ts'))+list(pathlib.Path('stage1/typescript').rglob('*.ts'))+list(pathlib.Path('internal/lower').glob('*.go'))+[pathlib.Path('internal/native/runtime/utf8.c')]
lines=['CODE UNDER TEST: Lower diagnostic construction; ESTree TypeScript main/run, answer, convertJsx and protocol traversal; original Parser.file/typeLiteral; native adamic_utf8_length. Setup constructions and witnessed comparison are listed separately in plan.json.','ORACLES: Node fixture outputs plus self-written gap labels; external Go cohere ESTree output; upstream installed npm libraries (oracle-only row); expected timeout; sanitizer diagnostic strings; self-written shard stability and cache-key assertions.','The following is a conservative complete source-function inventory, including unreached declarations. Exact Go diagnostic-path reachability is in lower-functions.txt. TypeScript functions are an upper bound, not a measured call graph.']
for p in files:
 if p.name.endswith('_test.go'):continue
 for n,line in enumerate(p.read_text().splitlines(),1):
  if re.search(r'^func |function | => |^    (?:[a-zA-Z_][a-zA-Z_0-9]*|constructor)\(',line):lines.append(f'{p}:{n}: {line.strip()}')
(D/'inventory.txt').write_text('\n'.join(lines)+'\n')
