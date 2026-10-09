package release

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// LicenseExpression is cohere's license, as the SPDX expression every package it publishes declares:
// MIT or Apache-2.0, at the user's option, as Rust and Biome are licensed (Kirk, #m5wmxnk).
//
// The OR is uppercase because SPDX requires it. A lowercase one is not an operator, so a registry or a
// license scanner reads the field as one unknown license rather than a choice of two.
const LicenseExpression = "MIT OR Apache-2.0"

// LicenseFileNames are the license texts at the module root, shipped in every package beside the code.
//
// Both, because a dual license is two offers and a recipient may take either, which they can only do with
// each text in hand. Named as Rust names them, which GitHub and license scanners recognize as a pair.
var LicenseFileNames = []string{"LICENSE-APACHE", "LICENSE-MIT"}

// NoticeFileName and ThirdPartyNoticesFileName are the notices internal/release/notices generates at the
// module root: the TypeScript compiler's Apache-2.0 notice, and the license of every other project cohere
// is built on.
const (
	NoticeFileName            = "NOTICE"
	ThirdPartyNoticesFileName = "THIRD_PARTY_NOTICES.md"
)

// LegalFileNames are what every package ships beside its code: both license texts and both notices. The
// notices credit the code each package carries, and the licenses of what it is built on require that.
var LegalFileNames = append(append([]string{}, LicenseFileNames...), NoticeFileName, ThirdPartyNoticesFileName)

// CopyrightHolder is cohere's copyright holder, the legal name Kirk confirmed (#m5wmxnk), and
// CopyrightLine is how LICENSE-MIT and NOTICE state it.
const (
	CopyrightHolder = "System, Inc."
	CopyrightLine   = "Copyright (c) 2026 " + CopyrightHolder
)

// stageLicenses copies the licenses and notices into a package, refusing one that is missing or empty: a
// package published without them carries a license field that points at nothing, or code it doesn't
// credit.
func stageLicenses(moduleDirectory string, packageDirectory string) error {
	for _, name := range LegalFileNames {
		contents, err := os.ReadFile(filepath.Join(moduleDirectory, name))
		if err != nil {
			return fmt.Errorf("reading the license every package ships: %w", err)
		}
		if len(bytes.TrimSpace(contents)) == 0 {
			return fmt.Errorf("%s is empty, so the package would ship a license or notice with nothing in it", name)
		}
		if err := os.WriteFile(filepath.Join(packageDirectory, name), contents, 0o644); err != nil {
			return fmt.Errorf("staging %s: %w", name, err)
		}
	}
	return nil
}
