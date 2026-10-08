package ast

import nativeast "github.com/microsoft/TypeScript/tsc/internal/ast"

func AdamicSourceFileName(name string) *SourceFile {
	return nativeast.AdamicSourceFileName(name)
}
