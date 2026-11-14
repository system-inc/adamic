/* Private stdout width cache, included only by adamic.c. Node initializes the
 * stream on first use and refreshes it only at an event-loop checkpoint.
 * output_lock covers initialization, reads and refresh; the signal handler
 * publishes only its sig_atomic_t bit and wake byte. Neither pipe end closes:
 * a late signal must never write into a reused descriptor. */
#ifndef ADAMIC_TARGET_WASI
#include <sys/ioctl.h>

static bool stdout_tty_started, stdout_width_known;
static size_t stdout_width;
static volatile sig_atomic_t stdout_winch_pending;
static int stdout_winch_pipe[2] = {-1, -1};

static void stdout_winch(int signal_number) {
    (void)signal_number;
    int saved = errno;
    stdout_winch_pending = 1;
    unsigned char byte = 1;
    /* The nonblocking pipe lives until process exit; the handler never takes
     * a lock, allocates, invokes a host hook, or races descriptor reuse. */
    ssize_t result;
    do { result = write(stdout_winch_pipe[1], &byte, 1); }
    while (result == -1 && errno == EINTR);
    errno = saved;
}

static void stdout_tty_start(void) {
    if (stdout_tty_started) return;
    stdout_tty_started = true;
    if (!isatty(adamic_stdout)) return;
    struct winsize size;
    if (ioctl(adamic_stdout, TIOCGWINSZ, &size) == 0) {
        stdout_width = size.ws_col;
        stdout_width_known = true;
    }
    if (pipe(stdout_winch_pipe) != 0) abort();
    for (unsigned index = 0; index < 2; index++) {
        int flags = fcntl(stdout_winch_pipe[index], F_GETFL);
        if (flags == -1 || fcntl(stdout_winch_pipe[index], F_SETFL, flags | O_NONBLOCK) == -1 ||
            fcntl(stdout_winch_pipe[index], F_SETFD, FD_CLOEXEC) == -1) abort();
    }
    struct sigaction action;
    memset(&action, 0, sizeof action);
    action.sa_handler = stdout_winch;
    sigemptyset(&action.sa_mask);
    action.sa_flags = SA_RESTART;
    if (sigaction(SIGWINCH, &action, NULL) != 0) abort();
}

void adamic_host_stdout_start(void) {
    pthread_mutex_lock(&output_lock);
    stdout_tty_start();
    pthread_mutex_unlock(&output_lock);
}

adamic_maybe_boolean adamic_host_stdout_is_tty(void) {
    pthread_mutex_lock(&output_lock);
    stdout_tty_start();
    adamic_maybe_boolean result = {isatty(adamic_stdout) != 0, true};
    pthread_mutex_unlock(&output_lock);
    return result;
}

adamic_maybe_number adamic_host_stdout_columns(void) {
    pthread_mutex_lock(&output_lock);
    stdout_tty_start();
    adamic_maybe_number result = {stdout_width_known, (double)stdout_width};
    pthread_mutex_unlock(&output_lock);
    return result;
}

int adamic_host_stdout_wake_descriptor(void) {
    pthread_mutex_lock(&output_lock);
    int descriptor = stdout_winch_pipe[0];
    pthread_mutex_unlock(&output_lock);
    return descriptor;
}

bool adamic_host_stdout_refresh_pending(void) { return stdout_winch_pending != 0; }

void adamic_host_stdout_refresh(void) {
    pthread_mutex_lock(&output_lock);
    if (!stdout_winch_pending) {
        pthread_mutex_unlock(&output_lock);
        return;
    }
    stdout_winch_pending = 0;
    unsigned char bytes[256];
    ssize_t result;
    do { result = read(stdout_winch_pipe[0], bytes, sizeof bytes); }
    while (result > 0 || (result == -1 && errno == EINTR));
    struct winsize size;
    if (ioctl(adamic_stdout, TIOCGWINSZ, &size) == 0) {
        stdout_width = size.ws_col;
        stdout_width_known = true;
        pthread_mutex_unlock(&output_lock);
        return;
    }
    /* The oracle's stdout error listener ends through the same broken-output
     * convention as a closed pipe: no stderr, status 70. Never a source throw. */
    broken[adamic_stdout] = true;
    pthread_mutex_unlock(&output_lock);
    finish();
    adamic_invariant("stdout refresh failure returned from finish"); /* finish cannot return with broken stdout. */
}
#else
void adamic_host_stdout_start(void) {}
adamic_maybe_boolean adamic_host_stdout_is_tty(void) {
    return (adamic_maybe_boolean){isatty(adamic_stdout) != 0, true};
}
int adamic_host_stdout_wake_descriptor(void) { return -1; }
bool adamic_host_stdout_refresh_pending(void) { return false; }
void adamic_host_stdout_refresh(void) {}
#endif
