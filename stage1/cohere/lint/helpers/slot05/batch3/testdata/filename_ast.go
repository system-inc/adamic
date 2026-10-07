// Oracle-only constructor: normal parser construction forbids unnormalized filenames.
// This sets the field read by FileName without changing NormalizedFileName.
package ast

func AdamicSourceFileName(name string) *SourceFile {
	return &SourceFile{parseOptions: SourceFileParseOptions{FileName: name}}
}
