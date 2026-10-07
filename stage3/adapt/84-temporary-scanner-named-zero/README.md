Temporary: comes out when zero flag assignment inference lands

Plan published before implementation. Slice scanner.ts only: replace the lone
tokenFlags = 0 assignment in reScanInvalidIdentifier with tokenFlags = TokenFlags.None.
Both values are exactly zero; the named member carries the closed flag proof.
This is the next observed adamic/enum-flags refusal after adaptation 83.
Validate Node equality and all upstream baselines.

Validated with 85-86: full upstream baseline 106,367 passing, zero failures/
pending/differences (252.427 seconds). Node matches all tokens. Restoring raw
zero independently is caught by enum-flags at scanner:1864:22. The next syntax
refusal after this reset is a numeric code passed to a CharacterCodes parameter.

Current profile: superseded by open numeric enums ec67b02. The newest scratch
compiler reaches lowering without this workaround. adapt-slice.sh omits it;
the script and earlier validation remain reproducible historical evidence.
