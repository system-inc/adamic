#include <stdio.h>
#include <unistd.h>
int main(void) { char text[4096]; puts(getcwd(text, sizeof text)); FILE *f=fopen("reading/hello.txt","r"); puts(f ? "opened" : "missing"); if (f) fclose(f); return 0; }
