package main
import (
 "context"
 "fmt"
 "os"
 "time"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
 "github.com/system-inc/adamic/internal/native"
)
func main() {
 start:=time.Now()
 program,err:=load.Load([]string{"stage1/typescript/parser/main.ts"});if err!=nil {panic(err)}
 lowered,err:=lower.Lower(context.Background(),program);if err!=nil {panic(err)}
 if err:=native.Build(native.C(lowered),os.Args[1],native.Options{Sanitize:true});err!=nil {panic(err)}
 fmt.Printf("sanitized native build seconds %.3f\n",time.Since(start).Seconds())
}
