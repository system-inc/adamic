// host_platform.h: the name Node gives process.platform. Node's is the platform it was built for, so
// here it's the target this runtime is compiled for, read from the target's predefined macros in
// Node's order (Runtime's shape, Platforms' ruling on #t90fcv5), never the build host: a Mac
// building for Linux says linux. A target Node may name otherwise is refused when the runtime is
// built. wasm32-wasi defines no name: a module can't see its host's platform, so the read panics
// there (node_host.c). Macros only, so a test preprocesses it alone (TestHostPlatformNames).
#ifndef ADAMIC_HOST_PLATFORM_H
#define ADAMIC_HOST_PLATFORM_H
#if defined(__wasi__)
#elif defined(__ANDROID__)
#define ADAMIC_HOST_PLATFORM "android"
#elif defined(__linux__)
#define ADAMIC_HOST_PLATFORM "linux"
#elif defined(__APPLE__)
#define ADAMIC_HOST_PLATFORM "darwin"
#elif defined(__FreeBSD__)
#define ADAMIC_HOST_PLATFORM "freebsd"
#elif defined(__OpenBSD__)
#define ADAMIC_HOST_PLATFORM "openbsd"
#elif defined(__sun)
#define ADAMIC_HOST_PLATFORM "sunos"
#elif defined(_AIX)
#define ADAMIC_HOST_PLATFORM "aix"
#else
#error "process.platform: a target host_platform.h doesn't name; add it in Node's order"
#endif
#endif
