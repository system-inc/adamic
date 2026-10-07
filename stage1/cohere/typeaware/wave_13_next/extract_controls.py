"""Export unchanged upstream rule fixtures through test-file-only Go overlays."""
import json, pathlib, subprocess, sys
repo = pathlib.Path(sys.argv[1]).resolve()
out = pathlib.Path(sys.argv[2]).resolve()
out.mkdir(parents=True, exist_ok=True)
rules = repo / 'cohere/internal/lint/rules/nexus'
replace = {}
for name in ('no_process_exit_after_output', 'no_uncleared_race_timeout', 'require_blocking_standard_streams'):
    original = rules / ('correctness_' + name + '_test.go')
    target = out / original.name
    text = original.read_text().replace('rule_testing.RunTypedFilesWithSetup(', 'wave13CaptureSetup(').replace('rule_testing.RunTypedFiles(', 'wave13Capture(')
    target.write_text(text)
    replace[str(original)] = str(target)
helper = out / 'wave13_capture_test.go'
helper.write_text(r'''package nexus
import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "os"
 "path/filepath"
 "sort"
 "strings"
 "testing"
 "github.com/system-inc/cohere/internal/lint/rule"
 rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)
func wave13Save(t *testing.T, files map[string]string, subject string, setup func(string)) {
 sum:=sha256.Sum256([]byte(t.Name())); directory:=filepath.Join(os.Getenv("WAVE13_CAPTURE"),hex.EncodeToString(sum[:8]));if err:=os.MkdirAll(directory,0755);err!=nil{t.Fatal(err)}
 var roots []string
 for name,text:=range files {path:=filepath.Join(directory,strings.TrimPrefix(name,"/"));if err:=os.MkdirAll(filepath.Dir(path),0755);err!=nil{t.Fatal(err)};if err:=os.WriteFile(path,[]byte(rule_testing.FixtureText(text)),0644);err!=nil{t.Fatal(err)};if !strings.HasSuffix(path,".d.ts"){roots=append(roots,path)}}
 sort.Strings(roots);if err:=os.WriteFile(filepath.Join(directory,"roots.manifest"),[]byte(strings.Join(roots,"\n")+"\n"),0644);err!=nil{t.Fatal(err)}
 config:=`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"moduleDetection":"force","types":[]},"include":["**/*.ts","**/*.tsx"]}`
 if setup!=nil {setup(directory);if text,err:=os.ReadFile(filepath.Join(directory,"tsconfig.json"));err==nil{config=string(text)}}
 if err:=os.WriteFile(filepath.Join(directory,"tsconfig.json"),[]byte(config),0644);err!=nil{t.Fatal(err)}
 info,_:=json.Marshal(map[string]string{"test":t.Name(),"subject":subject});if err:=os.WriteFile(filepath.Join(directory,"case.json"),info,0644);err!=nil{t.Fatal(err)}
}
func wave13Capture(t *testing.T,subject rule.Rule,files map[string]string,name string)rule_testing.Result{wave13Save(t,files,name,nil);return rule_testing.RunTypedFiles(t,subject,files,name)}
func wave13CaptureSetup(t *testing.T,subject rule.Rule,files map[string]string,name string,setup func(string))rule_testing.Result{wave13Save(t,files,name,setup);return rule_testing.RunTypedFilesWithSetup(t,subject,files,name,setup)}
''')
replace[str(rules / helper.name)] = str(helper)
overlay = out / 'overlay.json'
overlay.write_text(json.dumps({'Replace': replace}))
import os
env = dict(os.environ, WAVE13_CAPTURE=str(out / 'controls'))
with (out / 'export.log').open('wb') as log:
    result = subprocess.run(['go','test','-overlay',str(overlay),'./internal/lint/rules/nexus','-run','^TestCorrectness(NoProcessExitAfterOutput|NoUnclearedRaceTimeout|RequireBlockingStandardStreams)','-count=1','-v','-timeout=15m'],cwd=repo/'cohere',env=env,stdout=log,stderr=subprocess.STDOUT)
if result.returncode: raise SystemExit(result.returncode)
cases = sorted((out / 'controls').glob('*/case.json'))
print('exported', len(cases), 'upstream cases')
