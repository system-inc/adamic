#define _GNU_SOURCE
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/resource.h>
#include <sys/wait.h>
#include <unistd.h>

// Run from a fresh small address space so Python's pre-exec RSS cannot inflate the workload's peak.
int main(int argc, char **argv) {
    if (argc < 3) return 2;
    pid_t child = fork();
    if (child < 0) { perror("fork"); return 2; }
    if (child == 0) { execvp(argv[2], argv + 2); perror("execvp"); _exit(127); }
    int status;
    struct rusage usage;
    while (wait4(child, &status, 0, &usage) < 0) { if (errno != EINTR) { perror("wait4"); return 2; } }
    FILE *out = fopen(argv[1], "w");
    if (out == NULL) { perror("usage file"); return 2; }
    fprintf(out, "%ld %.6f %.6f\n", usage.ru_maxrss,
            (double)usage.ru_utime.tv_sec + (double)usage.ru_utime.tv_usec / 1000000.0,
            (double)usage.ru_stime.tv_sec + (double)usage.ru_stime.tv_usec / 1000000.0);
    if (fclose(out) != 0) return 2;
    return WIFEXITED(status) ? WEXITSTATUS(status) : 128 + WTERMSIG(status);
}
