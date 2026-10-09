package main
import (
 "fmt"
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "github.com/system-inc/adamic/internal/ir"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/native"
)
func main() {
 directory,err:=os.MkdirTemp("", "u061-survivor-"); if err!=nil {panic(err)}; defer os.RemoveAll(directory)
 path:=filepath.Join(directory,"optional.ts")
 if err:=os.WriteFile(path,[]byte("interface P { readonly p?: string; } const p: P = { p: undefined }; console.log('ok');"),0600);err!=nil{panic(err)}
 _,err=load.Load([]string{path}); fmt.Printf("optional explicit undefined: accepted=%v error=%v\n",err==nil,err)
 program:=&ir.Program{Source:"survivor.a",Strings:[]string{"a b"},Main:[]ir.Statement{ir.WriteLine{Value:ir.StringConstant{Index:0}}}}
 code:=native.C(program)
 for _,line:=range strings.Split(code,"\n") {if strings.Contains(line,"ADAMIC_STRING("){fmt.Println(line)}}
 binary:=filepath.Join(directory,"native"); if err:=native.Build(code,binary,native.Options{Sanitize:true});err!=nil{panic(err)}
 output,err:=exec.Command(binary).CombinedOutput();fmt.Printf("native output=%q error=%v\n",output,err)
}
