import pathlib,re,json,difflib
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-markdownblocks-array_growth_gap';(out/'diffs').mkdir(exist_ok=True)
mutations=[
('M1','internal/native/emit_branches.go','condition := e.value(conditional.Condition)','condition := "!(" + e.value(conditional.Condition) + ")"','flip native conditional selection'),
('M2','internal/native/runtime/array.c','is outside an array of length','is outside a sequence of length','change bounds diagnostic constant'),
('M3','stage1/cohere/markdownblocks/inputChunks.ts','Math.ceil(column / 4) * 4','Math.ceil(column / 4) * 8','change tab expansion constant'),
('M4','stage1/cohere/markdownblocks/leaves.ts',"? '**'","? '__'",'change strong delimiter constant'),
('S1','stage1/cohere/markdownblocks/preflight_shards_test.go','shard := preflightShard(input.Name)','shard := (preflightShard(input.Name) + 1) % testWholeDocumentOraclePreflightShards','construction: change bucket constant'),
('S3','stage1/cohere/markdownblocks/tokenizer_events_union_test.go','name, occurrences[name])','name, i)','construction: swap local occurrence argument for global index'),
('S4','stage1/cohere/markdownblocks/leaf_composition_independent_test.go','leafCompositionPrepared, leafCompositionBuilds = fixture, products','leafCompositionBuilds = products','construction: drop prepared fixture assignment'),
('W1','stage1/cohere/markdownblocks/tokenizer_events_union_test.go','if bytes.Equal(want, got) {','if len(want) >= 0 {','witness: loosen output comparison'),
('W2','stage1/cohere/markdownblocks/leaf_composition_independent_test.go','if !bytes.Equal(actual, expected) {','if len(actual) < 0 || len(expected) < 0 {','witness: disable planted output comparison'),
('P_C','internal/native/emit.go','return cProgram(program, -1)','return ""','empty native compiler entry'),
('P_chunks','stage1/cohere/markdownblocks/inputChunks.ts','export function inputChunks(units: readonly number[]): InputChunkInterface[] {','export function inputChunks(units: readonly number[]): InputChunkInterface[] {\n    return [];','empty chunk entry'),
('P_decode','stage1/cohere/markdownblocks/decodeString.ts','export function decodeString(value: string): string {',"export function decodeString(value: string): string {\n    return '';",'empty decoder entry'),
('P_identifier','stage1/cohere/markdownblocks/identifier.ts','export function identifier(source: string): string {',"export function identifier(source: string): string {\n    return '';",'empty identifier entry'),
('P_front','stage1/cohere/markdownblocks/frontmatter.ts','export function parseFrontMatter(text: string): FrontMatterResultInterface {',"export function parseFrontMatter(text: string): FrontMatterResultInterface {\n    return {frontMatter: undefined, content: ''};",'empty frontmatter entry'),
('P_ast','stage1/cohere/markdownblocks/astPreprocess.ts','run(root: number): number {','run(root: number): number {\n        return 0;','empty AST entry'),
('P_leaf','stage1/cohere/markdownblocks/leaves.ts','export function printLeaf(arena: DocumentArena, frame: LeafFrameInterface): number {','export function printLeaf(arena: DocumentArena, frame: LeafFrameInterface): number {\n    return 0;','empty leaf entry'),
]
# Drop the whole union checker, preserving imports used by other functions.
p='stage1/cohere/markdownblocks/tokenizer_events_union_test.go';text=(root/p).read_text();start=text.index('func tokenizerEventUnion(');end=text.index('\nfunc validateTokenizerEventUnion',start)
a=text[start:end];b='func tokenizerEventUnion(total int, shards [][]int) error {\n\treturn nil\n}\n';mutations.append(('S2',p,a,b,'construction: return early from union validation'))
entries=[]
for id,p,a,b,desc in mutations:
 old=(root/p).read_text();assert old.count(a)==1,(id,old.count(a));new=old.replace(a,b,1)
 diff=''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+p,tofile='b/'+p))
 (out/'diffs'/f'{id}.diff').write_text(diff)
 entries.append(dict(id=id,file=p,line=old[:old.index(a)].count('\n')+1,from_=a,to=b,menu=desc,kind='production' if id.startswith('M') else 'probe' if id.startswith('P') else 'construction' if id.startswith('S') else 'witness'))
(out/'menu.json').write_text(json.dumps(entries,indent=2))
# Inventory the transitive port source closure, excluding transported Go/Node oracles.
entry_names=['ast_probe.ts','chunks_probe.ts','decode_probe.ts','events_probe.ts','frontmatter_probe.ts','identifier_probe.ts','list_probe.ts']
seen=set();queue=[root/'stage1/cohere/markdownblocks/testdata'/x for x in entry_names];inventory=[]
while queue:
 p=queue.pop()
 if p in seen or not p.exists():continue
 seen.add(p);data=p.read_text()
 for spec in re.findall(r"from\s+['\"]([^'\"]+)['\"]",data):
  if spec.startswith('.'):queue.append((p.parent/spec).resolve())
 for n,l in enumerate(data.splitlines(),1):
  m=re.search(r'(?:export\s+)?function\s+(\w+)\s*\(|^\s*(?:static\s+)?(constructor|\w+)\([^;]*\)\s*(?::\s*[^={]+)?\s*\{',l)
  if m and not l.lstrip().startswith(('if(','while(','for(','switch(')):
   inventory.append(dict(file=str(p.relative_to(root)),line=n,function=next(x for x in m.groups() if x)))
(out/'port-function-inventory.json').write_text(json.dumps(inventory,indent=2))
# All compiler functions are a conservative inventory, no exact dynamic reachability claimed.
with (out/'compiler-function-inventory.txt').open('w') as f:
 for p in sorted((root/'internal/native').glob('*.go')):
  if p.name.endswith('_test.go'):continue
  for n,l in enumerate(p.read_text().splitlines(),1):
   if l.startswith('func '):f.write(f'{p.relative_to(root)}:{n}: {l}\n')
listed=set(re.findall(r'^Test\w+', (out/'list.log').read_text(),re.M))
assigned=['TestArrayGrowthWitness','TestMarkdownASTPreprocessing','TestWholeDocumentOraclePreflight_Setup','TestMicromarkInputChunks','TestMarkdownSourceDecoding','TestTokenizerEventShardUnion','TestTokenizerEventShardPlantedDisagreement','TestTokenizerEventShardGrowth','TestTokenizerEventsUnion']+[f'TestTokenizerEvents_{i:03d}' for i in range(512)]+['TestFrontMatterStage','TestParserRepresentationProbes','TestMdastIdentifierScalars','TestMarkdownLeafComposition_Setup','TestMarkdownLeafCompositionUnion','TestMarkdownLeafCompositionPlantedDisagreement']+[f'TestMarkdownLeafComposition_{i:03d}' for i in range(7)]
locations={}
for p in (root/'stage1/cohere/markdownblocks').glob('*_test.go'):
 for n,l in enumerate(p.read_text().splitlines(),1):
  m=re.match(r'func (Test\w+)\(',l)
  if m:locations[m[1]]=dict(file=str(p.relative_to(root)),line=n)
(out/'scope.json').write_text(json.dumps(dict(assigned_count=len(assigned),package_count=len(listed),missing=[n for n in assigned if n not in listed],members={n:locations.get(n) for n in assigned}),indent=2))
print('Fixed menu written before mutation execution; inventory',len(inventory),'port functions, scope',len(assigned),'missing',set(assigned)-listed)
