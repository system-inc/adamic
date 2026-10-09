#ifndef SOURCEKITD_SHIM_H
#define SOURCEKITD_SHIM_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

/*
 The part of sourcekitd's C API the engine calls, declared by hand because Xcode ships
 sourcekitdInProc.framework without its headers.

 Function pointer types rather than functions, for the reason the libclang shim gives: the engine loads
 the framework with dlopen from the toolchain `xcrun` names, so a binary never carries one Xcode's path.

 The layouts match tools/SourceKit/tools/sourcekitd/include/sourcekitd/sourcekitd.h in swiftlang/swift,
 the header SourceKit-LSP and SwiftLint compile against. A variant is three 64-bit words passed by value,
 which is why it needs a C declaration at all: Swift cannot pass its own struct through a C function
 pointer.
 */

typedef void *CohereSourcekitdUid;
typedef void *CohereSourcekitdObject;
typedef void *CohereSourcekitdResponse;

typedef struct {
    uint64_t data[3];
} CohereSourcekitdVariant;

typedef void (*CohereSourcekitdInitialize)(void);
typedef CohereSourcekitdUid (*CohereSourcekitdUidGetFromCString)(const char *string);
typedef CohereSourcekitdObject (*CohereSourcekitdRequestDictionaryCreate)(const CohereSourcekitdUid *keys, const CohereSourcekitdObject *values, size_t count);
typedef void (*CohereSourcekitdRequestDictionarySetString)(CohereSourcekitdObject dictionary, CohereSourcekitdUid key, const char *string);
typedef void (*CohereSourcekitdRequestDictionarySetUid)(CohereSourcekitdObject dictionary, CohereSourcekitdUid key, CohereSourcekitdUid uid);
typedef void (*CohereSourcekitdRequestDictionarySetInt64)(CohereSourcekitdObject dictionary, CohereSourcekitdUid key, int64_t value);
typedef void (*CohereSourcekitdRequestDictionarySetValue)(CohereSourcekitdObject dictionary, CohereSourcekitdUid key, CohereSourcekitdObject value);
typedef CohereSourcekitdObject (*CohereSourcekitdRequestArrayCreate)(const CohereSourcekitdObject *objects, size_t count);
typedef void (*CohereSourcekitdRequestArraySetString)(CohereSourcekitdObject array, size_t index, const char *string);
typedef void (*CohereSourcekitdRequestRelease)(CohereSourcekitdObject object);
typedef CohereSourcekitdResponse (*CohereSourcekitdSendRequestSync)(CohereSourcekitdObject request);
typedef bool (*CohereSourcekitdResponseIsError)(CohereSourcekitdResponse response);
typedef const char *(*CohereSourcekitdResponseErrorGetDescription)(CohereSourcekitdResponse response);
typedef CohereSourcekitdVariant (*CohereSourcekitdResponseGetValue)(CohereSourcekitdResponse response);
typedef char *(*CohereSourcekitdVariantJsonDescriptionCopy)(CohereSourcekitdVariant variant);
typedef void (*CohereSourcekitdResponseDispose)(CohereSourcekitdResponse response);

#endif
