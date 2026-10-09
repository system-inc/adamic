The fresh literal initially has no force field. A structural assertion plants
force='yes' before the host call. The static type at rmSync still lacks force.
This write/escape cannot prove absence, regardless of the const binding.
Node v24.19.0 reads the planted runtime field and prints:

    The "options.force" property must be of type boolean. Received type string ('yes')

Defaulting force to false would instead throw ENOENT. The proof mutant skips
the binding-use analysis, accepts this literal-backed const and fails the direct
absence assertion. The compiler's independent cast rules may refuse earlier;
they must not mask an unsound classification in the host proof itself.
The separate hidden-field-in-initializer assertion is also refused and tested.
