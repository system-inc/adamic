import pathlib,json,re,subprocess,time
p=pathlib.Path('/workspace/adamic/review/test-audit/internal-lower-interface_cast')
rows=['TestDefaultTaggedInterfaceAdmission','TestDefaultTaggedInterfaceNeedsNoFlag','TestIteratorGapsAreExplicit','TestIteratorViewsCannotHideReturn','TestIteratorViewsCannotEraseReceivers','TestIteratorDestructuringDoesNotLieAboutExhaustion','TestGeneratorsAreRefusedEvenWithoutYield','TestLiteralMethodCapturesCannotMakeCycles','TestLiteralMethodViewsDoNotLoseThis','TestIteratorDescriptorReasons','TestRepresentedMethodReplacementIsNotYet','TestDestructuredMethodsCannotLoadOwnSlots','TestGenericIteratorViewsPreserveNativeArguments','TestIteratorMapperIndexHasNumberRepresentation','TestIteratorSymbolKeysAreNotStringKeys']
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
pattern='^('+ '|'.join(rows)+')$'; (p/'scope-pattern.txt').write_text(pattern+'\n')
start=time.monotonic()
with (p/'slice-baseline.log').open('w') as f:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverprofile='+str(p/'scope.cover'),'./internal/lower/','-run',pattern],cwd='/workspace/adamic',stdout=f,stderr=subprocess.STDOUT)
(p/'slice-baseline-time.json').write_text(json.dumps({'seconds':time.monotonic()-start,'exit':r.returncode})+'\n')
if r.returncode:raise SystemExit('slice baseline failed')
r=subprocess.run(['go','tool','cover','-func='+str(p/'scope.cover')],cwd='/workspace/adamic',capture_output=True,text=True,check=True)
(p/'scope-functions-all.txt').write_text(r.stdout)
reached=[line for line in r.stdout.splitlines() if re.search(r'\s(?!0\.0%)\d+\.\d+%$',line) and not line.startswith('total:')]
(p/'code-under-test-functions.txt').write_text('\n'.join(reached)+'\n')
print('reached functions',len(reached))
