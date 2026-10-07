# Owned rule port

The .a listener, exact messages, upstream Go adapter, raw witness and semantic
mutant are registered by rule.json. Full independent validation and limitations
are recorded in ../typescript-no-unnecessary-type-constraint/REPORT.md.

The new repair-heavy rules retain complete typed records in their Rule.records.
Their registered finish hook explicitly refuses unsupported serialization rather
than expose findings with truncated fixes or suggestions. The owned validation
driver consumes those complete records directly. Shared registration and harness
files are untouched. The announced codex/lint-harness-dot-a branch was absent
from the fetched origin heads when work began.

The original unused-expression rule's seven captured JSX profiles and Tailwind's
prior JSX profile remain shared-parser gaps. Every excluded JSX profile has an
independently successful Go verdict and an explicit parser refusal on source
Node, emitted JavaScript and sanitized native in the owned validation logs.
