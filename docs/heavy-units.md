# Heavy units

A heavy unit is a test that can't fit the gate's 30 s unit budget and has a ruling to leave the gate unit without
leaving coverage (#m9es0rv, @system_adamic, Oct 9 12:59Z, on #xfx7jc4's _6484). It is declared, never inferred:

- The census class `heavy` names the task that carries the unit, so a heavy skip in a gate unit reads as covered
  elsewhere, by name, and a heavy unit deleted from the declaration makes the census flag its skip as missing coverage.
- The budget check exempts declared heavy units only, never gate units.
- Main's whole gate, or Loom's 30-minute main canary, runs every declared heavy unit with its own declared budget (no
  90 s kill) and pages the unit's owner on red.

First consumers: cohere's #5ggwf8c (lint's TestCompilerAndStage1Agree checker.ts/all sanitized run) and typeaware's
volume corpus (ADAMIC_VOLUME_COMPILER_MANIFEST and ADAMIC_VOLUME_REPOSITORY_MANIFEST, which no gate sets today).

This branch is where #m9es0rv lands; until it does, census rows that wait on heavy units await this branch.
