//go:build never

// A test file under a tag no build sets.
package main

import "testing"

func TestNever(t *testing.T) {}
