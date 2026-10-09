from pathlib import Path
import json,subprocess,os
root=Path.cwd();scratch=Path('/tmp/adamic-stable-splitter-mutants');scratch.mkdir(exist_ok=True)
mutants=[('content-owner','units_stable.go','} else if !d.function && (strings.HasPrefix(d.name, "adamic_string_")','} else if false && !d.function && (strings.HasPrefix(d.name, "adamic_string_")','^TestSharedIdentitiesDoNotMigrate$','shared identity migrated'),('helper-owner','units_stable.go','item.owner = fmt.Sprintf("helpers_%02d.c", int(digest[0])%16)','_ = digest','^TestSharedIdentitiesDoNotMigrate$','shared identity migrated'),('positional-grouping','units.go',None,None,'^TestStableUnitInsertion$','insertion changed'),('flags-key','units.go','append([]string{"adamic-units-v3"}, flags...)','[]string{"adamic-units-v3"}','^TestUnitCacheFlagsHoldSanitizer$','sanitized rebuild reused uninstrumented object'),('abi-relocation','units_stable.go','guards.WriteString("};\\n")','guards.Reset(); guards.WriteString("\\n")','^TestUnitDeclarationDisagreement$','declaration mutant silently linked'),('cross-module-local','units_stable.go','if seenLocals[token.text] {','if false {','^TestModuleMainRejectsCrossRangeLocal$','cross-range local accepted')]
mutants.extend([
 ('constant-owner','units_stable.go','strings.Contains(" "+declaration, " const ")','false','^TestConstantDescriptorDoesNotMigrate$','constant descriptor migrated'),
 ('adapter-owner','units_stable.go','if len(owners) == 1 {','if false {','^TestDeclarationAdapterFollowsCallee$','adapter did not follow'),
 ('forwarder-owner','units_stable.go','} else if !d.function && strings.HasPrefix(d.name, "adamic_global_") {','} else if false && !d.function && strings.HasPrefix(d.name, "adamic_global_") {','^TestForwarderGlobalsDoNotMigrate$','forwarder global migrated'),
 ('declaration-key','units_stable.go','name + "\\x00" + declaration','name','^TestUnitDeclarationDisagreement$','declaration mutant silently linked'),
])
mutants.append(('whole-program-temporaries','emit.go','\te.temporaries = 0\n','','^TestStableLintUnitChanges$','want only grammar unit'))
for label,file,old,new,pattern,expected in mutants:
 source=(root/'internal/native'/file).read_text()
 if old is None: source=subprocess.check_output(['git','show','aca41891:internal/native/units.go'],cwd=root,text=True)
 else:
  assert old in source,label
  source=source.replace(old,new) if label=="helper-owner" else source.replace(old,new,1)
 replacement=scratch/(label+'.go');replacement.write_text(source)
 overlay=scratch/(label+'.json');overlay.write_text(json.dumps({'Replace':{str(root/'internal/native'/file):str(replacement)}}))
 log=scratch/(label+'.log')
 with log.open('w') as f: result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run',pattern,'-count=1','-v'],cwd=root,stdout=f,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_STABLE_LINT_PROBE':'1','ADAMIC_STABLE_READ_INPUTS':'0','ADAMIC_STABLE_C_INPUTS':''})
 assert result.returncode and expected in log.read_text(),(label,result.returncode,log.read_text())
 print(label,'caught',flush=True)
