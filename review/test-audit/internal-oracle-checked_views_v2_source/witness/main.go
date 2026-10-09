package main
import (
 "context"
 "encoding/json"
 "fmt"
 "os"
 "github.com/system-inc/adamic/internal/ir"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
)
func main() {
 if os.Args[1]=="tags" {
  contracts:=[]ir.ViewContract{{Kind:ir.ViewUnion,Of:ir.Object,Name:"Root",Members:[]ir.ViewContractID{2}},{Kind:ir.ViewObject,Of:ir.Object,Name:"Member",Fields:[]ir.ViewFieldContract{{Name:"tag",Contract:3,Optional:true}}},{Kind:ir.ViewScalar,Of:ir.String,Allowed:[]ir.ViewLiteral{{Of:ir.String,String:"x"}}}}
  members,err:=lower.UntaggedViewMembers(contracts,1);value,_:=json.Marshal(members);fmt.Printf("%s error=%v\n",value,err);return
 }
 loaded,err:=load.Load([]string{os.Args[1]});if err!=nil {fmt.Println(err);return};p,err:=lower.Lower(context.Background(),loaded);if err!=nil {fmt.Println(err);return};data,_:=json.Marshal(p.ViewContracts);fmt.Println(string(data))
}
