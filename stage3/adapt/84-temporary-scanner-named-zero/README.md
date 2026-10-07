Temporary: comes out when zero flag assignment inference lands

Plan published before implementation. Slice scanner.ts only: replace the lone
tokenFlags = 0 assignment in reScanInvalidIdentifier with tokenFlags = TokenFlags.None.
Both values are exactly zero; the named member carries the closed flag proof.
This is the next observed adamic/enum-flags refusal after adaptation 83.
Validate Node equality and all upstream baselines.
