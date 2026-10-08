from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
expression=root/'internal/lower/expression.go'
statements=root/'internal/lower/statements.go'
cases=[('skip-call-boxing',expression,'arguments[index] = fit(arguments[index], takes)','if takes != ir.Union { arguments[index] = fit(arguments[index], takes) }','call boundary argument is not boxed'),('skip-return-boxing',statements,'return []ir.Statement{ir.Return{Value: fit(value, l.function.Returns)}}, nil','if l.function.Returns == ir.Union { return []ir.Statement{ir.Return{Value: value}}, nil }; return []ir.Statement{ir.Return{Value: fit(value, l.function.Returns)}}, nil','return boundary value is not boxed')]
for name,path,old,new,caught in cases:
 original=path.read_text()
 assert original.count(old)==1
 try:
  path.write_text(original.replace(old,new))
  log=Path('/tmp/function-values-boundary-mutant-'+name+'.log')
  with log.open('w') as out:
   result=subprocess.run(['go','test','./internal/lower','-run','^TestFunctionValueBoundaryBoxing$','-count=1','-timeout','10m'],cwd=root,stdout=out,stderr=subprocess.STDOUT)
  output=log.read_text()
  assert result.returncode!=0 and caught in output and '[build failed]' not in output
  print(name,'exit',result.returncode,'caught by fixture representation assertion:',log,flush=True)
 finally:
  path.write_text(original)
