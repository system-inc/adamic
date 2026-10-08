"""Run semantic/proof mutants serially; always restore production sources."""
from pathlib import Path
import json, subprocess, time
cases = [
 ('erase-without-proof', 'internal/native/view_writes.go', 'if !write.WriteProven {', 'if false {', './internal/oracle', '^TestOptionalCheckedWrites/sentinel$', 'differs'),
 ('skip-shape-check', 'internal/native/runtime/object.c', 'void adamic_object_checked_write(adamic_object *object, const char *name, adamic_slot_cache *cache, const unsigned int *allowed, size_t count, const char *where) {', 'void adamic_object_checked_write(adamic_object *object, const char *name, adamic_slot_cache *cache, const unsigned int *allowed, size_t count, const char *where) {\n if (object != NULL) return;', './internal/oracle', '^TestOptionalCheckedWrites/stored$', 'differs'),
 ('stored-is-fresh', 'internal/fresh/fresh.go', 'recorded.Unaliased = recorded.Unaliased && a.unaliased(holder)', 'recorded.Unaliased = true', './internal/lower', '^TestOptionalWriteErasure/stored$', 'stored literal was counted unaliased'),
 ('judge-unreduced-source', 'internal/lower/optional_widening.go', 'source = reducedOptionalSource(l.checker, source)', '// MUTANT: source reduction erased', './internal/lower', '^TestOptionalWideningReducedSource$', 'reduced inhabited source needs no optional view'),
 ('object-covariance', 'internal/ir/view_writes.go', 'if !field.Readonly && (own.Readonly || !assignable(field.Contract, own.Contract, seen)) {', 'if false {', './internal/oracle', '^TestOptionalCheckedWrites/object-invariant$', 'differs'),
 ('ignore-nominal-slot', 'internal/ir/view_writes.go', 'if target.Nominal != "" && source.Nominal != target.Nominal {', 'if false {', './internal/oracle', '^TestOptionalCheckedWrites/nominal-slot$', 'exit codes differ'),
 ('forget-call-kills-tag', 'internal/lower/view_write_erasure.go', 'clear(facts)', '// MUTANT: retain stale tag facts', './internal/lower', '^TestOptionalWriteErasure/tag-call$', 'erasure=true, want false'),
 ('direct-subclasses-only', 'internal/lower/optional_widening.go', 'if found := l.optionalClassBase(base, source, seen); found != nil {', 'if found := base; found.Symbol() == source {', './internal/oracle', '^TestOptionalClassReads/class_transitive.a$', 'differs'),
 ('never-is-no-op', 'internal/lower/never.go', 'body = append(body, ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}})', 'body = append(body, ir.Return{Value: ir.NumberConstant{Value: float64(len(message))*0}})', './internal/oracle', '^TestNeverReached$', 'exit codes differ'),
 ('omit-compound-contract', 'internal/lower/class.go', 'if l.result.OptionalViewFields[name] {', 'if false {', './internal/oracle', '^TestOptionalCheckedWrites/compound-literal$', 'differs'),
 ('virtual-base-only', 'internal/lower/view_write_erasure.go', 'range program.CallTargets(value)', 'range []int{0}', './internal/lower', '^TestOptionalVirtualReceiverIsNotUnreachable$', 'a returning override was treated as unreachable'),
 ('close-numeric-enum-slot', 'internal/lower/view_contracts.go', 'if l.openNumericEnumType(target) {', 'if false {', './internal/oracle', '^TestOptionalCheckedWrites/enum-slot$', 'exit codes differ'),
]
results=[]
for name,path,old,new,package,test,expected in cases:
 p=Path(path); original=p.read_text(); assert old in original, (name,path)
 log=Path('/tmp/optional-write-mutant-'+name+'.log'); start=time.monotonic()
 try:
  p.write_text(original.replace(old,new))
  with log.open('w') as out:
   result=subprocess.run(['go','test',package,'-run',test,'-count=1','-timeout','5m'],stdout=out,stderr=subprocess.STDOUT)
  body=log.read_text()
  assert result.returncode and '--- FAIL' in body and expected in body, (name,body)
  assert not any(x in body for x in ['build failed','compiler bug','runtime error:','AddressSanitizer:','error: unknown','error: incompatible','clang failed','compiling runtime']), (name,body)
  results.append({'mutant':name,'test':test,'caught':expected,'seconds':round(time.monotonic()-start,3),'log':str(log)})
  print(name, 'caught', flush=True)
 finally:
  p.write_text(original)
Path('/tmp/optional-write-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
