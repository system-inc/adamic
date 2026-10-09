#ifndef INDEX_STORE_SHIM_H
#define INDEX_STORE_SHIM_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

/*
 The part of the index store's C API the engine reads, declared by hand because Xcode ships
 libIndexStore.dylib without indexstore.h.

 Function pointer types, loaded with dlopen from the toolchain `xcrun` names, for the reason the libclang
 shim gives. The layouts match clang/include/indexstore/indexstore.h in swiftlang/llvm-project, the header
 IndexStoreDB and SourceKit-LSP compile against. A string is a pointer and a length, not NUL-terminated,
 and it is borrowed: valid only until the reader it came from is disposed.
 */

typedef struct {
    const char *data;
    size_t length;
} CohereIndexString;

typedef void *CohereIndexStore;
typedef void *CohereIndexError;
typedef void *CohereIndexUnitReader;
typedef void *CohereIndexUnitDependency;
typedef void *CohereIndexRecordReader;
typedef void *CohereIndexOccurrence;
typedef void *CohereIndexSymbol;
typedef void *CohereIndexSymbolRelation;

typedef CohereIndexStore (*CohereIndexStoreCreate)(const char *path, CohereIndexError *error);
typedef void (*CohereIndexStoreDispose)(CohereIndexStore store);
typedef bool (*CohereIndexStoreUnitsApply)(CohereIndexStore store, unsigned sorted, void *context, bool (*applier)(void *context, CohereIndexString unitName));
typedef CohereIndexUnitReader (*CohereIndexUnitReaderCreate)(CohereIndexStore store, const char *unitName, CohereIndexError *error);
typedef void (*CohereIndexUnitReaderDispose)(CohereIndexUnitReader reader);
typedef CohereIndexString (*CohereIndexUnitReaderGetMainFile)(CohereIndexUnitReader reader);
typedef CohereIndexString (*CohereIndexUnitReaderGetModuleName)(CohereIndexUnitReader reader);
typedef CohereIndexString (*CohereIndexUnitReaderGetOutputFile)(CohereIndexUnitReader reader);
typedef bool (*CohereIndexUnitReaderIsSystemUnit)(CohereIndexUnitReader reader);
typedef bool (*CohereIndexUnitReaderDependenciesApply)(CohereIndexUnitReader reader, void *context, bool (*applier)(void *context, CohereIndexUnitDependency dependency));
typedef int (*CohereIndexUnitDependencyGetKind)(CohereIndexUnitDependency dependency);
typedef CohereIndexString (*CohereIndexUnitDependencyGetName)(CohereIndexUnitDependency dependency);
typedef CohereIndexString (*CohereIndexUnitDependencyGetFilePath)(CohereIndexUnitDependency dependency);
typedef CohereIndexString (*CohereIndexUnitDependencyGetModuleName)(CohereIndexUnitDependency dependency);
typedef CohereIndexRecordReader (*CohereIndexRecordReaderCreate)(CohereIndexStore store, const char *recordName, CohereIndexError *error);
typedef void (*CohereIndexRecordReaderDispose)(CohereIndexRecordReader reader);
typedef bool (*CohereIndexRecordReaderOccurrencesApply)(CohereIndexRecordReader reader, void *context, bool (*applier)(void *context, CohereIndexOccurrence occurrence));
typedef CohereIndexSymbol (*CohereIndexOccurrenceGetSymbol)(CohereIndexOccurrence occurrence);
typedef uint64_t (*CohereIndexOccurrenceGetRoles)(CohereIndexOccurrence occurrence);
typedef void (*CohereIndexOccurrenceGetLineColumn)(CohereIndexOccurrence occurrence, unsigned *line, unsigned *column);
typedef bool (*CohereIndexOccurrenceRelationsApply)(CohereIndexOccurrence occurrence, void *context, bool (*applier)(void *context, CohereIndexSymbolRelation relation));
typedef uint64_t (*CohereIndexSymbolRelationGetRoles)(CohereIndexSymbolRelation relation);
typedef CohereIndexSymbol (*CohereIndexSymbolRelationGetSymbol)(CohereIndexSymbolRelation relation);
typedef CohereIndexString (*CohereIndexSymbolGetString)(CohereIndexSymbol symbol);
typedef int (*CohereIndexSymbolGetKind)(CohereIndexSymbol symbol);

#endif
