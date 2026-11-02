import pathlib,shutil,subprocess
b=pathlib.Path('/workspace/scratch'); tree=b/'99d-main-after'; oracle=b/'99d-main-oracle-tree'
shutil.copytree(tree,oracle)
p=oracle/'src/compiler/commandLineParser.ts'
s=p.read_bytes(); needle=b'catch (e) {\r\n        return createCompilerDiagnostic(Diagnostics.Cannot_read_file_0_Colon_1'
assert s.count(needle)==1
counter=b'catch (e) {\r\n        require("fs").appendFileSync("/workspace/scratch/99d-main-catch.jsonl", JSON.stringify([typeof e, typeof e?.message]) + "\\n");\r\n        return createCompilerDiagnostic(Diagnostics.Cannot_read_file_0_Colon_1'
p.write_bytes(s.replace(needle,counter))
(b/'99d-main-catch.jsonl').write_text('')
with (b/'99d-main-oracle.log').open('w') as log:
 result=subprocess.run(['bash','stage3/oracle/run.sh',str(oracle),str(b/'99d-main-oracle')],stdout=log,stderr=subprocess.STDOUT)
print(result.returncode)
