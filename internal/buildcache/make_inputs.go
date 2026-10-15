package buildcache

// Inputs are everything that can change a product. Name is one per build function. Files are repository-relative
// files and directories, valued by content (a file by its git blob id and exec bit, a directory by every name and file
// under it, sorted). Flags are every flag and environment value the build reads. Toolchain names the tools, from Tool
// or runtime.Version(). Go is set for a product go builds from packages none of which hold cgo: what go reads of the
// tree and its release, as "<kind> <path>" lines a read set names, which stand under a trace for every read of go's own
// processes, since what they read depends on what go's build cache held (#mmpt615).
type Inputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain []string
	Go        []string `json:",omitempty"`
}
