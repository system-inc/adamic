These three native listeners receive a cached `Node`; their relevance dispatch
uses pinned numeric SyntaxKind values. Each rule.json declares both kind names
and numeric values. The isolated driver dispatches only the declared listeners.

Run `python3 validate.py` after sourcing the toolchain and building the stage0
compiler at /workspace/wave-22-sixth-adamic and the checker archive at
/workspace/wave-22-sixth-checker.a. It writes every subprocess output directly
to logs under /workspace/wave-22-sixth-work. No shared generator or harness is
changed. See WAVE_22_SIXTH_REPORT.md for pinned dependencies and commands.

The shared JSX parser is not on main yet. Validation copies the published
parser/scanner from a8a62d62ca49db7415e14c3887dd305022b17309 into an isolated
source tree; it never changes repository parser files or injects Go ASTs.
The independent Go oracle calls production rule.Run, and canonicalizes complete
findings including fixes and suggestions. Shared harness registration and the
future Diagnostic model remain integration work.
