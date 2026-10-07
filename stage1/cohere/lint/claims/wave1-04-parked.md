# Parked Tailwind ports

Preserved implementation and evidence: commit 2055f6f3560b4afee1f059bc3702b0fb3c96ae0f.
These rules are removed from the landing module graph and descriptors:

- better-tailwindcss/enforce-consistent-important-position
- better-tailwindcss/enforce-consistent-variable-syntax
- better-tailwindcss/enforce-consistent-variant-order

Reproduce at that commit: source /workspace/adamic-tools/env.sh;
go run ./cmd/lint-registry; go run ./cmd/adamic js stage1/cohere/lint/main.ts.
The lowerer refuses rules/tailwind-important-position/surface.a:41:132:
RegExp with a nonconstant pattern. Source contains new RegExp(pattern, 'u').
This shared Surface is imported by all three ports. No manual regex fallback.
Variant order additionally needs actual Tailwind program/design-system inputs:
shared TestOwnedWitnesses reports that its witness has no Go findings.
The original HARNESS_LANDING.md and preserved commit contain full logs.
The two unblocked ports, consistent-this and func-name-matching, remain.
