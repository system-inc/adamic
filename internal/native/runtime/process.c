// Synchronous child execution for stage 1's command runner. One unlinked temporary descriptor
// captures both streams, so large output cannot fill a pipe while the parent waits.
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <unistd.h>
extern char **environ;

static void *process_alloc(size_t size) {
    void *result = malloc(size);
    if (!result) adamic_panic("out of memory", 13);
    return result;
}
static const char *process_errno(int error) {
    switch (error) {
    case ENOENT: return "ENOENT";
    case ENOTDIR: return "ENOTDIR";
    case EACCES: return "EACCES";
    case EPERM: return "EPERM";
    case ENOEXEC: return "ENOEXEC";
    case E2BIG: return "E2BIG";
    case EINVAL: return "EINVAL";
    case EMFILE: return "EMFILE";
    case ENFILE: return "ENFILE";
    default: return "EIO";
    }
}
static adamic_object *process_result(const unsigned char *bytes, size_t length, int code, int signal, const char *stage, int error) {
    static const char *const names[] = {"output", "exitCode", "signal", "error"};
    static const bool references[] = {true, false, false, true};
    static const adamic_shape shape = {4, names, references, NULL};
    adamic_object *result = adamic_object_new(&shape);
    result->slots[0].reference = adamic_decode_utf8(bytes, length);
    result->slots[1].number = code;
    result->slots[2].number = signal;
    char message[80] = "";
    if (stage) snprintf(message, sizeof message, "%s:%s", stage, process_errno(error));
    result->slots[3].reference = adamic_decode_utf8((const unsigned char *)message, strlen(message));
    return result;
}
static void free_vector(char **vector, size_t count) {
    for (size_t i = 0; i < count; i++) free(vector[i]);
    free(vector);
}
// Both vectors are owned by the parent before fork; the child allocates nothing before exec.
adamic_object *adamic_run_process(const adamic_string *executable, const adamic_array *arguments, const adamic_string *directory, const adamic_array *patches) {
    adamic_output_flush();
    char *file = adamic_path_bytes(executable), *cwd = adamic_path_bytes(directory);
    size_t argc = arguments->length + 1, inherited = 0;
    while (environ[inherited]) inherited++;
    char **argv = process_alloc((argc + 1) * sizeof *argv);
    char **env = process_alloc((inherited + patches->length + 1) * sizeof *env);
    size_t envc = 0;
    argv[0] = file;
    bool valid = file && cwd && file[0] == '/';
    for (size_t i = 1; i < argc; i++) {
        argv[i] = adamic_path_bytes(arguments->elements[i - 1].reference);
        if (!argv[i]) valid = false;
    }
    argv[argc] = NULL;
    for (size_t i = 0; i < inherited; i++) {
        size_t n = strlen(environ[i]) + 1;
        env[envc] = process_alloc(n);
        memcpy(env[envc++], environ[i], n);
    }
    for (size_t i = 0; i < patches->length; i++) {
        char *entry = adamic_path_bytes(patches->elements[i].reference);
        if (!entry) { valid = false; continue; }
        char *equals = strchr(entry, '=');
        size_t key = equals ? (size_t)(equals - entry) : strlen(entry);
        for (size_t j = 0; j < envc;) {
            if (strncmp(env[j], entry, key) == 0 && env[j][key] == '=') {
                free(env[j]); memmove(env + j, env + j + 1, (--envc - j) * sizeof *env);
            } else j++;
        }
        if (equals) env[envc++] = entry;
        else free(entry);
    }
    env[envc] = NULL;
    const char *stage = "exec";
    int error = EINVAL, status = 0, code = 1, signal = 0;
    FILE *capture = NULL;
    int handshake[2] = {-1, -1};
    unsigned char *output = NULL;
    size_t used = 0, capacity = 4096;
    if (!valid) goto done;
    capture = tmpfile();
    stage = "capture";
    if (!capture) { error = errno; goto done; }
    if (pipe(handshake) != 0) { error = errno; goto done; }
    if (fcntl(handshake[1], F_SETFD, FD_CLOEXEC) < 0) { error = errno; goto done; }
    pid_t child = fork();
    stage = "exec";
    if (child < 0) { error = errno; goto done; }
    if (child == 0) {
        close(handshake[0]);
        int failure[2] = {0, 0};
        int input = open("/dev/null", O_RDONLY);
        if (input < 0 || dup2(input, STDIN_FILENO) < 0 || dup2(fileno(capture), STDOUT_FILENO) < 0 || dup2(fileno(capture), STDERR_FILENO) < 0) {
            failure[1] = errno;
        } else {
            if (input > STDERR_FILENO) close(input);
            if (fileno(capture) > STDERR_FILENO) close(fileno(capture));
            if (cwd[0] && chdir(cwd) != 0) { failure[0] = 1; failure[1] = errno; }
            else { execve(file, argv, env); failure[1] = errno; }
        }
        // One atomic pipe write communicates launch errors, not a guessed child exit status.
        ssize_t written;
        do { written = write(handshake[1], failure, sizeof failure); } while (written < 0 && errno == EINTR);
        _exit(127);
    }
    close(handshake[1]); handshake[1] = -1;
    int failure[2] = {0, 0};
    size_t got = 0;
    while (got < sizeof failure) {
        ssize_t n = read(handshake[0], (char *)failure + got, sizeof failure - got);
        if (n < 0 && errno == EINTR) continue;
        if (n <= 0) break;
        got += (size_t)n;
    }
    pid_t waited;
    do { waited = waitpid(child, &status, 0); } while (waited < 0 && errno == EINTR);
    if (waited < 0) { error = errno; stage = "wait"; goto done; }
    stage = NULL; error = 0;
    if (got) { stage = failure[0] ? "chdir" : "exec"; error = failure[1]; }
    else if (WIFEXITED(status)) code = WEXITSTATUS(status);
    else if (WIFSIGNALED(status)) signal = WTERMSIG(status);
    if (fseek(capture, 0, SEEK_SET) != 0) { stage = "capture"; error = errno; goto done; }
    output = process_alloc(capacity);
    for (;;) {
        if (used == capacity) {
            capacity *= 2;
            unsigned char *grown = realloc(output, capacity);
            if (!grown) { free(output); adamic_panic("out of memory", 13); }
            output = grown;
        }
        size_t n = fread(output + used, 1, capacity - used, capture);
        used += n;
        if (!n) {
            if (ferror(capture)) { stage = "capture"; error = errno; }
            break;
        }
    }
 done:
    if (handshake[0] >= 0) close(handshake[0]);
    if (handshake[1] >= 0) close(handshake[1]);
    if (capture) fclose(capture);
    free_vector(argv, argc); free_vector(env, envc); free(cwd);
    adamic_object *result = process_result(output, used, code, signal, stage, error);
    free(output);
    return result;
}
