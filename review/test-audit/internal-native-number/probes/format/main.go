package main

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
)

func main() {
	err := native.Build(`#include "adamic.h"
#include <stdio.h>
int main(void){double values[]={4503599627370496.0,4503599627370497.0,9007199254740991.0,9007199254740992.0};for(int i=0;i<4;i++){char b[ADAMIC_NUMBER_FORMAT_MAX];size_t n=adamic_number_format(values[i],b);fwrite(b,1,n,stdout);putchar('\n');}return 0;}
`, os.Args[1], native.Options{Sanitize: true})
	if err != nil {
		panic(err)
	}
}
