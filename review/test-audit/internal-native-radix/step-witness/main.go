package main
import("fmt";"os";"os/exec";"github.com/system-inc/adamic/internal/native")
func main(){
 const source=`#include "adamic.h"
#include <stdio.h>
static const adamic_regex_instruction code[] = {{.op=3,.x=1},{.op=0}};
static const adamic_regex_program program = {.code=code};
int main(void) {
 static adamic_string empty=ADAMIC_STRING("");
 adamic_object *regex=adamic_regex_new(&program,&empty,&empty);
 adamic_regex_set_step_limit(1);
 printf("matched %d\n",adamic_regex_test(regex,&empty));
 adamic_release(regex);
 return 0;
}
`
 binary:="/tmp/u052-step-witness"
 if err:=native.Build(source,binary,native.Options{Sanitize:true});err!=nil{panic(err)}
 for _,id:=range []string{"","M12"}{cmd:=exec.Command(binary);cmd.Env=append(os.Environ(),"ADAMIC_MUTANT="+id);out,err:=cmd.CombinedOutput();fmt.Printf("selector=%q error=%v output=%q\n",id,err,out)}
}
