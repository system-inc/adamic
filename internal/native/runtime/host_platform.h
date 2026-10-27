// host_platform.h: the platform names Node gives process.platform, from uname()'s sysname. Only the
// names that lowercase to Node's own are here; any other sysname has no answer, because Node maps
// some (to android or cygwin, say) and a lowercased guess would be a silent wrong one. It has no
// runtime dependency, so a test compiles it alone (TestHostPlatformNames).
#ifndef ADAMIC_HOST_PLATFORM_H
#define ADAMIC_HOST_PLATFORM_H
#include <string.h>

static inline const char *adamic_host_platform_name(const char *sysname) {
	static const char *const names[][2] = {
		{"Linux", "linux"}, {"Darwin", "darwin"}, {"FreeBSD", "freebsd"},
		{"OpenBSD", "openbsd"}, {"SunOS", "sunos"}, {"AIX", "aix"},
	};
	for (size_t at = 0; at < sizeof names / sizeof names[0]; at++) {
		if (strcmp(sysname, names[at][0]) == 0) {
			return names[at][1];
		}
	}
	return NULL;
}
#endif
