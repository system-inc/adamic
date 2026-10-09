// A cgo package whose header sits in another directory, reached by '#cgo CFLAGS: -I'.
package main

// #cgo CFLAGS: -I${SRCDIR}/../include
// #include "greeting.h"
// static const char *greeting(void) { return GREETING; }
import "C"

import "fmt"

func main() { fmt.Println(C.GoString(C.greeting())) }
