// Package base holds the rules that check Base's own conventions.
//
// Base is the backend framework `api-phi-health` is built on, at
// `libraries/base/source/` in that repository. These rules are named for it the way
// `react` and `typescript` are named for the plugins they came from, except that
// nothing upstream ships them: Base is ours, and so is every rule here.
//
// # `Verify` in these names is Base's word, not this program's
//
// Two rules read `verify-array-parity` and `verify-optional-parity`, and several
// tables in them hold names like `VerifyIsEmail` and `VerifyArrayUnique`. None of
// that refers to the program this used to be called. It refers to Base's validation
// decorators, at
// `libraries/base/source/foundation/validation/decorators/`, which is 32 classes
// used across 27 files:
//
//	class FormSubmission {
//	    @VerifyIsEmail()
//	    emailAddress: string;
//	}
//
// The decorator marks a property for runtime validation. `VerifyIsEmail` will still
// be spelled that way after this program is called cohere everywhere else, because it
// is Base's public API and renaming it would break every call site.
//
// This is written down because the question is a good one and was asked: during the
// rename from verify to cohere, these names looked like residue of the old program
// and a sweep nearly took them. It would have been silent. The rules key on those
// strings, so a renamed table matches nothing, the rules report nothing, and the
// tree goes green having checked less than it appears to.
package base
