from pathlib import Path
import os,subprocess
root=Path(__file__).resolve().parents[3]
p=root/'internal/lower/expression.go'
original=p.read_text()
start=original.index('func (l *lowering) functionValue(')
end=original.index('\n// optionalCall',start)
body=original[start:end]
cases=[('return', '\t\tforwarder.Body = []ir.Statement{ir.Return{Value: call}}', '\t\tvar returned ir.Expression = call\n\t\tif callee.Returns == ir.Union {\n\t\t\treturned = fit(ir.NumberConstant{Value: 99}, ir.Union)\n\t\t}\n\t\tforwarder.Body = []ir.Statement{ir.Return{Value: returned}}', 'return'),('arguments','\t\targuments = append(arguments, ir.Read{Local: local, Of: declared.Type})','\t\tvar argument ir.Expression = ir.Read{Local: local, Of: declared.Type}\n\t\tif declared.Type == ir.Union {\n\t\t\targument = fit(ir.StringConstant{Index: l.constant("mutant")}, ir.Union)\n\t\t}\n\t\targuments = append(arguments, argument)', '(options|signature|chain|diagnostic|array)')]
for name,old,new,pattern in cases:
 assert body.count(old)==1
 try:
  p.write_text(original[:start]+body.replace(old,new)+original[end:])
  log=Path('/tmp/function-values-mutant-'+name+'.log')
  with log.open('w') as output:
   result=subprocess.run(['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/function_values_'+pattern,'-count=1','-timeout','10m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  text=log.read_text()
  print(name,'exit',result.returncode,flush=True)
  assert result.returncode!=0 and 'FAIL: TestNativeAgreesWithNode/' in text
  assert 'clang:' not in text and 'Lower:' not in text and '[build failed]' not in text
  if name=='arguments':
   for fixture in ['options','signature','chain','diagnostic','array']:
    assert 'FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/function_values_'+fixture+'.a' in text,fixture
  print('caught by behavioral oracle:',log,flush=True)
 finally:
  p.write_text(original)

try:
    old = "return !(fromKept == ir.Union && !toKept.IsReference() || toKept == ir.Union && !fromKept.IsReference()) && same(inside, viewed)"
    assert original.count(old) == 1
    # Keep the reads live so this mutant compiles without an unused-variable error.
    p.write_text(original.replace(old, "_ = fromKept\n\t\t\t_ = toKept\n\t\t\treturn same(inside, viewed)"))
    log = Path("/tmp/function-values-mutant-views.log")
    with log.open("w") as output:
        result = subprocess.run(["go", "test", "./internal/lower", "-run", "TestFunctionValueUnionViewsStayNotYet", "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    text = log.read_text()
    assert result.returncode != 0 and text.count("callable view requires a union slot adapter, got <nil>") == 4
    print("views exit", result.returncode, "caught all four unsafe callable views:", log, flush=True)
finally:
    p.write_text(original)
