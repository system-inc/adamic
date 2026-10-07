# Lint wave 1 slot 04 claim

Branch: codex/lint-wave1-04. Main base: d090af5.

The helper report delegates the ordered list to ../HELPERS.md. Positions 10,
11 and 12 in that list are:

1. consistent-this: claimed for this unit.
2. func-name-matching: claimed for this unit.
3. max-classes-per-file: skipped because origin/codex/stage1-lint-batch4
   already ports it in stage1/cohere/lint/max_classes_per_file.ts;
   implementation commit 7bd94a2f92cc17d12bb69bbd78bfe2c4ac4ca63d.

All origin branches were fetched and searched before claiming. No implementation
of consistent-this or func-name-matching was found in stage1/cohere/lint on
origin branches. This claim is pushed before writing either port.

The requested foundations did not merge cleanly together. Registration merged
cleanly onto current main. Helpers conflicted in README.md, lint.ts,
lint_test.go, main.ts, settings.ts and testdata/oracle.go. The merge preserves
registration's versions of those six files and retains helper additions.

The registration generator currently requires rule.ts, generates rule.ts
imports, rejects .a mutant files, and the test source copier omits .a modules.
New Adamic source in this unit will use .a as instructed. Shared compatibility
changes require expanded territory; this is a prerequisite, not a rule port.

## Continuation claim, October 7

After pushing all existing work and fetching every origin head without recursive
submodule history, this unit reserves the next three names:

1. structure/tailwind-no-physical-direction, helper-ready position 46.
2. @eslint-community/eslint-comments/require-description, first remaining
   syntax-only inventory entry.
3. @next/next/google-font-display, next remaining syntax-only inventory entry.

Selection snapshot: origin/main ef3d907ecdc4c771b016f7d9c52372def057a340,
297 origin refs and 27 unique Markdown blobs beneath claims/. Published
assignments, including previously skipped rules, are treated as occupied so
this continuation does not duplicate any worker's prior selection. None of
these three names appears in those claim documents or in main's executable
lint sources. Inventory order is its rules array filtered to neither type
information nor binding-only requirements. The helper REPORT.md delegates its
ordered list to HELPERS.md. This update is pushed before writing rule code.

Previous claims remain reserved; this continuation does not release them or
claim that their blockers have been repaired. New Adamic source uses .a.

Continuation status: the three rules are reserved but unimplemented. The owned
.a substrate probe and original Go corpus measurements are recorded in
wave1-04-continuation-report.md. Tailwind and Google Fonts witnesses refuse in
the independent parser; require-description's JSX witness silently lacks its JSX
node, though its comment is discovered correctly. No rule certification or
throughput claim is made. The shared .a registration contract is still absent.
