# runtime-modules

An unsuffixed exact source-file request returns version 1, `runtime-modules`, program file count; for each source: canonical file name, declaration-file flag, computed runtime-load flag, edge count, resolved target file names. Runtime edges include non-type static imports and reexports, external import-equals, and literal dynamic import/require calls. Declaration files contribute no runtime edges. Self edges and unresolved/outside-program targets are excluded. Duplicate edges retain source scan order. Computed-load detection uses the production Go rule's literal `import(`/`require(` text precondition.

This exports module-resolution facts only. The native decoder builds imported sets and decides which files can reach the StandardStreams module, using its filename suffixes and transitive closure. Native code also decides entry status, block ordering and findings. Suffixes and requests on non-source nodes are refused.
