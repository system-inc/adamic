# Owned rule port

The .a listener, exact messages, upstream Go adapter, raw witness and semantic
mutant are registered by rule.json. Full independent validation and limitations
were recorded on the pre-dedup branch c74d385b6. Current landing evidence is
recorded in ../no-async-promise-executor/LANDING.md.

The record helper now lives in typescript-no-unused-expressions/records.a and
publishes through the landed reportRange API, preserving complete suggestions.
Shared registration and harness files are untouched.

The original unused-expression rule's seven captured JSX profiles remain shared-parser gaps. Every excluded JSX profile has an
independently successful Go verdict and an explicit parser refusal on source
Node, emitted JavaScript and sanitized native in the owned validation logs.
