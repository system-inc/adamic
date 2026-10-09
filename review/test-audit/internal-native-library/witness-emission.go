package main
import("context";"fmt";"os";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/lower";"github.com/system-inc/adamic/internal/native")
func main(){p,e:=load.Load([]string{os.Args[1]});if e!=nil{panic(e)};out,e:=lower.Lower(context.Background(),p);if e!=nil{panic(e)};text:=native.C(out);fmt.Print(text);for _,l:=range out.Locals {fmt.Fprintf(os.Stderr,"local %s borrowed=%t\n",l.Name,l.Borrowed)}}
