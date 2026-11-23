# Read the old signature map after a definedness check

Family 44, debug.ts:235:16 from builder.ts:921:1 in the scout, for Compiler under the split. Adamic's Nov 22 ruling on #bbssdcy permits the defined-container check. See docs/rulings.md rows 68, 141, 142 and 158 at the pin.

BuilderState declares oldSignatures as an optional Map<Path, string | false>. updateShapeSignature in builderState.ts initializes it with ||= and sets the prior signature or false. The write-file callback in builder.ts initializes it with ??= and sets the prior signature or false when the path has no saved entry. builderState.ts and builder.ts each clear its existing contents through optional .clear(). Reads also appear in state reuse, build info, and the write-file callback. The adapter changes only the read in isChangedSignature.

The original checkDefined calls assertIsDefined before .get(path). A missing value takes Debug.fail, which throws an Error with the message “Debug Failure.” The adapter performs the same nullish test after exactly one read of the property and calls Debug.fail before either map lookup. It preserves both the old map's false-to-undefined conversion and the second check on fileInfos.

The JavaScript diff is javascript.diff in the patch evidence. It covers isChangedSignature.
