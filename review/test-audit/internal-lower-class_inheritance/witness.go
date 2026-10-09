// Independent survivor observation, not a test or a mutation of the oracle.
package main
import("context";"encoding/json";"fmt";"os";"path/filepath";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/lower")
func main(){
 sources:=[]string{
 `class Base {} class Source extends Base { readonly name = "source"; read(): string { return this.name; } } const pair: readonly [boolean, boolean, Source] = [false, true, new Source()]; const [, ignored, source] = pair; console.log(ignored ? source.read() : "none");`,
 `class A { readonly x = 1; } class B extends A { readonly y = 2; } const b = new B(); console.log(b.x.toString());`,
 `class A { f(): void {} } class B extends A { override f(): boolean { return true; } } const a: A = new B();`,
 `class A {} class B extends A { readonly value = 42; constructor() { super(); } } const b = new B(); console.log(b.value.toString());`,
 }
 for i,s:=range sources {dir:=filepath.Join("/tmp/u029",fmt.Sprintf("witness-source-%d",i));e:=os.MkdirAll(dir,0700);if e!=nil{panic(e)};path:=filepath.Join(dir,"main.ts");os.WriteFile(path,[]byte(s),0600);loaded,e:=load.Load([]string{path});if e!=nil{fmt.Printf("W%d load: %v\n",i,e);continue};p,e:=lower.Lower(context.Background(),loaded);if e!=nil{fmt.Printf("W%d %T: %v\n",i,e,e);continue};b,e:=json.Marshal(p);if e!=nil{panic(e)};fmt.Printf("W%d %s\n",i,b)}
}
