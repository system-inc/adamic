// Linux test interposition: short successful I/O on regular files only.
// stdout/stderr and nonregular descriptors keep their normal behavior.
#define _GNU_SOURCE
#include <dlfcn.h>
#include <stddef.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

static int regular(int fd) {
    struct stat info;
    return fd > 2 && fstat(fd, &info) == 0 && S_ISREG(info.st_mode);
}
ssize_t read(int fd, void *data, size_t length) {
    static ssize_t (*real_read)(int, void *, size_t);
    if (real_read == NULL) {
        void *symbol = dlsym(RTLD_NEXT, "read");
        memcpy(&real_read, &symbol, sizeof real_read);
    }
    if (regular(fd) && length > 257) { length = 257; }
    return real_read(fd, data, length);
}
ssize_t write(int fd, const void *data, size_t length) {
    static ssize_t (*real_write)(int, const void *, size_t);
    if (real_write == NULL) {
        void *symbol = dlsym(RTLD_NEXT, "write");
        memcpy(&real_write, &symbol, sizeof real_write);
    }
    if (regular(fd) && length > 3) { length = 3; }
    return real_write(fd, data, length);
}
