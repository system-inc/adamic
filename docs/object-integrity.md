# Object descriptors and integrity

Descriptor operations require a proven complete plain data shape. Snapshots expose value and attribute observations, remain in tracked readonly const bindings, and enumerate only proven keys. Structural calls, returns and aggregate escapes are refused because they erase the boxed storage proof. Definitions update existing number, boolean or string fields without changing their representations. Accessors, new fields and individual attribute changes are refused; omitted attributes or literal true preserve existing attributes.

Seal, freeze and preventExtensions preserve receiver identity and expose the corresponding integrity queries. Freezing is shallow. Map and Set integrity applies to own properties; their internal entries remain mutable. Frozen property writes and Object.assign failures throw catchable TypeError objects after evaluating the right-hand side.

The five witnesses in internal/oracle/testdata/object_integrity cover descriptors, definitions, snapshot shape, plain-object integrity and Map/Set integrity. Their native and wasm32 behavior is compared with Node. Date shapes, RegExp integrity and captured-stack enumerability witnesses are deferred to the owning library pieces.
