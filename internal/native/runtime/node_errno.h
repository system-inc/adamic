#ifndef ADAMIC_NODE_ERRNO_H
#define ADAMIC_NODE_ERRNO_H

#include <errno.h>

// adamic_node_errno is the errno number Node gives an error of this libc errno: libuv's, the host's
// errno negated. Native is built for its host, so that's -error. A wasm32-wasi module runs on every
// host and can't see which, and wasi-libc's constants are WASI's own numbering (ENOENT is 44), so under
// ADAMIC_TARGET_WASI each code node_fs_file.c's system_error_words names maps to Linux's number, Node's
// on the oracle's host and on every box, and anything else to libuv's UV_UNKNOWN, -4094, beside the
// code UNKNOWN that table gives it (Adamic on #z4a9p9e, Nov 13 2026). Linux's numbers are Node's on
// macOS too except for EAGAIN, ENAMETOOLONG, ENOSYS, ENOTEMPTY and ELOOP, and are never Windows':
// docs/wasm-limits.md, "The host's errno numbers". internal/oracle/wasi_errno_test.go holds the table
// to Node on Linux.
static inline double adamic_node_errno(int error) {
#ifdef ADAMIC_TARGET_WASI
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
    default: return -4094;
    }
#else
    return -(double)error;
#endif
}

#endif
