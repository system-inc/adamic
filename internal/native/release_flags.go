package native

// RuntimeFlagsKey uses the runtime library's actual key function to fingerprint flags. Fixing the
// other inputs lets the release guard compare configurations without compiling a runtime archive.
func RuntimeFlagsKey(flags []string) string {
	return runtimeKey(nil, flags, "", "")
}
