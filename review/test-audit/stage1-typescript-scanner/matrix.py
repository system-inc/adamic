exec(open('/tmp/u159/audit.py').read().split('for name,pattern in rows.items():')[0])
meta=json.loads((E/'runs.json').read_text())
mutants=[('M1',P+'/scanner.ts','this.pos += code > 0xffff ? 2 : 1;','this.pos += code > 0xffff ? 1 : 1;','constant'),('M2',P+'/main.ts','        count++;','        count += 2;','constant'),('M3',P+'/characters.ts','else if(code > upper)','else if(code >= upper)','bound'),('M4','internal/lower/diagnostics.go','What: what','What: "unsupported"','option'),('P1',P+'/main.ts','function run(path: string, mode: string, countOnly: boolean): number {','function run(path: string, mode: string, countOnly: boolean): number {\n    return 0;','probe'),('P2','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\tif program != nil { return nil, nil }','probe'),('P3',P+'/scanner_products_test.go','func scannerFetch(t *testing.T, inputs buildcache.Inputs, recipe func(string) error) string {','func scannerFetch(t *testing.T, inputs buildcache.Inputs, recipe func(string) error) string {\n\tif t != nil { return "" }','construction-probe'),('S1',P+'/scanner_products_test.go','\tprepared.once.Do(func() { prepared.directory = buildcache.Product(t, inputs, recipe) })','\t// audit: dropped the product construction statement','construction'),('S2',P+'/profile_test.go','"-g", "-o"','"-audit-invalid-profile-option", "-o"','construction'),('W1',P+'/scanner_agreement_shards_test.go','if bytes.Equal(got, want) {','if true {','weakened-check')]
(E/'menu.json').write_text(json.dumps(mutants,indent=2))
(E/'diffs').mkdir(exist_ok=True)
for id,path,old,new,kind in mutants:
 f=R/path; original=f.read_text();assert original.count(old)==1,(id,original.count(old));changed=original.replace(old,new)
 line=original[:original.index(old)].count('\n')+1
 (E/'diffs'/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),'a/'+path,'b/'+path)))
 f.write_text(changed)
 try:
  if path.endswith('.go'):
   with open(E/(id+'-vet.log'),'w') as out:check=subprocess.run(['go','vet','./'+str(pathlib.Path(path).parent)+'/'],cwd=R,env=env,stdout=out,stderr=subprocess.STDOUT)
   if check.returncode:print(id,'VET FAILED',flush=True);continue
  pattern='.'
  if id=='P2':pattern='^(TestGapStandsWhereGapsMdSays|TestBigintGapStandsWhereGapsMdSays)$'
  if id in ('P3','S1'):pattern='^TestProduct_Scanner'
  if id=='S2':pattern='^TestProfileArtifacts$'
  if id=='W1':pattern='^(TestScannerAgreesWithTypescriptGo_|TestScannerShardCoverage$)'
  run(id,pattern,id)
 finally:f.write_text(original)
print('RESTORED',flush=True)
