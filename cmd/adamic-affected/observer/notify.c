#define _GNU_SOURCE
#include <errno.h>
#include <fcntl.h>
#include <linux/audit.h>
#include <linux/filter.h>
#include <linux/seccomp.h>
#include <stddef.h>
#include <signal.h>
#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <sys/poll.h>
#include <sys/prctl.h>
#include <sys/socket.h>
#include <sys/syscall.h>
#include <sys/types.h>
#include <sys/uio.h>
#include <sys/wait.h>
#include <unistd.h>

/* The kernel blocks every matching syscall until its notification is consumed.
 * Children inherit the filter through fork and exec; no ptrace state is set.
 * CONTINUE observes syscall entry, as strace's pathname decoding does. It does
 * not emulate a syscall or change the caller's file operation. */
static int install_filter(void) {
#if defined(__x86_64__)
 const unsigned architecture = AUDIT_ARCH_X86_64;
#elif defined(__aarch64__)
 const unsigned architecture = AUDIT_ARCH_AARCH64;
#else
 errno = ENOTSUP; return -1;
#endif
 struct sock_filter code[128];
 size_t count = 0;
 code[count++] = (struct sock_filter) BPF_STMT(BPF_LD|BPF_W|BPF_ABS, offsetof(struct seccomp_data, arch));
 code[count++] = (struct sock_filter) BPF_JUMP(BPF_JMP|BPF_JEQ|BPF_K, architecture, 1, 0);
 code[count++] = (struct sock_filter) BPF_STMT(BPF_RET|BPF_K, SECCOMP_RET_USER_NOTIF);
 code[count++] = (struct sock_filter) BPF_STMT(BPF_LD|BPF_W|BPF_ABS, offsetof(struct seccomp_data, nr));
 code[count++] = (struct sock_filter) BPF_JUMP(BPF_JMP|BPF_JSET|BPF_K, 0x40000000U, 0, 1);
 code[count++] = (struct sock_filter) BPF_STMT(BPF_RET|BPF_K, SECCOMP_RET_USER_NOTIF);
 const int calls[] = {
#ifdef SYS_open
 SYS_open,
#endif
 SYS_openat,
#ifdef SYS_io_uring_setup
 SYS_io_uring_setup, SYS_io_uring_enter, SYS_io_uring_register,
#endif
#ifdef SYS_openat2
 SYS_openat2,
#endif
#ifdef SYS_stat
 SYS_stat, SYS_lstat,
#endif
#ifdef SYS_newfstatat
 SYS_newfstatat,
#endif
#ifdef SYS_fstatat
 SYS_fstatat,
#endif
#ifdef SYS_statx
 SYS_statx,
#endif
#ifdef SYS_access
 SYS_access,
#endif
 SYS_faccessat,
#ifdef SYS_faccessat2
 SYS_faccessat2,
#endif
#ifdef SYS_readlink
 SYS_readlink,
#endif
 SYS_readlinkat, SYS_execve,
#ifdef SYS_execveat
 SYS_execveat,
#endif
#ifdef SYS_getdents
 SYS_getdents,
#endif
 SYS_getdents64,
#ifdef SYS_rename
 SYS_rename, SYS_link, SYS_symlink,
#endif
 SYS_renameat, SYS_linkat, SYS_symlinkat,
#ifdef SYS_renameat2
 SYS_renameat2,
#endif
 };
 for (size_t index = 0; index < sizeof(calls)/sizeof(calls[0]); index++) {
  code[count++] = (struct sock_filter) BPF_JUMP(BPF_JMP|BPF_JEQ|BPF_K, (unsigned)calls[index], 0, 1);
  code[count++] = (struct sock_filter) BPF_STMT(BPF_RET|BPF_K, SECCOMP_RET_USER_NOTIF);
 }
 code[count++] = (struct sock_filter) BPF_STMT(BPF_RET|BPF_K, SECCOMP_RET_ALLOW);
 struct sock_fprog filter = {(unsigned short)count, code};
 if (prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0) < 0) return -1;
 return (int)syscall(SYS_seccomp, SECCOMP_SET_MODE_FILTER, SECCOMP_FILTER_FLAG_NEW_LISTENER, &filter);
}
static int send_descriptor(int socket, int descriptor) {
 char payload = 'F'; struct iovec vector = {&payload, 1};
 union {struct cmsghdr alignment; char bytes[CMSG_SPACE(sizeof(int))];} control;
 memset(&control, 0, sizeof(control));
 struct msghdr message = {0}; message.msg_iov = &vector; message.msg_iovlen = 1;
 message.msg_control = control.bytes; message.msg_controllen = sizeof(control.bytes);
 struct cmsghdr *header = CMSG_FIRSTHDR(&message);
 header->cmsg_level = SOL_SOCKET; header->cmsg_type = SCM_RIGHTS; header->cmsg_len = CMSG_LEN(sizeof(int));
 memcpy(CMSG_DATA(header), &descriptor, sizeof(descriptor));
 return sendmsg(socket, &message, 0) == 1 ? 0 : -1;
}
static int receive_descriptor(int socket) {
 char payload; struct iovec vector = {&payload, 1};
 union {struct cmsghdr alignment; char bytes[CMSG_SPACE(sizeof(int))];} control;
 struct msghdr message = {0}; message.msg_iov = &vector; message.msg_iovlen = 1;
 message.msg_control = control.bytes; message.msg_controllen = sizeof(control.bytes);
 if (recvmsg(socket, &message, 0) != 1 || (message.msg_flags & MSG_CTRUNC)) return -1;
 struct cmsghdr *header = CMSG_FIRSTHDR(&message);
 if (!header || header->cmsg_level != SOL_SOCKET || header->cmsg_type != SCM_RIGHTS || header->cmsg_len != CMSG_LEN(sizeof(int))) return -1;
 int descriptor; memcpy(&descriptor, CMSG_DATA(header), sizeof(descriptor)); return descriptor;
}
static int remote_string(pid_t process, uint64_t address, char *buffer, size_t size) {
 size_t offset = 0;
 while (offset + 1 < size) {
  size_t length = size - offset - 1; if (length > 256) length = 256;
  struct iovec local = {buffer + offset, length};
  struct iovec remote = {(void *)(uintptr_t)(address + offset), length};
  ssize_t got = process_vm_readv(process, &local, 1, &remote, 1, 0);
  if (got <= 0) {
   char memory_path[80]; snprintf(memory_path, sizeof(memory_path), "/proc/%d/mem", process);
   int memory = open(memory_path, O_RDONLY | O_CLOEXEC);
   if (memory < 0) return -1;
   got = pread(memory, buffer + offset, length, (off_t)(address + offset));
   close(memory);
   if (got <= 0) return -1;
  }
  if (memchr(buffer + offset, 0, (size_t)got)) return 0;
  offset += (size_t)got;
 }
 return -1;
}
static int descriptor_path(pid_t process, int descriptor, char *buffer, size_t size) {
 char link[128];
 if (descriptor == AT_FDCWD) snprintf(link, sizeof(link), "/proc/%d/cwd", process);
 else snprintf(link, sizeof(link), "/proc/%d/fd/%d", process, descriptor);
 ssize_t length = readlink(link, buffer, size - 1);
 if (length < 0 || (size_t)length >= size - 1) return -1;
 buffer[length] = 0; return 0;
}
static void quoted(FILE *log, const char *text) {
 fputc('"', log);
 for (const unsigned char *cursor = (const unsigned char *)text; *cursor; cursor++) {
  if (*cursor == '"' || *cursor == '\\') fprintf(log, "\\%c", *cursor);
  else if (*cursor < 32 || *cursor >= 127) fprintf(log, "\\%03o", *cursor);
  else fputc(*cursor, log);
 }
 fputc('"', log);
}
static void observe(FILE *log, const struct seccomp_notif *request) {
#if defined(__x86_64__)
 if (request->data.arch != AUDIT_ARCH_X86_64 || ((unsigned)request->data.nr & 0x40000000U)) {fputs("uncertain foreign syscall architecture\n",log); return;}
#elif defined(__aarch64__)
 if (request->data.arch != AUDIT_ARCH_AARCH64) {fputs("uncertain foreign syscall architecture\n",log); return;}
#endif
 int call = request->data.nr; const __u64 *args = request->data.args;
 const char *name = NULL; int directory = AT_FDCWD; uint64_t pointer = 0; uint64_t flags = 0; int listing = 0;
#define AT_CALL(number, label) if (call == number) {name=label; directory=(int)args[0]; pointer=args[1];}
#define PATH_CALL(number, label) if (call == number) {name=label; pointer=args[0];}
 AT_CALL(SYS_openat, "openat"); if (call == SYS_openat) flags = args[2];
#ifdef SYS_open
 PATH_CALL(SYS_open, "open"); if (call == SYS_open) flags = args[1];
#endif
#ifdef SYS_openat2
 /* Its flags live in a remote struct: uncertain until that struct is decoded. */
 if (call == SYS_openat2) {fprintf(log, "uncertain openat2\n"); return;}
#endif
#ifdef SYS_stat
 PATH_CALL(SYS_stat, "stat"); PATH_CALL(SYS_lstat, "lstat");
#endif
#ifdef SYS_newfstatat
 AT_CALL(SYS_newfstatat, "newfstatat");
#endif
#ifdef SYS_fstatat
 AT_CALL(SYS_fstatat, "fstatat");
#endif
#ifdef SYS_statx
 AT_CALL(SYS_statx, "statx");
#endif
#ifdef SYS_access
 PATH_CALL(SYS_access, "access");
#endif
 AT_CALL(SYS_faccessat, "faccessat");
#ifdef SYS_faccessat2
 AT_CALL(SYS_faccessat2, "faccessat2");
#endif
#ifdef SYS_readlink
 PATH_CALL(SYS_readlink, "readlink");
#endif
 AT_CALL(SYS_readlinkat, "readlinkat"); PATH_CALL(SYS_execve, "execve");
#ifdef SYS_execveat
 AT_CALL(SYS_execveat, "execveat");
#endif
#ifdef SYS_getdents
 if (call == SYS_getdents) {name="getdents64"; directory=(int)args[0]; listing=1;}
#endif
 if (call == SYS_getdents64) {name="getdents64"; directory=(int)args[0]; listing=1;}
 char path[4096], base[4096];
 if (!name) {fprintf(log, "uncertain multi-path syscall %d\n", call); return;}
 if (descriptor_path((pid_t)request->pid, directory, base, sizeof(base)) < 0) {fprintf(log, "uncertain directory for %s\n", name); return;}
 if (strchr(base, '<') || strchr(base, '>') || strchr(base, '\\') || strchr(base, '\n')) {fprintf(log, "uncertain directory encoding\n"); return;}
 if (listing) {fprintf(log, "getdents64(%d<%s>, entries, bytes) = entry\n", directory, base); return;}
 if (remote_string((pid_t)request->pid, pointer, path, sizeof(path)) < 0) {fprintf(log, "uncertain pathname for %s\n", name); return;}
 /* Empty-path stat calls refer to the directory descriptor itself. */
 fprintf(log, "%s(AT_FDCWD<%s>, ", name, base); quoted(log, path);
 if ((flags & O_ACCMODE) == O_WRONLY) fputs(", O_WRONLY", log);
 else if ((flags & O_ACCMODE) == O_RDWR) fputs(", O_RDWR", log);
 else fputs(", O_RDONLY", log);
 if (flags & O_CREAT) fputs("|O_CREAT", log);
 if (flags & O_TRUNC) fputs("|O_TRUNC", log);
 /* Resolve persistent symlink aliases too, including /proc/self/fd paths.
  * The repository is snapshotted before the run and must remain unchanged. */
 char joined[8192], translated[8192], resolved[PATH_MAX];
 if (path[0] == '/') snprintf(joined, sizeof(joined), "%s", path);
 else snprintf(joined, sizeof(joined), "%s/%s", base, path);
 if (!strncmp(joined, "/proc/self/", 11)) snprintf(translated, sizeof(translated), "/proc/%u/%s", request->pid, joined+11);
 else if (!strncmp(joined, "/proc/thread-self/", 18)) snprintf(translated, sizeof(translated), "/proc/%u/%s", request->pid, joined+18);
 else snprintf(translated, sizeof(translated), "%s", joined);
 if (realpath(translated, resolved)) {
  if (strchr(resolved, '<') || strchr(resolved, '>') || strchr(resolved, '\\') || strchr(resolved, '\n')) {fputs(") = entry\nuncertain resolved pathname encoding\n", log); return;}
  fprintf(log, ") = entry<%s>\n", resolved);
 } else fprintf(log, ") = entry\n");
}
static volatile sig_atomic_t process_group = 0;
static volatile sig_atomic_t interrupted = 0;
static void stop_children(int number) {
 interrupted = number;
 if (process_group > 0) {kill(-process_group, SIGKILL); kill(process_group, SIGKILL);}
}
int main(int count, char **arguments) {
 if (count < 5 || strcmp(arguments[1], "-o") || strcmp(arguments[3], "--")) {fprintf(stderr,"usage: observer -o trace -- command [args]\n"); return 2;}
 FILE *log = fopen(arguments[2], "w"); if (!log) {perror("trace"); return 2;}
 int sockets[2]; if (socketpair(AF_UNIX, SOCK_DGRAM|SOCK_CLOEXEC, 0, sockets) < 0) {perror("socketpair"); return 2;}
 pid_t child = fork(); if (child < 0) {perror("fork"); return 2;}
 if (child == 0) {
  setpgid(0,0); prctl(PR_SET_PDEATHSIG, SIGKILL,0,0,0);
  close(sockets[0]); fclose(log);
  int listener = install_filter(); if (listener < 0) {perror("seccomp listener"); _exit(125);}
  if (send_descriptor(sockets[1], listener) < 0) {perror("listener transfer"); _exit(125);}
  close(listener); close(sockets[1]); execvp(arguments[4], arguments+4); perror("exec"); _exit(127);
 }
 process_group = child; setpgid(child,child);
 struct sigaction action = {0}; action.sa_handler = stop_children; sigemptyset(&action.sa_mask);
 sigaction(SIGTERM,&action,NULL); sigaction(SIGINT,&action,NULL); sigaction(SIGHUP,&action,NULL); sigaction(SIGALRM,&action,NULL);
 alarm(65*60);
 close(sockets[1]);
 struct pollfd transfer = {sockets[0], POLLIN, 0};
 if (poll(&transfer, 1, 10000) <= 0) {fprintf(stderr,"listener transfer timed out\n"); return 2;}
 int listener = receive_descriptor(sockets[0]); close(sockets[0]);
 if (listener < 0) {fprintf(stderr,"listener transfer failed\n"); return 2;}
 struct seccomp_notif_sizes sizes;
 if (syscall(SYS_seccomp, SECCOMP_GET_NOTIF_SIZES, 0, &sizes) < 0) {perror("notification sizes"); return 2;}
 struct seccomp_notif *request = calloc(1, sizes.seccomp_notif);
 struct seccomp_notif_resp *response = calloc(1, sizes.seccomp_notif_resp);
 if (!request || !response) return 2;
 int status = 0, finished = 0;
 for (;;) {
  if (interrupted) {waitpid(child,&status,0); return 128+interrupted;}
  if (!finished) {
   pid_t completed = waitpid(child, &status, WNOHANG);
   if (completed == child) finished = 1;
   if (completed < 0 && errno != EINTR) {perror("waitpid"); stop_children(SIGTERM); return 2;}
  }
  struct pollfd ready = {listener, POLLIN, 0};
  int available = poll(&ready, 1, 100);
  if (available < 0 && errno == EINTR) continue;
  if (available < 0) {perror("notification poll"); return 2;}
  if (available == 0) continue;
  if (finished && (ready.revents & POLLHUP) && !(ready.revents & POLLIN)) break;
  memset(request, 0, sizes.seccomp_notif);
  if (ioctl(listener, SECCOMP_IOCTL_NOTIF_RECV, request) < 0) {
   /* A withdrawn notification never received CONTINUE, so its syscall did
    * not execute. A restarted syscall creates a fresh notification. */
   if (errno == EINTR || errno == ENOENT) continue;
   perror("notification receive"); return 2;
  }
  observe(log, request); if (fflush(log) != 0) {perror("trace flush"); return 2;}
  memset(response, 0, sizes.seccomp_notif_resp); response->id=request->id; response->flags=SECCOMP_USER_NOTIF_FLAG_CONTINUE;
  if (ioctl(listener, SECCOMP_IOCTL_NOTIF_SEND, response) < 0) {
   if (errno == ENOENT) continue; /* Already recorded; operation was withdrawn. */
   perror("notification continue"); return 2;
  }
 }
 if (fclose(log) != 0) {perror("trace close"); return 2;}
 close(listener); free(request); free(response);
 if (WIFEXITED(status)) return WEXITSTATUS(status);
 if (WIFSIGNALED(status)) return 128+WTERMSIG(status);
 return 2;
}
