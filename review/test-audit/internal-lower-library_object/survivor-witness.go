package main

import (
 "context"
 "fmt"
 "os"
 "path/filepath"

 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
)

func main() {
 cases := []struct{name,source string}{
  {"const freeze", "const item = {value: 1}; Object.freeze(item);"},
  {"grouping diagnostic", "Object.groupBy([1,2], (value: number): string => value === 1 ? 'one' : 'other');"},
  {"immutable regex offset", "const pattern = /a/; console.log('a'.replace(pattern, (match: string, offset: number) => `${match}:${offset}`));"},
  {"nested tuple array", "function rows(): readonly (readonly number[])[] { return [[1, 2] as readonly [number, number]]; }"},
  {"typed tuple field", "const held: { readonly points: readonly [number, number] } = { points: [1,2] }; const view: {readonly points: readonly number[]} = held;"},
 }
 dir,err:=os.MkdirTemp("","u036-witness-");if err!=nil{panic(err)}
 for _,c:=range cases {
  path:=filepath.Join(dir,"main.a");if err:=os.WriteFile(path,[]byte(c.source),0600);err!=nil{panic(err)}
  source,err:=load.Load([]string{path});if err!=nil{fmt.Printf("%s: Load %v\n",c.name,err);continue}
  program,err:=lower.Lower(context.Background(),source)
  if err!=nil {fmt.Printf("%s: %T %v\n",c.name,err,err)} else {fmt.Printf("%s: success, main=%d functions=%d\n",c.name,len(program.Main),len(program.Functions))}
 }
}
