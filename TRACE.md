# Checked write trace for #rgrtt32

Checkpoint. The requested compiler revisions are available as `49ecfe75` and `fcc0f8da`. The latter is reachable through `compiler/view-field-unsupported-4`. The two detached worktrees use the same TypeScript v6.0.3 source pin `050880ce59e30b356b686bd3144efe24f875ebc8`.

The static return matching the reported diagnostic is in `(*lowering).recordSourceSlotWrite` at `internal/lower/source_slot_certificate.go:114` on `fcc0f8da`. `judgeViewAccess` calls it from the view registration walk. A store-only probe could miss this return, so the diagnostic patch instruments both `setProperty` and the exact `NotYet` constructor. It queries receiver reach on a disposable snapshot after the selected census unit has lowered.

The two runtime traces are pending. The diagnostic injector is [trace-overlay.py](trace-overlay.py). It reads the compiler source and writes replacements into the generated entry overlay.
