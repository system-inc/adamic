//go:build lintoracle

package main
import (
 "encoding/json"
 "github.com/system-inc/cohere/internal/lint/rule"
 rules "github.com/system-inc/cohere/internal/lint/rules/boundaries"
)
func oracleDependencies() rule.Rule {return rules.Dependencies}
func oracleDependenciesOptions(fields []string) any {
 if len(fields)<=5 || fields[5]=="" || fields[5]=="null" {return nil}
 var options rules.DependenciesOptions
 if err:=json.Unmarshal([]byte(fields[5]),&options);err!=nil {panic(err)}
 return options
}
