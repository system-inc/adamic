from pathlib import Path
import difflib
import subprocess

root = Path.cwd()
source = root / 'internal/lower/element_access_fields.go'
original = source.read_text()
mutated = original.replace('func (l *lowering) refuseViewElementReads(fields map[string]bool) error {', 'func (l *lowering) refuseViewElementReads(fields map[string]bool) error {\n fields = nil // mutant: route element reads around the refusal', 1)
evidence = root / 'review/compiler/fx6-key-read'
(evidence / 'bypass-refusal.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile='a/internal/lower/element_access_fields.go', tofile='b/internal/lower/element_access_fields.go')))
probe = root / 'internal/lower/fx6_mutant_probe_test.go'
probe_text = '''package lower
import ("os"; "path/filepath"; "testing"; "github.com/system-inc/adamic/internal/native")
func TestFX6NativeMutantWitness(t *testing.T) {
 t.Parallel()
 source := viewElementFixture(t, "p05")
 path := filepath.Join(t.TempDir(), "main.a")
 if err := os.WriteFile(path, []byte(source), 0644); err != nil { t.Fatal(err) }
 program, err := lowerSource(t, source)
 if err != nil { t.Fatal(err) }
 want := runAgreementNode(t, path)
 got := runAgreementNative(t, native.C(program))
 t.Logf("source Node: exit=%d stdout=%q stderr=%q", want.code, want.stdout, want.stderr)
 t.Logf("native: exit=%d stdout=%q stderr=%q", got.code, got.stdout, got.stderr)
 compareNativeAgreement(t, got, want)
}
'''

(evidence / 'native-mutant-probe.go.txt').write_text(probe_text)
try:
 source.write_text(mutated)
 with (evidence / 'mutant-refusal.log').open('w') as log:
  refusal = subprocess.run(['go', 'test', './internal/lower', '-run', 'TestViewStringElementReadRefused', '-count=1', '-v', '-timeout', '90s'], stdout=log, stderr=subprocess.STDOUT, timeout=180)
 probe.write_text(probe_text)
 with (evidence / 'mutant-native.log').open('w') as log:
  native = subprocess.run(['go', 'test', './internal/lower', '-run', 'TestFX6NativeMutantWitness', '-count=1', '-v', '-timeout', '90s'], stdout=log, stderr=subprocess.STDOUT, timeout=180)
 print('refusal mutant exit:', refusal.returncode)
 print('native disagreement witness exit:', native.returncode)
 assert refusal.returncode != 0 and 'got <nil>' in (evidence / 'mutant-refusal.log').read_text()
 assert native.returncode != 0 and 'Native backend stdout' in (evidence / 'mutant-native.log').read_text()
finally:
 source.write_text(original)
 if probe.exists(): probe.unlink()
