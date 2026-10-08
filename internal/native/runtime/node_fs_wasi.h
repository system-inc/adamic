#ifndef ADAMIC_NODE_FS_WASI_H
#define ADAMIC_NODE_FS_WASI_H

#ifdef ADAMIC_TARGET_WASI
// libuv exposes Linux system-error numbers even when libc uses WASI errno.
static inline int adamic_node_fs_errno(int error) {
    switch (error) {
    case EPERM: return -1;
    case ENOENT: return -2;
    case EINTR: return -4;
    case EIO: return -5;
    case ENXIO: return -6;
    case EBADF: return -9;
    case EAGAIN: return -11;
    case ENOMEM: return -12;
    case EACCES: return -13;
    case EFAULT: return -14;
    case EEXIST: return -17;
    case ENOTDIR: return -20;
    case EISDIR: return -21;
    case EINVAL: return -22;
    case ENFILE: return -23;
    case EMFILE: return -24;
    case EFBIG: return -27;
    case ENOSPC: return -28;
    case EROFS: return -30;
    case ENAMETOOLONG: return -36;
    case ENOSYS: return -38;
    case ENOTEMPTY: return -39;
    case ELOOP: return -40;
    case EOVERFLOW: return -75;
    default: return -error;
    }
}

#else
static inline int adamic_node_fs_errno(int error) { return -error; }
#endif

#ifdef ADAMIC_TARGET_WASI
#include <fcntl.h>
#include <time.h>
// wasi-libc normalizes dot components before looking up a path. Resolve the
// parent first, so a symlink before '..' is followed in kernel order. Keep the
// leaf unchanged: unlink/lstat must not follow it, and creation may name a leaf
// that does not exist. Original spellings stay available for Node error text.
static inline char *adamic_node_fs_wasi_path(const char *name) {
    if (name[0] == 0) { errno = ENOENT; return NULL; }
    const char *slash = strrchr(name, '/');
    char *parent;
    const char *leaf = slash == NULL ? name : slash + 1;
    if (slash == NULL) { parent = realpath(".", NULL); }
    else {
        size_t length = (size_t)(slash - name);
        char *prefix = malloc(length + 2);
        if (prefix == NULL) { errno = ENOMEM; return NULL; }
        if (length == 0) { prefix[length++] = '/'; }
        else { memcpy(prefix, name, length); }
        prefix[length] = 0;
        parent = realpath(prefix, NULL);
        int error = errno; free(prefix); errno = error;
    }
    if (parent == NULL) { return NULL; }
    struct stat information;
    if (stat(parent, &information) != 0) {
        int error = errno; free(parent); errno = error; return NULL;
    }
    if (!S_ISDIR(information.st_mode)) { free(parent); errno = ENOTDIR; return NULL; }
    size_t length = strlen(parent), rest = strlen(leaf);
    char *resolved = malloc(length + rest + 2);
    if (resolved == NULL) { free(parent); errno = ENOMEM; return NULL; }
    memcpy(resolved, parent, length); resolved[length] = '/';
    memcpy(resolved + length + 1, leaf, rest + 1); free(parent);
    return resolved;
}

#define ADAMIC_WASI_PATH_OPERATION(name, parameters, arguments) \
    static inline int adamic_node_wasi_##name parameters { \
        char *resolved = adamic_node_fs_wasi_path(path); \
        if (resolved == NULL) { return -1; } \
        int result = name arguments; \
        int error = errno; free(resolved); errno = error; return result; \
    }
ADAMIC_WASI_PATH_OPERATION(stat, (const char *path, struct stat *info), (resolved, info))
ADAMIC_WASI_PATH_OPERATION(lstat, (const char *path, struct stat *info), (resolved, info))
ADAMIC_WASI_PATH_OPERATION(mkdir, (const char *path, mode_t mode), (resolved, mode))
ADAMIC_WASI_PATH_OPERATION(unlink, (const char *path), (resolved))
ADAMIC_WASI_PATH_OPERATION(rmdir, (const char *path), (resolved))
ADAMIC_WASI_PATH_OPERATION(open, (const char *path, int flags, mode_t mode), (resolved, flags, mode))
ADAMIC_WASI_PATH_OPERATION(utimensat, (int fd, const char *path, const struct timespec times[2], int flags), (fd, resolved, times, flags))
#undef ADAMIC_WASI_PATH_OPERATION

// Node on Linux reports ENOENT for an empty symlink target. Preview 1
// reports EINVAL instead; the input distinguishes this case without masking
// any other syscall failure.
static inline int adamic_node_wasi_symlink(const char *target, const char *path) {
    int result = symlink(target, path);
    if (result != 0 && target[0] == 0 && errno == EINVAL) { errno = ENOENT; }
    return result;
}

static inline DIR *adamic_node_wasi_opendir(const char *path) {
    char *resolved = adamic_node_fs_wasi_path(path);
    if (resolved == NULL) { return NULL; }
    DIR *result = opendir(resolved);
    int error = errno; free(resolved); errno = error; return result;
}
#define stat(...) adamic_node_wasi_stat(__VA_ARGS__)
#define lstat(...) adamic_node_wasi_lstat(__VA_ARGS__)
#define mkdir(...) adamic_node_wasi_mkdir(__VA_ARGS__)
#define unlink(...) adamic_node_wasi_unlink(__VA_ARGS__)
#define rmdir(...) adamic_node_wasi_rmdir(__VA_ARGS__)
#define open(...) adamic_node_wasi_open(__VA_ARGS__)
#define utimensat(...) adamic_node_wasi_utimensat(__VA_ARGS__)
#define symlink(...) adamic_node_wasi_symlink(__VA_ARGS__)
#define opendir(...) adamic_node_wasi_opendir(__VA_ARGS__)
#endif
#endif
