// Owned overlay around upstream tests; production calls and assertions are unchanged.
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
	"os"
	"path/filepath"
	"testing"
)

func wave08Save(t *testing.T, subject rule.Rule, filename, source string, options any, typed bool) {
 if typed {source=rule_testing.FixtureText(source)}
 payload,err:=json.Marshal(struct {Name,Rule,Subject,Config string; Files map[string]string; Options any; Typed bool}{t.Name(),subject.Name,filename,wave08DefaultConfig,map[string]string{filename:source},options,typed})
 if err!=nil {t.Fatal(err)}
 hash:=sha256.Sum256(payload);name:=hex.EncodeToString(hash[:])
 directory:=os.Getenv("ADAMIC_WAVE08_CAPTURE")
 if directory!="" {if err:=os.MkdirAll(directory,0755);err!=nil {t.Fatal(err)};if err:=os.WriteFile(filepath.Join(directory,name+".json"),payload,0644);err!=nil {t.Fatal(err)}}
}
func wave08Capture(t *testing.T, subject rule.Rule, filename, source string) rule_testing.Result {
 wave08Save(t,subject,filename,source,nil,true);return rule_testing.RunTyped(t,subject,filename,source)
}
func wave08CaptureOptions(t *testing.T, subject rule.Rule, filename, source string, options any) rule_testing.Result {
 wave08Save(t,subject,filename,source,options,true);return rule_testing.RunTypedWithOptions(t,subject,filename,source,options)
}
func wave08CapturePlain(t *testing.T, subject rule.Rule, filename, source string) rule_testing.Result {
 wave08Save(t,subject,filename,source,nil,false);return rule_testing.Run(t,subject,filename,source)
}

const wave08DefaultConfig = `{
	"compilerOptions": {
		"strict": true,
		"target": "ES2022",
		"lib": ["ES2022"],
		"moduleDetection": "force",
		"types": []
	},
	"include": ["**/*.ts", "**/*.tsx"]
}`
