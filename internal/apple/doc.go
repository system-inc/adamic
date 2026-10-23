// Package apple holds the macOS witnesses for Adamic's Objective-C bridge (docs/apple.md): each program in testdata is
// held to its Objective-C or Swift twin, and the leaf table to the SDK. Its tests run only on darwin. This file has no
// build constraint, so go list names the package on every platform and the compiler census counts the tests'
// compiler use there too (cloud/fast-gate/compiler-dependencies.json declares it).
package apple
