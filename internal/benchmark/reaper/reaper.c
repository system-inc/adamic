// reaper.c: the process internal/benchmark starts every program through on Linux, for bench and
// cmd/adamic-performance alike, so the peak resident set it reads is the program's own.
//
//	reaper <fd> <program> [arguments...]
//
// Go starts a child with clone(CLONE_VFORK|CLONE_VM), and Linux's exec folds the old memory's
// high-water resident set into the new program's maximum (fs/exec.c, exec_mmap), so the old memory,
// the measuring process's own, set a floor under every program's peak: a bench holding 250 MB of inputs
// read 1321.4 MB for every program, one in C that only prints among them (J-wme5jp2). A fork from any
// large process carries the same. Between the measuring process and the program the reaper is the old
// memory, and its resident set is a few hundred kilobytes. It lives in its own directory because a .c
// file beside internal/benchmark's Go files would make it a cgo package, which the sealed jobs' Go
// refuses (J-gx6b2hv) and nothing that measures wants.
//
// It forks, and the child runs the program, which inherits standard input, output and error. It waits
// for the program, writes one line to <fd>, "<user microseconds> <system microseconds> <maxrss
// kilobytes>", and exits with the program's exit code, or 128 plus the signal that killed it. It
// allocates nothing beyond its stack. internal/benchmark embeds it, BuildReaper builds it with clang
// -O2, and Run reads the line.

#define _GNU_SOURCE
#include <errno.h>
#include <limits.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/prctl.h>
#include <sys/resource.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

// reaper_failed is the reaper's own failure, apart from 127, the program not starting.
static const int reaper_failed = 125;

// microseconds is a time from rusage in whole microseconds.
static long long microseconds(struct timeval value) {
	return (long long)value.tv_sec * 1000000LL + (long long)value.tv_usec;
}

int main(int argc, char **argv) {
	if (argc < 3) {
		fprintf(stderr, "usage: reaper <fd> <program> [arguments...]\n");
		return reaper_failed;
	}
	char *end = NULL;
	errno = 0;
	const long parsed = strtol(argv[1], &end, 10);
	if (errno != 0 || end == argv[1] || *end != '\0' || parsed < 3 || parsed > INT_MAX) {
		fprintf(stderr, "reaper: %s is not a descriptor beside standard input, output and error\n", argv[1]);
		return reaper_failed;
	}
	const int descriptor = (int)parsed;

	const pid_t reaper = getpid();
	const pid_t child = fork();
	if (child < 0) {
		perror("reaper: fork");
		return reaper_failed;
	}
	if (child == 0) {
		// The program dies with the reaper, so a timeout that kills the reaper kills the program.
		if (prctl(PR_SET_PDEATHSIG, SIGKILL) != 0) {
			perror("reaper: prctl");
			_exit(127);
		}
		// The reaper may have died before the request took, and then the parent is already another.
		if (getppid() != reaper) {
			_exit(127);
		}
		// The line is the reaper's to write; the program starts with the descriptors it would have had started directly.
		(void)close(descriptor);
		execvp(argv[2], argv + 2);
		fprintf(stderr, "reaper: %s: %s\n", argv[2], strerror(errno));
		_exit(127);
	}

	int status = 0;
	struct rusage usage;
	pid_t waited;
	do {
		waited = wait4(child, &status, 0, &usage);
	} while (waited < 0 && errno == EINTR);
	if (waited != child) {
		perror("reaper: wait4");
		return reaper_failed;
	}

	// A line that isn't written whole is no line: Run reads it as a failed measurement, never a zero.
	char line[96];
	const int length = snprintf(line, sizeof line, "%lld %lld %ld\n", microseconds(usage.ru_utime), microseconds(usage.ru_stime), usage.ru_maxrss);
	if (length < 0 || length >= (int)sizeof line) {
		fprintf(stderr, "reaper: the usage line didn't fit\n");
	} else {
		for (int offset = 0; offset < length;) {
			const ssize_t wrote = write(descriptor, line + offset, (size_t)(length - offset));
			if (wrote < 0 && errno == EINTR) {
				continue;
			}
			if (wrote <= 0) {
				perror("reaper: writing the usage line");
				break;
			}
			offset += (int)wrote;
		}
	}
	(void)close(descriptor);

	if (WIFEXITED(status)) {
		return WEXITSTATUS(status);
	}
	if (WIFSIGNALED(status)) {
		return 128 + WTERMSIG(status);
	}
	return reaper_failed;
}
