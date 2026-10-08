// Scheduling perturbations for the race tests only. No synchronization is added.
#ifndef ADAMIC_TSAN_TEST_H
#define ADAMIC_TSAN_TEST_H
#ifdef ADAMIC_TSAN_TEST
#if !defined(__has_feature)
#error ADAMIC_TSAN_TEST requires ThreadSanitizer
#elif !__has_feature(thread_sanitizer)
#error ADAMIC_TSAN_TEST requires ThreadSanitizer
#endif
enum adamic_tsan_point {
	adamic_tsan_plain_count,
	adamic_tsan_shared_count,
	adamic_tsan_remote_free,
	adamic_tsan_publication,
	adamic_tsan_claim,
	adamic_tsan_cache_publication,
	adamic_tsan_points,
};
void adamic_tsan_pause(enum adamic_tsan_point point);
#define ADAMIC_TSAN_PAUSE(point) adamic_tsan_pause(point)
#else
#define ADAMIC_TSAN_PAUSE(point) ((void)0)
#endif
#endif
