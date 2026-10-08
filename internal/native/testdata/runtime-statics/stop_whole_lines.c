#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <signal.h>
#include <unistd.h>
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 adamic_string first = ADAMIC_STRING("first whole line");
 adamic_string second = ADAMIC_STRING("second whole line");
 adamic_write_line(adamic_stdout, &first);
 adamic_write_line(adamic_stdout, &second);
 raise(SIGUSR2);
 for (;;) { pause(); }
}
