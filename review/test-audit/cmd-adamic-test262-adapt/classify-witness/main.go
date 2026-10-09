package main
import("encoding/json";"fmt";"strings")
func main(){r:=classify("t.js","/*---\ndescription: x\n---*/\nvar value = 1;\nassert.sameValue(value == 1, true);\n",true);b,_:=json.Marshal(map[string]any{"const_value":strings.Contains(r.Program,"const value"),"let_value":strings.Contains(r.Program,"let value"),"var_to_let_count":r.Adaptations[adaptVarToLet]});fmt.Println(string(b))}
