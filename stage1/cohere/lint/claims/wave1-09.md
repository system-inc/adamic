# Lint wave 1, slot 09

Branch: `codex/lint-wave1-09`.

Assigned helper-ready positions, 1-based:

25. `react/jsx-no-useless-fragment`
26. `react/no-invalid-html-attribute`
27. `react/no-unescaped-entities`

`helpers/REPORT.md` links the ordered handoff in `HELPERS.md`; the same order is
recorded in `helpers/readiness.json` on `origin/codex/lint-helpers`.
No implementation candidate was found on main or any fetched origin branch by
searching stage1 lint source and registration files outside inventory/helper data.
None is skipped as already ported.

Claim published before any rule code. Implementation is blocked pending proof of
the shared parser's JSX support and resolution of foundation merge conflicts.
The helper merge was attempted and aborted after six shared-file conflicts;
registration is merged. No shared dispatch was replaced to hide those conflicts.

## Next three, October 7

The existing claim and evidence were pushed before fetching all origin heads.
At origin/main ef3d907e, the following are neither ported on main nor named by
any claim document under stage1/cohere/lint/claims on any fetched origin branch:

1. `structure/tailwind-no-physical-direction`: final unclaimed entry in the
   46-rule helper handoff linked by helpers/REPORT.md.
2. `@eslint-community/eslint-comments/require-description`: first available
   syntax-only inventory row after helper exhaustion.
3. `@next/next/google-font-display`: next available syntax-only inventory row.

Inventory source: origin/codex/lint-inventory, 73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf.
Syntax-only means needs_type_information=false and binding_only=false, retaining
inventory order and including rules whose helper dependencies remain explicit.
All Markdown documents under claims were inspected across every origin ref;
rule-name matches exclude an entry. Main source/registration files were inspected
outside inventory/helper/testdata artifacts. These new claims precede rule code.
