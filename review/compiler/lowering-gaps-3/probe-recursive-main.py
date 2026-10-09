from pathlib import Path
import subprocess, difflib
out=Path('review/compiler/lowering-gaps-3')
p=Path('internal/lower/locals.go'); original=p.read_bytes()
added=['internal/lower/recursive_initializer.go','internal/oracle/recursive_initializer_test.go','internal/oracle/testdata/recursive_initializer_gap.a','internal/oracle/testdata/recursive_initializer_tdz.a']
try:
 for name in added: Path(name).write_bytes(subprocess.check_output(['git','show','bffd95b3a:'+name]))
 s=original.decode().replace('\t\tvar value ir.Expression','\t\tpreallocated := l.result.Locals[local].Preallocated\n\t\tvar value ir.Expression',1)
 s=s.replace('\t\t\tvalue, err = l.expression(initializer)','\t\t\tif list.Flags&ast.NodeFlagsConst != 0 && !l.result.Locals[local].Global && ast.SkipParentheses(initializer).Kind == ast.KindArrowFunction {\n\t\t\t\tvalue, err = l.constantInitializerClosure(local, ast.SkipParentheses(initializer))\n\t\t\t} else {\n\t\t\t\tvalue, err = l.expression(initializer)\n\t\t\t}',1)
 s=s.replace('\t\tif closure, literal := value.(ir.MakeClosure);','\t\tif !preallocated && l.result.Locals[local].Preallocated { statements = append(statements, ir.Declare{Local: local, Uninitialized: true}) }\n\t\tif closure, literal := value.(ir.MakeClosure);',1)
 start=s.index('\tif name, isInitializing := l.initializing[local];')
 end=s.index('\n\tstart := 0',start)
 s=s[:start]+'\tif _, isInitializing := l.initializing[local]; isInitializing { l.result.Locals[local].Preallocated = true }'+s[end:]
 p.write_text(s)
 subprocess.run(['gofmt','-w',str(p)],check=True)
 (out/'recursive-main-port.patch').write_text(''.join(difflib.unified_diff(original.decode().splitlines(True),p.read_text().splitlines(True),fromfile=str(p),tofile=str(p))))
 with (out/'recursive-main-probe.log').open('wb') as log:
  result=subprocess.run(['timeout','120','go','test','./internal/oracle','-run','^TestNativeAgreesWithNode$/internal/oracle/testdata/recursive_initializer_(gap|tdz)\\.a$','-count=1','-timeout=90s','-v'],stdout=log,stderr=subprocess.STDOUT)
 print('recursive main probe exit',result.returncode)
finally:
 p.write_bytes(original)
 for name in added: Path(name).unlink()
