# Step 1 measurement in progress

Base: efe887763, fetched origin/main tip. Pinned TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8.

The complete dumpTypes owner has five later reads of objectFlags: Anonymous in the display condition, Reference for instantiatedType/typeArguments/referenceLocation, ReverseMapped for reverseMappedSourceType/reverseMappedMappedType/reverseMappedConstraintType, EvolvingArray for evolvingArrayElementType/evolvingArrayFinalType, and Tuple for isTuple. All five reads apply bit masks. No raw objectFlags value is serialized. descriptor.intrinsicName has no later property access: JSON.stringify(descriptor) serializes it, omitting undefined and retaining strings.

Three oracle projects were run by Node tsc --generateTrace. DifferentNamesNotSpecified visited 30559 types, DefaultExcludeOnlyNodeModules visited 27038, and emitDecoratorMetadataCommonJSIsolatedModule visited 27045. The first two exited 0. The third exited 2 for missing angular2/core and an implicit-any return, but wrote its trace; its diagnostics are retained.

Total visits: 84642. Own objectFlags: 50995. Own intrinsicName: 88. getObjectFlags raw differences: 33693. Real nonzero objectFlags on non-ObjectFlagsType types: 46 (33 hold 52953088; 13 hold 14680064). All five mask differences: 0. Intrinsic flag guard value differences: 0. Property-existence read value differences: 0.

With both runtime replacements applied only inside the private Node bundle's dumpTypes, all three written types.json files are byte-identical. Byte sizes: 7153496, 6227528, 6229391. Child exits are unchanged (0, 0, 2). Individual records, file hashes and retained complete outputs accompany this note.

Whole lane and stage-0 acceptance remain pending. No source adaptation has been added and no claim of a completed road is made yet.
