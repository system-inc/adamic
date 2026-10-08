"""Build valid callback-reference mutants and require semantic control failures."""
import pathlib,subprocess
root=pathlib.Path.cwd();scratch=pathlib.Path('/tmp/shape-callback-join-mutants');scratch.mkdir(exist_ok=True)
source=(root/'stage3/shape-conformance/latent/lower.go.txt').read_text()
mutants={
 'drop-conditional-reference':('transparent = conditional.WhenTrue == node || conditional.WhenFalse == node','transparent = false'),
 'drop-parenthesized-reference':('transparent = parent.AsParenthesizedExpression().Expression == node','transparent = parent.AsParenthesizedExpression().Expression == node && false'),
 'drop-asserted-reference':('transparent = parent.AsAsExpression().Expression == node','transparent = parent.AsAsExpression().Expression == node && false'),
 'drop-generic-reference':('function,ok:=m.Functions[declaration];ok && declaration.Kind==ast.KindFunctionDeclaration && !function.Skipped {','function,ok:=m.Functions[declaration];ok && declaration.Kind==ast.KindFunctionDeclaration && !function.Skipped && len(declaration.TypeParameters())==0 {'),
}
for name,(old,new) in mutants.items():
 assert source.count(old)==1,(name,source.count(old))
 changed=source.replace(old,new)
 # Keep the conditional local used so the mutant is valid Go.
 if name=='drop-conditional-reference':changed=changed.replace('transparent = false\n        }','transparent = conditional.WhenTrue == node && false\n        }')
 template=scratch/(name+'.go.txt');template.write_text(changed)
 overlay=scratch/name;binary=scratch/(name+'-binary')
 with (root/'stage3/shape-conformance/overnight'/(name+'.log')).open('w') as log:
  assert subprocess.run(['python3','stage3/shape-conformance/latent/make-overlay.py',str(overlay),str(template)],stdout=log,stderr=subprocess.STDOUT).returncode==0
  assert subprocess.run(['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(binary),'./stage3/shape-conformance/latent/tool'],stdout=log,stderr=subprocess.STDOUT).returncode==0,name+' did not compile'
  status=subprocess.run(['python3','stage3/shape-conformance/callback-joins/controls.py',str(binary)],stdout=log,stderr=subprocess.STDOUT).returncode
  assert status != 0,name+' escaped'
 print(name+': valid measurement binary; semantic controls caught mutant',flush=True)
