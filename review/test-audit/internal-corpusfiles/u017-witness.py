import pathlib,subprocess,os,json,time
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/internal-corpusfiles';p=r/'internal/corpusfiles/files.go';base=p.read_text();w=r/'internal/corpusfiles/u017_witness_test.go'
source='''package corpusfiles
import "testing"
func TestU017Observation(t *testing.T) {
 t.Logf("names=%q", names([]byte("tracked.ts\\x00")))
 d, _, _ := fixture(t, "inputs", ".ts")
 files, _, counts, err := selectFiles(d, "", []string{"inputs"}, []string{"*.ts"})
 t.Logf("files=%d counts=%v err=%v",len(files),counts,err)
}
'''
(e/'survivor-witness.go.txt').write_text(source);env=os.environ.copy();env.update(GOWORK='off',ADAMIC_CSS_FIXTURES='/tmp/u017-prettier');meta=json.loads((e/'menu.json').read_text())
try:
 w.write_text(source)
 for mid in ['clean','M01','M13']:
  m=next((x for x in meta if x['id']==mid),None);p.write_text(base if m is None else base.replace(m['old'],m['new'],1))
  with (e/('witness-'+mid+'.log')).open('w') as f:subprocess.run(['go','test','-v','-count=1','./internal/corpusfiles/','-run','^TestU017Observation$'],cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
finally:p.write_text(base);w.unlink()
with (e/'final-baseline.log').open('w') as f:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/corpusfiles/','-run','.'],cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
