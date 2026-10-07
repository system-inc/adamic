// Package tsgo holds the public C ABI of the external checker library.
package tsgo

import _ "embed"

//go:embed tsgo.h
var Header []byte
