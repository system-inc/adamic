import pathlib,shutil,subprocess,os,json
repo=pathlib.Path(__file__).resolve().parents[3];scratch=pathlib.Path('/tmp/skip-census-source-copy');scratch.mkdir(exist_ok=True)
files=subprocess.check_output(['git','ls-files','*_test.go'],cwd=repo,text=True).splitlines()
files += [str(p.relative_to(repo)) for p in (repo/'internal/skipcensus').rglob('*_test.go')]
for name in files:
 p=scratch/name;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(repo/name,p)
(scratch/'internal/skipcensus/new_skip_test.go').write_text('package skipcensus\nimport "testing"\nfunc TestUndeclaredWitness(t *testing.T) { t.Skip("the gate forgot this input") }\n')
env=dict(os.environ,ADAMIC_SKIP_CENSUS_ROOT=str(scratch))
with open('/tmp/skip-census-new-skip-mutant.log','w') as log:
 result=subprocess.run(['/tmp/skip-census.test','-test.run=^TestCensus$','-test.v'],cwd=repo/'internal/skipcensus',env=env,stdout=log,stderr=subprocess.STDOUT)
assert result.returncode==1
text=pathlib.Path('/tmp/skip-census-new-skip-mutant.log').read_text();assert 'undeclared skip internal/skipcensus/new_skip_test.go:TestUndeclaredWitness:' in text
print('scratch new t.Skip: exit 1; TestCensus named TestUndeclaredWitness')
mutants=[('allow-required','log.go','if required+unknown > 0 {','if unknown > 0 {','TestLogClassesAndMutants'),('allow-added','census.go','errors = append(errors, fmt.Sprintf("undeclared skip %s (%s:%d)", k, r.Test, r.Line))','// mutant allows added skip','TestASTAndDriftMutants'),('allow-removed','census.go','errors = append(errors, "removed skip "+k)','_ = k // mutant allows removed skip','TestASTAndDriftMutants')]
for name,file,old,new,test in mutants:
 root=pathlib.Path('/tmp/skip-census-mutant-'+name);root.mkdir(exist_ok=True)
 for p in (repo/'internal/skipcensus').glob('*.go'):shutil.copyfile(p,root/p.name)
 (root/'go.mod').write_text('module mutant\ngo 1.27\n')
 p=root/file;source=p.read_text();assert source.count(old)==1;(root/file).write_text(source.replace(old,new,1))
 with open('/tmp/skip-census-'+name+'-mutant.log','w') as log:
  result=subprocess.run(['go','test','-v','-count=1','-run=^'+test+'$','.'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
 text=pathlib.Path('/tmp/skip-census-'+name+'-mutant.log').read_text();assert result.returncode==1 and '--- FAIL: '+test in text,text
 print(name+': exit 1; '+test+' caught the implementation mutant')
with open('/tmp/skip-census-plain-check.log','w') as log:
 result=subprocess.run(['/tmp/skip-census-command','/tmp/skip-census-plain/gate-out/test.jsonl'],cwd=repo,stdout=log,stderr=subprocess.STDOUT)
text=pathlib.Path('/tmp/skip-census-plain-check.log').read_text();assert result.returncode==1;assert 'skips=33 required-input=17 unknown=0' in text
required=[line.split('\t')[2] for line in text.splitlines() if line.startswith('required-input\t')]
expected=['TestSplitTSGoAgrees','TestThePortParsesAsGoCohereDoes/PostCSS','TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower','TestThePortParsesAsGoCohereDoes/as_graphql-js','TestUpstreamNumericSeparatorGap','TestExternalComparisonCatchesThreePrinterMutants','TestUpstreamRepositoryCorpusParity','TestCompilerAndStage1Agree','TestCSSPrinterAgreesWithGo/default','TestCSSPrinterAgreesWithGo/narrow','TestCSSPrinterBoundaryProofs','TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser','TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars','TestThePortParsesAsGoCohereDoes/as_postcss-selector-parser','TestThePortParsesAsGoCohereDoes/as_postcss-values-parser','TestCompilerExpressionsAgree','TestWholeCompilerAgrees']
assert sorted(required)==sorted(expected),required
print('historical plain log: exit 1; all 17 exact expected required-input skips; TestWholeCompilerAgrees included')
