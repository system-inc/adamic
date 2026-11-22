// spawnloop: starts each binary named, in turn, rounds times, interleaved, its stdout a pipe this process drains, and
// prints each binary's wall per run in microseconds: the median, the 10th and 90th percentiles and the mean. #7z36qn6's
// lever 2 measure; never lands.
#define _POSIX_C_SOURCE 200809L
#include <spawn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <time.h>
#include <unistd.h>

extern char **environ;

static int compare(const void *left, const void *right) {
	double a = *(const double *)left, b = *(const double *)right;
	return a < b ? -1 : a > b;
}

static double now(void) {
	struct timespec moment;
	clock_gettime(CLOCK_MONOTONIC, &moment);
	return (double)moment.tv_sec * 1e6 + (double)moment.tv_nsec / 1e3;
}

int main(int count, char **values) {
	if (count < 4) {
		fprintf(stderr, "usage: spawnloop <rounds> <pipe|null> <binary>...\n");
		return 2;
	}
	int rounds = atoi(values[1]);
	int pipe_mode = strcmp(values[2], "pipe") == 0;
	int binaries = count - 3;
	double *samples = calloc((size_t)rounds * (size_t)binaries, sizeof *samples);
	for (int round = 0; round < rounds; round++) {
		for (int turn = 0; turn < binaries; turn++) {
			// Mirror the order on odd rounds, so neither binary always runs first.
			int index = round % 2 == 0 ? turn : binaries - 1 - turn;
			char *arguments[] = {values[3 + index], NULL};
			posix_spawn_file_actions_t actions;
			posix_spawn_file_actions_init(&actions);
			int ends[2] = {-1, -1};
			if (pipe_mode) {
				if (pipe(ends) != 0) {
					perror("pipe");
					return 1;
				}
				posix_spawn_file_actions_adddup2(&actions, ends[1], 1);
				posix_spawn_file_actions_addclose(&actions, ends[0]);
				posix_spawn_file_actions_addclose(&actions, ends[1]);
			} else {
				posix_spawn_file_actions_addopen(&actions, 1, "/dev/null", 1, 0);
			}
			double started = now();
			pid_t child;
			if (posix_spawn(&child, arguments[0], &actions, NULL, arguments, environ) != 0) {
				perror(arguments[0]);
				return 1;
			}
			if (pipe_mode) {
				close(ends[1]);
				char buffer[4096];
				while (read(ends[0], buffer, sizeof buffer) > 0) {
				}
				close(ends[0]);
			}
			int status;
			waitpid(child, &status, 0);
			double elapsed = now() - started;
			posix_spawn_file_actions_destroy(&actions);
			if (!WIFEXITED(status) || WEXITSTATUS(status) != 0) {
				fprintf(stderr, "%s failed: status %d\n", arguments[0], status);
				return 1;
			}
			samples[(size_t)index * (size_t)rounds + (size_t)round] = elapsed;
		}
	}
	for (int index = 0; index < binaries; index++) {
		double *own = samples + (size_t)index * (size_t)rounds;
		double sum = 0;
		for (int round = 0; round < rounds; round++) {
			sum += own[round];
		}
		qsort(own, (size_t)rounds, sizeof *own, compare);
		printf("%-14s %s median %7.1f us  p10 %7.1f  p90 %7.1f  mean %7.1f  (%d runs)\n", strrchr(values[3 + index], '/') + 1, values[2], own[rounds / 2], own[rounds / 10], own[rounds * 9 / 10], sum / rounds, rounds);
	}
	return 0;
}
