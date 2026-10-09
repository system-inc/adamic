package main
import("encoding/json";"fmt")
func main(){ cases:=[]string{"function cb(x){return x;} [1].forEach(cb);", "var x=1e+2;", "var x=1; class Box { value: number; }"}; out:=map[string]adapted{};for _,s:=range cases {out[s]=adaptSource(s)};b,_:=json.MarshalIndent(out,"","  ");fmt.Println(string(b))}
