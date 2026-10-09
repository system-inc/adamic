package tsoptions

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/contentmapper"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
)

type ParsedOptions struct {
	CompilerOptions *core.CompilerOptions `json:"compilerOptions"`
	TypeAcquisition *core.TypeAcquisition `json:"typeAcquisition"`

	FileNames         []tspath.RootedFilePath  `json:"fileNames"`
	ProjectReferences []*core.ProjectReference `json:"projectReferences"`
	ContentMappers    []*contentmapper.Mapper  `json:"contentMappers"`

	// SourceExtensions are the extensions, beyond TypeScript's own, whose files this config's program reads
	// as TypeScript source under their own names: the top-level "sourceExtensions" key. A file name keeps
	// its extension everywhere; only enumeration, loading and resolution are told the extension exists.
	SourceExtensions []string `json:"sourceExtensions,omitempty"`
}

// Equals compares parsed configuration values using each option type's equality semantics.
func (p *ParsedOptions) Equals(other *ParsedOptions) bool {
	if p == other {
		return true
	}
	if p == nil || other == nil {
		return false
	}
	if !p.CompilerOptions.Equals(other.CompilerOptions) ||
		!p.TypeAcquisition.Equals(other.TypeAcquisition) {
		return false
	}
	return (p.FileNames == nil) == (other.FileNames == nil) &&
		slices.Equal(p.FileNames, other.FileNames) &&
		(p.ProjectReferences == nil) == (other.ProjectReferences == nil) &&
		slices.EqualFunc(p.ProjectReferences, other.ProjectReferences, func(a, b *core.ProjectReference) bool {
			return a == b || a != nil && b != nil && *a == *b
		}) &&
		(p.ContentMappers == nil) == (other.ContentMappers == nil) &&
		slices.EqualFunc(p.ContentMappers, other.ContentMappers, (*contentmapper.Mapper).Equals) &&
		(p.SourceExtensions == nil) == (other.SourceExtensions == nil) &&
		slices.Equal(p.SourceExtensions, other.SourceExtensions)
}
