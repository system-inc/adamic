package main

import (
 "context"
 "fmt"
 "os"
 "os/exec"
 "github.com/system-inc/adamic/internal/javascript"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
)

func main() {
 mode := os.Args[1]
 source := "class C {constructor(public value:number){}} const c=new C(7); console.log(`${c.value}`);"
 if mode == "index" { source = "const xs:number[]=[]; const x:number=xs[0]; console.log(`${x}`);" }
 path := "/tmp/u041-witness-"+mode+".a"
 if err := os.WriteFile(path, []byte(source), 0644); err != nil { panic(err) }
 checked, err := load.Load([]string{path})
 if mode == "index" { fmt.Printf("load error: %v\n", err); return }
 if err != nil { panic(err) }
 program, err := lower.Lower(context.Background(), checked); if err != nil { panic(err) }
 generated := "/tmp/u041-witness-"+mode+".mjs"
 if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0644); err != nil { panic(err) }
 command := exec.Command("node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", generated)
 output, err := command.CombinedOutput()
 fmt.Print(string(output)); if err != nil { panic(err) }
}
