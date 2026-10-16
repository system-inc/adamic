package buildcache

// Inputs are everything that can change a product. Name is one per build function. Files are repository-relative
// files and directories, valued by content (a file by its git blob id and exec bit, a directory by every name and file
// under it, sorted). Flags are every flag and environment value the build reads. Toolchain names the tools, from Tool
// or runtime.Version().
type Inputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain []string
}
