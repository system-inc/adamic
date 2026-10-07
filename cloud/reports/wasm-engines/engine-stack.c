#include <stdio.h>
static volatile unsigned count;
__attribute__((noinline)) static int recurse(void) {
 count++;
 return recurse();
}
int main(void) { fputs("start\n", stdout); return recurse(); }
