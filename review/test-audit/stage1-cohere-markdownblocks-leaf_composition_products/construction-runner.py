from pathlib import Path
import subprocess,json,time,difflib
out=Path('review/test-audit/stage1-cohere-markdownblocks-leaf_composition_products');root=Path('stage1/cohere/markdownblocks')
cases=[
('S1','leaf_composition_independent_test.go','return filepath.Join(product, "oracle")','return filepath.Join(product, "")','^TestProduct_MarkdownLeafGo(Lists|DocLayout)$'),
('S2','leaf_composition_products_test.go','\tprepareLeafCompositionLowered(t, &p)\n','','^TestProduct_MarkdownLeafLowered$'),
('S3','leaf_composition_independent_test.go','return filepath.Join(directory, "port")','return filepath.Join(directory, "")','^TestProduct_MarkdownLeafNative(Sanitized|Release)$'),
('S4','list_layout_shards_test.go','sanitized: manifest.Paths[2]','sanitized: ""','^(TestMarkdownListLayout_Setup|TestMarkdownListLayout)$'),
('S5','quote_layout_shards_test.go','\t\t\tproducts.goBinary = binary\n','','^TestMarkdownQuoteLayout$'),
('S6','lists_test.go','return inputs, files','return inputs[:0], files','^TestMarkdownTableLayout$'),
('S7','leaf_composition_independent_test.go','\t\tleafCompositionPrepared, leafCompositionBuilds = fixture, products\n','','^TestMarkdownLeafComposition$'),
('S8','structure_layout_shards_test.go','\t\tstructureLayoutComplete = state\n','','^TestMarkdownStructureLayout_Setup$'),
('W1','lists_test.go','if bytes.Equal(result.stdout, want.stdout) {','if true {','^(TestMarkdownCodeBlockLayout|TestMarkdownHTMLBlockLayout|TestMarkdownRootLayout)$'),
('W2','list_layout_shards_test.go','if bytes.Equal(result.stdout, fixture.mutantWant) {','if true {','^TestMarkdownListLayout(Union|_00[0-9]|_01[0-5])$'),
('W3','malformed_events_independent_test.go','if bytes.Equal(mutant.stderr, []byte("adamic: panic: "+string(truth.stdout))) {','if true {','^TestMdastMalformedEvents(_00[0-2]|Union)$')]
res=[]
for mid,f,a,b,pattern in cases:
 p=root/f;old=p.read_text();assert old.count(a)==1,(mid,old.count(a));new=old.replace(a,b)
 (out/f'{mid}.diff').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+str(p),tofile='b/'+str(p))))
 try:
  p.write_text(new)
  started=time.monotonic()
  with (out/f'{mid}-vet.log').open('w') as log:v=subprocess.run(['go','vet','./stage1/cohere/markdownblocks/'],stdout=log,stderr=subprocess.STDOUT)
  with (out/f'{mid}.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',pattern],stdout=log,stderr=subprocess.STDOUT)
  res.append({'id':mid,'file':str(p),'line':old[:old.index(a)].count('\n')+1,'change':[a,b],'pattern':pattern,'vet':v.returncode,'exit':r.returncode,'wall':time.monotonic()-started});(out/'construction-runs.json').write_text(json.dumps(res,indent=2))
 finally:p.write_text(old)
