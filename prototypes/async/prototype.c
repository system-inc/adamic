#define _POSIX_C_SOURCE 200809L
#include <assert.h>
#include <errno.h>
#include <poll.h>
#include <pthread.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

/* Standalone feasibility model, deliberately not the Adamic heap ABI. */
typedef struct Object Object;
struct Object { size_t references; void (*destroy)(Object *); };
static size_t allocations, frees, retains, releases;
static void *allocate(size_t bytes, void (*destroy)(Object *)) {
    Object *object = calloc(1, bytes);
    if (!object) abort();
    object->references = 1;
    object->destroy = destroy;
    allocations++;
    return object;
}
static void retain(Object *object) { if (object) { object->references++; retains++; } }
static void release(Object *object) {
    if (object) {
        releases++;
        if (--object->references == 0) { object->destroy(object); frees++; free(object); }
    }
}
typedef struct { Object object; char bytes[128]; } Text;
static void destroy_text(Object *object) { (void)object; }
static Text *text_new(const char *bytes) {
    Text *text = allocate(sizeof(*text), destroy_text);
    snprintf(text->bytes, sizeof(text->bytes), "%s", bytes);
    return text;
}

typedef void (*Callback)(Object *, Text *, bool);
typedef struct Job Job;
struct Job { Callback callback; Object *context; Text *value; bool rejected; Job *next; };
static Job *micro_head, *micro_tail;
static Job *job_new(Callback callback, Object *context, Text *value, bool rejected) {
    Job *job = calloc(1, sizeof(*job));
    if (!job) abort();
    *job = (Job){callback, context, value, rejected, NULL};
    retain(context); retain((Object *)value);
    return job;
}
static void enqueue(Job *job) {
    if (micro_tail) micro_tail->next = job; else micro_head = job;
    micro_tail = job;
}
static void run_job(Job *job) {
    job->callback(job->context, job->value, job->rejected);
    release((Object *)job->value); release(job->context); free(job);
}
static void drain_microtasks(void) {
    while (micro_head) {
        Job *job = micro_head;
        micro_head = job->next;
        if (!micro_head) micro_tail = NULL;
        run_job(job);
    }
}

typedef struct { Object object; bool settled, rejected; Text *value; Job *first, *last; } Promise;
static void destroy_promise(Object *object) {
    Promise *promise = (Promise *)object;
    assert(!promise->first); /* Pending reactions must be drained by the root. */
    release((Object *)promise->value);
}
static Promise *promise_new(void) { return allocate(sizeof(Promise), destroy_promise); }
static void settle(Promise *promise, Text *value, bool rejected) {
    if (promise->settled) return;
    promise->settled = true; promise->rejected = rejected; promise->value = value;
    retain((Object *)value);
    Job *job = promise->first;
    promise->first = promise->last = NULL;
    while (job) {
        Job *next = job->next; job->next = NULL;
        job->value = value; job->rejected = rejected; retain((Object *)value);
        enqueue(job); job = next;
    }
}
static void observe(Promise *promise, Callback callback, Object *context) {
    Job *job = job_new(callback, context, promise->value, promise->rejected);
    if (promise->settled) {
#ifdef MUTANT_INLINE
        run_job(job); /* Wrong: even fulfilled await must enqueue. */
#else
        enqueue(job);
#endif
    } else {
        if (promise->last) promise->last->next = job; else promise->first = job;
        promise->last = job;
    }
}

