package main
import("os";"github.com/system-inc/adamic/internal/native")
func main(){source:="#include \"adamic.h\"\n#include <stdio.h>\nint main(void){printf(\"%llu\\n\",(unsigned long long)adamic_map_number_hash(1));return 0;}\n";if err:=native.Build(source,os.Args[1],native.Options{});err!=nil{panic(err)}}
