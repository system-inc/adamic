import json, os, pathlib, subprocess, sys, tempfile, time
root=pathlib.Path(__file__).resolve().parents[4]
cases=[
 ('symbol-presence','local_symbol_facts.go','symbol := c.GetSymbolAtLocation(node)','symbol := (*ast.Symbol)(nil)','TestNarrowSymbolContracts'),
 ('shorthand-binding-identity','local_symbol_facts.go','value := c.GetShorthandAssignmentValueSymbol(property)','value := c.GetSymbolAtLocation(node)','TestNarrowSymbolContracts'),
 ('local-declaration-filter','local_symbol_facts.go','ast.GetSourceFileOfNode(declaration) == source','ast.GetSourceFileOfNode(declaration) != source','TestNarrowSymbolContracts'),
 ('first-declaration-file-flag','local_symbol_facts.go','out.yes(file.IsDeclarationFile)','out.yes(!file.IsDeclarationFile)','TestNarrowSymbolContracts'),
 ('ordered-output-file-flags','output_symbol.go','out.yes(source.IsDeclarationFile)','out.yes(!source.IsDeclarationFile)','TestNarrowSymbolContracts'),
 ('library-member-provenance','library_member_provenance.go','out.yes(p.Compiler.IsSourceFileDefaultLibrary(file.PathKey()))','out.yes(!p.Compiler.IsSourceFileDefaultLibrary(file.PathKey()))','TestLibraryProvenanceMatchesChecker'),
 ('symbol-default-library','default_library_facts.go','out.yes(present)','out.yes(!present)','TestLibraryProvenanceMatchesChecker'),
 ('narrow-alias-identity','local_symbol_facts.go','target = c.GetAliasedSymbol(target)','target = symbol','TestAliasFactsMatchChecker'),
 ('local-skip-alias','local_symbol_facts.go','symbol = checker.SkipAlias(symbol, c)','symbol = c.GetSymbolAtLocation(node)','TestAliasFactsMatchChecker'),
 ('skip-alias-identity','local_symbol_facts.go','target = checker.SkipAlias(target, c)','target = symbol','TestAliasFactsMatchChecker'),
 ('skip-alias-target','binding_declarations.go','symbol = checker.SkipAlias(symbol, c)','symbol = c.GetSymbolAtLocation(node)','TestAliasFactsMatchChecker'),
 ('resolved-signature-kind','resolved_signature.go','out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))','out.text("VariableDeclaration")','TestResolvedSignatureMatchesChecker'),
]
results=[]
for name, filename, before, after, test in cases:
 original=root/'bridge/tsgo/checker'/filename
 text=original.read_text()
 expected = 3 if name=='local-declaration-filter' else 1
 if text.count(before)!=expected: raise RuntimeError(f'{name}: expected {expected} mutation anchors, got {text.count(before)}')
 directory=pathlib.Path(tempfile.mkdtemp(prefix='checker-mutant-',dir='/tmp'))
 altered=directory/filename; altered.write_text(text.replace(before,after,expected))
 overlay=directory/'overlay.json'; overlay.write_text(json.dumps({'Replace':{str(original):str(altered)}}))
 log=pathlib.Path('/tmp')/f'checker-narrow-mutant-{name}.log'
 start=time.monotonic()
 with log.open('wb') as output:
  done=subprocess.run(['go','test','-count=1','-timeout=2m','-overlay',str(overlay),'./bridge/tsgo/checker','-run',f'^{test}$','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
 report=log.read_text()
 if done.returncode==0 or 'build failed' in report or 'panic:' in report or '--- FAIL: '+test not in report:
  raise RuntimeError(f'{name}: mutant was not killed by the checker comparison; see {log}')
 results.append({'name':name,'test':test,'exit':done.returncode,'wall':time.monotonic()-start,'log':str(log)})
 print(name,'caught by',test,flush=True)
pathlib.Path('/tmp/checker-narrow-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
