// unwinding.c: the two ways to unwind through reference counting (docs/memory.md, "Exceptions,
// designed into counting"), measured on one workload in plain C, so the mechanisms are compared and
// nothing else is.
//
//	clang -O2 -o unwinding unwinding.c && ./unwinding
//
// The workload is trees.ts's shape: a recursive function makes a binary tree of counted nodes, each
// frame owning a reference while it calls down, and the tree is let go. It runs three ways: with no
// exceptions at all; by cleanup paths, where a call that can throw is followed by a test of a pending
// word and a frame that sees it set lets go of what it owns and returns; and by setjmp, where each try
// is a setjmp and every owned reference is pushed on a cleanup stack as it's taken and popped as it's
// let go, so a longjmp can release what's above the try's mark. Each runs with no throw, and with a
// throw from the deepest frame every 64th tree.

#include <setjmp.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

typedef struct node {
	size_t references;
	struct node *left, *right;
} node;

static size_t live;

static node *make_node(node *left, node *right) {
	node *made = malloc(sizeof *made);
	made->references = 1;
	made->left = left;
	made->right = right;
	live++;
	return made;
}

static void release(node *value) {
	// Recursion is fine here: the trees are shallow, and both ways free the same way.
	if (value != NULL && --value->references == 0) {
		release(value->left);
		release(value->right);
		free(value);
		live--;
	}
}

static int throw_at = -1;

// No exceptions.
static node *plain(int depth) {
	if (depth == 0) {
		return make_node(NULL, NULL);
	}
	node *left = plain(depth - 1);
	node *right = plain(depth - 1);
	return make_node(left, right);
}

// Cleanup paths: a pending word, tested after each call that can throw.
static void *thrown;

static node *cleanup(int depth) {
	if (depth == 0) {
		if (throw_at == 0) {
			thrown = &throw_at;
			return NULL;
		}
		return make_node(NULL, NULL);
	}
	node *left = cleanup(depth - 1);
	if (thrown != NULL) {
		return NULL;
	}
	node *right = cleanup(depth - 1);
	if (thrown != NULL) {
		release(left);
		return NULL;
	}
	return make_node(left, right);
}

// setjmp and a cleanup stack: every owned reference is on the stack while its frame holds it.
static node *stack[64];
static size_t stack_count;
static jmp_buf *handler;

static node *jumping(int depth) {
	if (depth == 0) {
		if (throw_at == 0) {
			longjmp(*handler, 1);
		}
		return make_node(NULL, NULL);
	}
	node *left = jumping(depth - 1);
	stack[stack_count++] = left;
	node *right = jumping(depth - 1);
	stack_count--;
	return make_node(left, right);
}

// The same three ways with no allocation, so only the mechanisms are left: a call-heavy recursion in
// which each frame retains one counted object while it calls down, and releases it after.
static node shared = {1, NULL, NULL};

static double plain_calls(int depth) {
	if (depth == 0) {
		return 1;
	}
	shared.references++;
	double sum = plain_calls(depth - 1) + plain_calls(depth - 1);
	shared.references--;
	return sum;
}

static double cleanup_calls(int depth) {
	if (depth == 0) {
		if (throw_at == 0) {
			thrown = &throw_at;
		}
		return 1;
	}
	shared.references++;
	double left = cleanup_calls(depth - 1);
	if (thrown != NULL) {
		shared.references--;
		return 0;
	}
	double right = cleanup_calls(depth - 1);
	if (thrown != NULL) {
		shared.references--;
		return 0;
	}
	shared.references--;
	return left + right;
}

static double jumping_calls(int depth) {
	if (depth == 0) {
		if (throw_at == 0) {
			longjmp(*handler, 1);
		}
		return 1;
	}
	shared.references++;
	stack[stack_count++] = &shared;
	double sum = jumping_calls(depth - 1) + jumping_calls(depth - 1);
	stack_count--;
	shared.references--;
	return sum;
}

static double seconds(void) {
	struct timespec now;
	timespec_get(&now, TIME_UTC);
	return (double)now.tv_sec + (double)now.tv_nsec / 1e9;
}

enum { trees = 4000, depth = 14 };

static double run(int way, bool throwing) {
	double start = seconds();
	for (int tree = 0; tree < trees; tree++) {
		throw_at = throwing && tree % 64 == 0 ? 0 : -1;
		// volatile: it's read after a longjmp may have come back past the setjmp.
		node *volatile made = NULL;
		switch (way) {
		case 0:
			made = plain(depth);
			break;
		case 1:
			made = cleanup(depth);
			thrown = NULL;
			break;
		case 2: {
			jmp_buf here;
			size_t mark = stack_count;
			handler = &here;
			if (setjmp(here) == 0) {
				made = jumping(depth);
			} else {
				while (stack_count > mark) {
					release(stack[--stack_count]);
				}
			}
			break;
		}
		}
		release(made);
	}
	return seconds() - start;
}

static volatile double sink;

static double run_calls(int way, bool throwing) {
	double start = seconds();
	for (int round = 0; round < trees; round++) {
		throw_at = throwing && round % 64 == 0 ? 0 : -1;
		switch (way) {
		case 0:
			sink = plain_calls(depth);
			break;
		case 1:
			sink = cleanup_calls(depth);
			thrown = NULL;
			break;
		case 2: {
			jmp_buf here;
			size_t mark = stack_count;
			handler = &here;
			if (setjmp(here) == 0) {
				sink = jumping_calls(depth);
			} else {
				// What each frame held is let go of from the stack.
				while (stack_count > mark) {
					stack[--stack_count]->references--;
				}
			}
			break;
		}
		}
	}
	return seconds() - start;
}

int main(void) {
	static const char *const ways[] = {"no exceptions", "cleanup paths", "setjmp and a cleanup stack"};
	double best[2][3][2];
	for (int workload = 0; workload < 2; workload++) {
		for (int way = 0; way < 3; way++) {
			best[workload][way][0] = best[workload][way][1] = 1e9;
		}
	}
	for (int round = 0; round < 7; round++) {
		for (int workload = 0; workload < 2; workload++) {
			for (int way = 0; way < 3; way++) {
				for (int throwing = 0; throwing < 2; throwing++) {
					if (way == 0 && throwing) {
						continue;
					}
					double took = workload == 0 ? run(way, throwing) : run_calls(way, throwing);
					if (took < best[workload][way][throwing]) {
						best[workload][way][throwing] = took;
					}
				}
			}
		}
	}
	if (live != 0) {
		printf("leaked %zu nodes\n", live);
		return 1;
	}
	if (shared.references != 1) {
		printf("counted wrong: %zu\n", shared.references);
		return 1;
	}
	static const char *const workloads[] = {"trees: allocating", "calls: no allocation"};
	printf("%d rounds of depth %d, best of 7\n", trees, depth);
	for (int workload = 0; workload < 2; workload++) {
		printf("%s\n", workloads[workload]);
		for (int way = 0; way < 3; way++) {
			printf("  %-28s %.3f s", ways[way], best[workload][way][0]);
			if (way > 0) {
				printf(", %.3f s throwing every 64th round", best[workload][way][1]);
			}
			printf("\n");
		}
	}
	return 0;
}
