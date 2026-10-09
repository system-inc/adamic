import pathlib,json,subprocess,time
rows=['TestOct6InheritanceMutants','TestOct6ReleaseMutant','TestOmittedArgumentZeroMutantIsCaught','TestOmittedOriginalProbePolicy','TestOmittedReaderZeroMutantIsCaught','TestNativeAgreesWithNode','TestTheOracleCatchesOneByte','TestOneFileHoldsNodesOrder','TestClosedStdoutEndsAsOnNode','TestFileWritesLandInNodesOrder','TestAPromptComesBeforeTheRead','TestASignalLeavesWhatWasPrinted','TestOverloadContractRulings','TestParameterPropertyMutants','TestParameterPropertyOwnershipMutant']
r=pathlib.Path('review/test-audit/internal-oracle-oct6_mutant');listed=pathlib.Path('/tmp/u066/list.log').read_text().splitlines();assert all(x in listed for x in rows)
fixtures=['interleaved.a','status_of_stdout.a','large_output.a','output_then_panic.a','write_stdout_order.a','write_stderr_order.a','prompt_then_read.a','killed_after_output.a','omitted_scanner.a','omitted_reader.a','class_oct6_deep.a','class_oct6_release.a','parameter_properties.a','parameter_properties_ownership.a']
regex='^('+'|'.join(x for x in rows if x!='TestNativeAgreesWithNode')+')$';nr='^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^('+'|'.join(x.replace('.','\\.') for x in fixtures)+')$'
(r/'scope.json').write_text(json.dumps(dict(base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),rows=rows,regex=regex,native_regex=nr,native_subcases=fixtures,nproc=5),indent=2))
start=time.monotonic()
for name,run in [('bounded-baseline',regex),('native-baseline',nr)]:
 with open('/tmp/u066/'+name+'.log','w') as out:p=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=github.com/system-inc/adamic/internal/oracle,github.com/system-inc/adamic/internal/lower,github.com/system-inc/adamic/internal/native','-coverprofile=/tmp/u066/'+name+'.cov','./internal/oracle/','-run',run],stdout=out,stderr=subprocess.STDOUT)
 print(name,p.returncode,flush=True)
 if p.returncode:break
pathlib.Path('/tmp/u066/narrow-wall.txt').write_text(str(time.monotonic()-start))