/* One timer and one file request suffice for this fixture. No fake file bytes. */
static Job *timer_job;
static struct timespec timer_deadline;
static int completion_pipe[2];
static pthread_t file_thread;
static bool file_active;
typedef struct { Promise *promise; const char *path; char bytes[128]; bool failed; } FileRequest;
static FileRequest file_request;
static void *read_file(void *unused) {
    (void)unused;
    FILE *file = fopen(file_request.path, "rb");
    if (!file) file_request.failed = true;
    else {
        size_t count = fread(file_request.bytes, 1, sizeof(file_request.bytes) - 1, file);
        file_request.failed = ferror(file) != 0;
        file_request.bytes[count] = '\0';
        if (fclose(file) != 0) file_request.failed = true;
        while (count && (file_request.bytes[count-1] == '\n' || file_request.bytes[count-1] == '\r'))
            file_request.bytes[--count] = '\0';
    }
    char ready = 1;
    if (write(completion_pipe[1], &ready, 1) != 1) abort();
    return NULL;
}
static Promise *file_start(const char *path) {
    assert(!file_active);
    Promise *promise = promise_new();
    file_request = (FileRequest){.promise = promise, .path = path};
    retain((Object *)promise); /* Request owns promise; worker never touches its count. */
    file_active = true;
    if (pthread_create(&file_thread, NULL, read_file, NULL) != 0) abort();
    return promise;
}
static void file_complete(void) {
    char ready;
    if (read(completion_pipe[0], &ready, 1) != 1) abort();
    if (pthread_join(file_thread, NULL) != 0) abort();
    file_active = false; /* Join publishes bytes, no shared RC accesses. */
    Text *value = text_new(file_request.failed ? "read failed" : file_request.bytes);
    settle(file_request.promise, value, file_request.failed);
    release((Object *)value); release((Object *)file_request.promise);
    file_request.promise = NULL;
}
static const char *mode, *path;
static bool cancelled;
static void timer_microtask(Object *context, Text *value, bool rejected) {
    (void)context; (void)value; (void)rejected; puts("timer microtask");
}
static void timer_fire(Object *context, Text *value, bool rejected) {
    (void)value; (void)rejected;
    puts("timer");
    if (strcmp(mode, "cancel") == 0) cancelled = true;
    settle((Promise *)context, NULL, false);
#ifdef MUTANT_CHECKPOINT
    /* Model dispatching a later host callback before the Promise checkpoint. */
    run_job(job_new(timer_microtask, NULL, NULL, false));
#else
    enqueue(job_new(timer_microtask, NULL, NULL, false));
#endif
}
static Promise *timer_start(void) {
    Promise *promise = promise_new();
    clock_gettime(CLOCK_MONOTONIC, &timer_deadline);
    timer_deadline.tv_nsec += 1000000;
    if (timer_deadline.tv_nsec >= 1000000000) { timer_deadline.tv_nsec -= 1000000000; timer_deadline.tv_sec++; }
    timer_job = job_new(timer_fire, (Object *)promise, NULL, false);
    return promise;
}
static void event_loop(void) {
    drain_microtasks();
    while (timer_job || file_active) {
        if (timer_job) {
            int result;
            do { result = clock_nanosleep(CLOCK_MONOTONIC, TIMER_ABSTIME, &timer_deadline, NULL); } while (result == EINTR);
            if (result) abort();
            Job *job = timer_job; timer_job = NULL; run_job(job);
        } else {
            struct pollfd descriptor = {completion_pipe[0], POLLIN, 0};
            int result;
            do { result = poll(&descriptor, 1, -1); } while (result < 0 && errno == EINTR);
            if (result != 1) abort();
            file_complete();
        }
        drain_microtasks();
    }
    drain_microtasks();
}

typedef struct { Object object; unsigned state; Promise *output; Text *held, *file; } Frame;
static void destroy_frame(Object *object) {
    Frame *frame = (Frame *)object;
#ifndef MUTANT_LEAK
    release((Object *)frame->held);
#endif
    release((Object *)frame->file); release((Object *)frame->output);
}
static void finish(Frame *frame, Text *value, bool rejected) {
    frame->state = 4; settle(frame->output, value, rejected);
}
static void fail(Frame *frame, const char *message) {
    Text *error = text_new(message); finish(frame, error, true); release((Object *)error);
}
static void resume(Object *context, Text *value, bool rejected) {
    Frame *frame = (Frame *)context;
    if (rejected) { finish(frame, value, true); return; }
    Promise *awaited;
    switch (frame->state) {
        case 0:
            puts("start");
            frame->state = 1;
            awaited = promise_new(); settle(awaited, NULL, false);
            break;
        case 1:
            printf("after sync %s\n", frame->held->bytes);
            frame->state = 2; awaited = timer_start(); break;
        case 2:
            if (cancelled) { fail(frame, "cancelled"); return; }
            printf("after timer %s\n", frame->held->bytes);
            frame->state = 3; awaited = file_start(path); break;
        case 3:
            frame->file = value;
#ifndef MUTANT_BORROW
            retain((Object *)value);
#endif
#ifdef MUTANT_COUNTS
            retain((Object *)value); release((Object *)value);
#endif
            printf("after file %s\n", frame->file->bytes);
            if (strcmp(mode, "throw") == 0) fail(frame, "boom");
            else finish(frame, frame->held, false);
            return;
        default: abort();
    }
    observe(awaited, resume, context);
    release((Object *)awaited); /* Awaited promise does not form a back edge from frame. */
}
static Promise *three_awaits(void) {
    Promise *output = promise_new();
    Frame *frame = allocate(sizeof(*frame), destroy_frame);
    frame->output = output; retain((Object *)output);
    Text *argument = text_new("held");
    frame->held = argument;
    retain((Object *)argument); /* Borrowed caller value must become frame-owned. */
    release((Object *)argument);
    resume((Object *)frame, NULL, false); /* JS async begins synchronously. */
    release((Object *)frame); /* Pending reaction now owns the frame. */
    return output;
}
static void root_complete(Object *context, Text *value, bool rejected) {
    (void)context; printf("%s %s\n", rejected ? "caught" : "done", value->bytes);
}
static void top_microtask(Object *context, Text *value, bool rejected) {
    (void)context; (void)value; (void)rejected; puts("top microtask");
}
int main(int argc, char **argv) {
    if (argc != 3) return 2;
    mode = argv[1]; path = argv[2];
    if (pipe(completion_pipe) != 0) abort();
    Promise *output = three_awaits(); observe(output, root_complete, NULL);
    enqueue(job_new(top_microtask, NULL, NULL, false)); puts("end");
    event_loop(); release((Object *)output);
    close(completion_pipe[0]); close(completion_pipe[1]);
    fprintf(stderr, "counts allocations=%zu frees=%zu retains=%zu releases=%zu\n", allocations, frees, retains, releases);
    return 0; /* LSan, independently of counts, owns the leak verdict. */
}
