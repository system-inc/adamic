#define _DEFAULT_SOURCE
#define _DARWIN_C_SOURCE
#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/resource.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <time.h>
#include <unistd.h>
#include <fcntl.h>

static volatile sig_atomic_t child_pid = 0;
static volatile sig_atomic_t timed_out = 0;

static void timeout_child(int signal_number) {
    (void)signal_number;
    timed_out = 1;
    if (child_pid > 0) {
        (void)kill(-child_pid, SIGKILL);
        (void)kill(child_pid, SIGKILL);
    }
}

static double milliseconds(const struct timeval *value) {
    return (double)value->tv_sec * 1000.0 + (double)value->tv_usec / 1000.0;
}

int main(int argc, char **argv) {
    if (argc < 6) {
        fprintf(stderr, "usage: wait4 result.json stdout stderr timeout_ms command [args...]\n");
        return 2;
    }
    char *end = NULL;
    errno = 0;
    const long timeout_ms = strtol(argv[4], &end, 10);
    if (errno != 0 || end == argv[4] || *end != '\0' || timeout_ms < 1) return 2;
    const int output = open(argv[2], O_WRONLY | O_CREAT | O_TRUNC, 0600);
    const int error = open(argv[3], O_WRONLY | O_CREAT | O_TRUNC, 0600);
    if (output < 0 || error < 0) { perror("open child streams"); return 1; }
    struct sigaction action = {0};
    action.sa_handler = timeout_child;
    if (sigemptyset(&action.sa_mask) != 0 || sigaction(SIGALRM, &action, NULL) != 0) return 1;
    struct timespec began, ended;
    if (clock_gettime(CLOCK_MONOTONIC, &began) != 0) return 1;
    const pid_t child = fork();
    if (child < 0) { perror("fork"); return 1; }
    if (child == 0) {
        if (setsid() < 0 || dup2(output, STDOUT_FILENO) < 0 || dup2(error, STDERR_FILENO) < 0) _exit(126);
        (void)close(output);
        (void)close(error);
        execvp(argv[5], argv + 5);
        perror("execvp");
        _exit(127);
    }
    child_pid = child;
    (void)close(output);
    (void)close(error);
    struct itimerval timer = {0};
    timer.it_value.tv_sec = timeout_ms / 1000;
    timer.it_value.tv_usec = (timeout_ms % 1000) * 1000;
    if (setitimer(ITIMER_REAL, &timer, NULL) != 0) { timeout_child(SIGALRM); }
    int status = 0;
    struct rusage usage;
    pid_t waited;
    do { waited = wait4(child, &status, 0, &usage); } while (waited < 0 && errno == EINTR);
    timer.it_value.tv_sec = 0;
    timer.it_value.tv_usec = 0;
    (void)setitimer(ITIMER_REAL, &timer, NULL);
    if (waited != child || clock_gettime(CLOCK_MONOTONIC, &ended) != 0) { perror("wait4/clock"); return 1; }
    const double wall = ((double)ended.tv_sec - (double)began.tv_sec) * 1000.0
        + ((double)ended.tv_nsec - (double)began.tv_nsec) / 1000000.0;
    const double user = milliseconds(&usage.ru_utime);
    const double system = milliseconds(&usage.ru_stime);
#ifdef __APPLE__
    const double rss = (double)usage.ru_maxrss;
#else
    const double rss = (double)usage.ru_maxrss * 1024.0;
#endif
    const int exit_code = WIFEXITED(status) ? WEXITSTATUS(status) : -WTERMSIG(status);
    FILE *result = fopen(argv[1], "w");
    if (result == NULL) { perror("open result"); return 1; }
    fprintf(result, "{\"pid\":%ld,\"wallMs\":%.9f,\"userMs\":%.9f,\"systemMs\":%.9f,\"cpuMs\":%.9f,"
        "\"maxRssBytes\":%.0f,\"minorFaults\":%ld,\"majorFaults\":%ld,\"voluntaryContextSwitches\":%ld,"
        "\"involuntaryContextSwitches\":%ld,\"exitCode\":%d,\"timedOut\":%s}\n",
        (long)child, wall, user, system, user + system, rss, usage.ru_minflt, usage.ru_majflt,
        usage.ru_nvcsw, usage.ru_nivcsw, exit_code, timed_out ? "true" : "false");
    return fclose(result) == 0 ? 0 : 1;
}
