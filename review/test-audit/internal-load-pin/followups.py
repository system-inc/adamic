import pathlib,subprocess,time,json
out=pathlib.Path('review/test-audit/internal-load-pin')
f=pathlib.Path('internal/load/load.go'); original=f.read_text()
try:
 subprocess.run(['git','apply',str(out/'P1.diff')],check=True)
 for test in ['TestTypeScriptIsThePinnedCommit','TestRegExpCaptureTypes']:
  with (out/('P1-'+test+'.log')).open('w') as log:
   subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/load/','-run','^'+test+'$'],stdout=log,stderr=subprocess.STDOUT)
finally: f.write_text(original)
witness=pathlib.Path('internal/load/audit_u025_witness_test.go')
witness.write_text('''package load
import ("testing"; "strings"; "github.com/microsoft/TypeScript/tsc/shim/bundled"; "github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs")
func TestAuditU025DeclarationWitness(t *testing.T) {
 fs := &regexpLibraryFS{FS:bundled.WrapFS(osvfs.FS())}
 text,ok := fs.ReadFile(bundled.LibPath().ResolveFile("lib.es5.d.ts"))
 t.Logf("read=%t explicit RegExp split returns undefined union=%t",ok,strings.Contains(text,"split(separator: RegExp, limit?: number): (string | undefined)[];"))
}
''')
f=pathlib.Path('internal/load/regexp_library.go'); original=f.read_text()
try:
 for name in ['original','M3']:
  if name=='M3': subprocess.run(['git','apply',str(out/'M3.diff')],check=True)
  with (out/('survivor-'+name+'.log')).open('w') as log:
   subprocess.run(['timeout','120','go','test','-v','-count=1','-timeout','90s','./internal/load/','-run','^TestAuditU025DeclarationWitness$'],stdout=log,stderr=subprocess.STDOUT)
finally:
 f.write_text(original); witness.unlink()
with (out/'coverage.log').open('w') as log:
 subprocess.run(['timeout','120','go','test','-count=1','-timeout','90s','-coverprofile='+str(out/'regexp.cover'),'./internal/load/','-run','^TestRegExpCaptureTypes$'],stdout=log,stderr=subprocess.STDOUT)
with (out/'coverage-functions.txt').open('w') as log: subprocess.run(['go','tool','cover','-func='+str(out/'regexp.cover')],stdout=log,stderr=subprocess.STDOUT)
