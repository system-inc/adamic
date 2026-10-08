Choice (b): unconditional standalone profile compilation

Base: origin/area/stage1-lint 2fbfe4215.

Shared profile_test.go:112-168 compiles main.ts and compares a no-var witness. TestProfileArtifacts at lines 17-60 builds the same registry main.ts; it does not load these four standalone profile.a entry points. The local tests cover those entry points and retain sanitized native, release native and emitted JavaScript compilation.

All four tests now resolve their local profile.a and use t.TempDir(); neither WAVE07_PROFILE nor WAVE07_ARTIFACTS is read. Persistent artifact validation scripts now invoke the compiler CLI directly, including mutant entry points. Python syntax checks and git diff --check passed; full standalone validation scripts were not rerun. No rule or shared harness changed.

Validation (both variables explicitly unset):

    env -u WAVE07_PROFILE -u WAVE07_ARTIFACTS go test ./stage1/cohere/lint/rules/no-underscore-dangle ./stage1/cohere/lint/rules/no-unsafe-negation ./stage1/cohere/lint/rules/no-unsafe-optional-chaining ./stage1/cohere/lint/rules/typescript-no-this-alias -run ^TestCompileProfiles$ -p 1 -count=1 -v -timeout 10m

PASS: no-underscore-dangle 15.203s, no-unsafe-negation 13.623s, no-unsafe-optional-chaining 15.638s, typescript-no-this-alias 15.111s. No skips. Raw log: wave1-07-profile-tests-evidence/tests.log.
