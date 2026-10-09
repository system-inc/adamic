#ifndef CLANG_DIAGNOSTICS_SHIM_H
#define CLANG_DIAGNOSTICS_SHIM_H

/*
 The handful of libclang's serialized-diagnostics API the types phase needs, declared by hand because Xcode
 ships libclang.dylib without the clang-c headers.

 Declared as function pointer types rather than functions, because the engine loads libclang with dlopen
 from whatever toolchain `xcrun` names instead of linking a hardcoded Xcode path: a binary linked against
 one Xcode's path stops starting when Xcode moves or updates, and fails with a loader error that names
 neither cohere nor the reason.

 The layouts match clang-c/CXString.h and clang-c/CXDiagnostic.h, which have been stable since libclang's
 first release.
 */

typedef struct {
    const void *data;
    unsigned privateFlags;
} CohereClangString;

typedef void *CohereClangDiagnosticSet;
typedef void *CohereClangDiagnostic;
typedef void *CohereClangFile;

typedef struct {
    const void *pointerData[2];
    unsigned integerData;
} CohereClangSourceLocation;

typedef CohereClangDiagnosticSet (*CohereClangLoadDiagnostics)(const char *file, int *error, CohereClangString *errorString);
typedef unsigned (*CohereClangGetNumDiagnosticsInSet)(CohereClangDiagnosticSet set);
typedef CohereClangDiagnostic (*CohereClangGetDiagnosticInSet)(CohereClangDiagnosticSet set, unsigned index);
typedef void (*CohereClangDisposeDiagnosticSet)(CohereClangDiagnosticSet set);
typedef int (*CohereClangGetDiagnosticSeverity)(CohereClangDiagnostic diagnostic);
typedef CohereClangString (*CohereClangGetDiagnosticSpelling)(CohereClangDiagnostic diagnostic);
typedef CohereClangString (*CohereClangGetDiagnosticCategoryText)(CohereClangDiagnostic diagnostic);
typedef CohereClangSourceLocation (*CohereClangGetDiagnosticLocation)(CohereClangDiagnostic diagnostic);
typedef void (*CohereClangGetFileLocation)(CohereClangSourceLocation location, CohereClangFile *file, unsigned *line, unsigned *column, unsigned *offset);
typedef CohereClangString (*CohereClangGetFileName)(CohereClangFile file);
typedef const char *(*CohereClangGetCString)(CohereClangString string);
typedef void (*CohereClangDisposeString)(CohereClangString string);

#endif
