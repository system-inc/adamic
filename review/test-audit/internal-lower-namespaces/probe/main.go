package main
import("context";"encoding/json";"fmt";"os";"path/filepath";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/lower")
func main(){
 cases:=[]struct{Name,Source string}{
 {"type-only-namespace","namespace N {export interface Box {readonly x:number;}} const value:N.Box={x:7}; console.log(`${value.x}`);"},
 {"ready-bits","namespace N {export let x=1;} console.log(`${N.x}`);"},
 {"callback-before-init","function read():boolean{return N.x;} [1].map(read); namespace N {export let x=false;}"},
 {"helper-before-init","function read():boolean{return N.x;} function helper():boolean{return read();} helper(); namespace N {export let x=false;}"},
 {"object-escape","namespace N {export let x=1;} const alias=N;"},
 {"returned-reference","namespace N {let text=''; export function set():string{return text='built'.repeat(2);}}"},
 {"receiver-this","namespace N {export function read(this:{readonly x:number}):number{return this.x;}}"},
 {"mutable-map-view","const empty = new Map<never, never>(); const wide: Map<string, number> = empty; wide.set('x', 1);"},
 {"class-merge","class Logger {} namespace Logger {export const level=1;} console.log(`${Logger.level}`);"},
 }
 directory:="/tmp/u038/probe-input";if err:=os.MkdirAll(directory,0755);err!=nil{panic(err)}
 for _,c:=range cases{
  p:=filepath.Join(directory,"main.a");if err:=os.WriteFile(p,[]byte(c.Source),0644);err!=nil{panic(err)}
  loaded,err:=load.Load([]string{p});if err!=nil{panic(err)}
  product,err:=lower.Lower(context.Background(),loaded)
  answer:=map[string]any{"case":c.Name}
  if err!=nil{answer["error"]=err.Error();answer["error_kind"]=fmt.Sprintf("%T",err)} else if product==nil{answer["product"]=nil} else {answer["product"]=map[string]any{"locals":product.Locals,"main":product.Main,"functions":product.Functions,"strings":product.Strings}}
  b,err:=json.Marshal(answer);if err!=nil{panic(err)};fmt.Println(string(b))
 }
}
