Built: corrected three rule.json manifests to Go AST kind names; native rule implementations unchanged.
Commits: enclosing owned-branch commit updates named metadata; previous implementation 3ffd45261 and landing evidence 64a3bffdf.
Commands and outputs: check_metadata.py PASS three Go listener/driver comparisons; fetched 580 origin refs, 33 claim blobs, 170 claimed ranked names.
Mutants: invalid kind-name substitutions rejected for all three declarations; earlier verdict mutants, sanitizer and released-handle results retained.
Not covered: no shared-registry integration or new full oracle run; two remaining unclaimed React rules await native JSX support.

The user corrected rule.json `kinds` from numeric values to typescript-go ast.Kind names without the Kind prefix, as validated by the shared registry. Inspection of origin/lint-rules/harness's registry.go and eqeqeq/rule.json confirms that spelling. The three owned manifests now match the unchanged production Go listener names and the private kind-indexed dispatch registrations. The metadata check rejects an unknown-name mutant per rule. No native .a implementation, shared harness, registry generator, bridge or compiler file changes in this correction.

Current main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 and is an ancestor of this branch. The completed nine ports retain their current-main full oracle, corpus, sanitizer, verdict-mutant and released-handle validation in REPORT.md; metadata-only changes are checked directly rather than rerunning unchanged implementations. Previous numeric manifest evidence is historical and is superseded by evidence/named-metadata.log. Shared-registry Discover integration is not asserted: these rules retain their private type-aware driver.

The fresh claim scan leaves react/static-property-placement and react/style-prop-object unclaimed. There are no unclaimed non-React checker-dependent rules. Native parser.ts remains bc0ee72ab6fa5cdf6b2dcca1e096d9c7f50fae8d and the retained JSX refusal still applies. No new claims are taken. The three React hook reservations remain parked with the blockers already recorded.
