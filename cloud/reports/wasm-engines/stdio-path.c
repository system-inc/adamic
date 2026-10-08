#include <errno.h>
#include <stdio.h>
#include <string.h>
int main(void) {
 FILE *file = fopen("/dev/stdout", "w");
 if (!file) { fprintf(stderr, "open errno=%d: %s\n", errno, strerror(errno)); return 1; }
 fputs("through path\n", file);
 fclose(file);
 return 0;
}
