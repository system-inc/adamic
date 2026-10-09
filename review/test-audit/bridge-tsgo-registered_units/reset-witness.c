#include "tsgo.h"
#include <stdio.h>
int main(void) {
 tsgo_result result = {0};
 result.kind = 7;
 tsgo_result_free(&result);
 printf("%u\n", result.kind);
 return 0;
}
