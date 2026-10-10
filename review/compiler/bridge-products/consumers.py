import pathlib,json,re,subprocess,os,time,difflib
root=pathlib.Path.cwd();evidence=root/'review/compiler/bridge-products'
exec((evidence/'run.py').read_text().split('for suffix, product in pairs:')[0])
rows=json.loads((evidence/'results.json').read_text())
base['ADAMIC_UNIT_BUDGET']='1'
consumer_names=['TestBridgeABI','TestBridgeInputLength','TestBridgeUnlinkedBuild','TestBridgeUnlinkedC','TestBridgeUnlinkedJavaScript','TestBridgeOutputLength','TestBridgeStaleHandle','TestBridgeLinkage','TestBridgeOutputFree','TestBridgeRegion','TestBridgeRegionOwnership','TestBridgeOracleSample','TestBridgeTimingRound1Sample','TestBridgeTimingRound2Sample','TestBridgeTimingRound3Sample','TestBridgeWrongPositionSample']
for name in consumer_names:
 run('consumer-'+name,name)
 lines=(evidence/('consumer-'+name+'.builds.txt')).read_text().splitlines()
 assert lines and all(line.split()[3]=='hit' for line in lines), name
# One cold process per input proof on the final source, including the extra
# key comparison that also checks deliberate ADAMIC_BUILD_CACHE=off runs.
for suffix, product in pairs: run('final-input-'+suffix,'TestBridgeProductInput_'+suffix)
# Check the actual compiler corpus, the load behind the pool's slow units.
base['ADAMIC_TSGO_CORPUS']='/workspace/test-split-typescript'
corpus_head=subprocess.check_output(['git','-C',base['ADAMIC_TSGO_CORPUS'],'rev-parse','HEAD'],text=True).strip()
(evidence/'corpus.txt').write_text(base['ADAMIC_TSGO_CORPUS']+'\n'+corpus_head+'\n')
for name in ['TestBridgeOracleChecker','TestBridgeTimingRound1Checker']:
 run('corpus-'+name,name)
 lines=(evidence/('corpus-'+name+'.builds.txt')).read_text().splitlines()
 assert lines and all(line.split()[3]=='hit' for line in lines), name
run('two-fresh-corpus-leaves','(?:TestBridgeOracleChecker|TestBridgeTimingRound1Checker)')
lines=(evidence/'two-fresh-corpus-leaves.builds.txt').read_text().splitlines()
assert lines and all(line.split()[3]=='hit' for line in lines)
# The count assertion itself has a planted failure.
source=root/'bridge/tsgo/product_inputs_test.go';text=source.read_text()
mutant=re.sub(r'^\s*"native":\s*TestProduct_Native,\n','\n',text,flags=re.M)
assert mutant!=text
replacement=scratch/'product-count.go';replacement.write_text(mutant)
overlay=scratch/'product-count.json';overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
(evidence/'product-count.patch').write_text(''.join(difflib.unified_diff(text.splitlines(True),mutant.splitlines(True),fromfile='a/bridge/tsgo/product_inputs_test.go',tofile='b/bridge/tsgo/product_inputs_test.go',n=0)))
run('product-count-mutant','TestBridgeProductUnitsCoverEveryProduct',overlay=overlay,expected=1)
assert '19 products, 18 units, want 19' in (evidence/'product-count-mutant.jsonl').read_text()
