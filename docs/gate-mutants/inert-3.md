# Inert landing 3 of 10

Gate mutant 5 (#fyvmsy8): one of ten docs-only commits on main ce1c5a2f. A canary gated against this chain's tip~10 selects no test, so a correct gate reads it void (under the canary's 500 passed tests), never green. Never merge.
