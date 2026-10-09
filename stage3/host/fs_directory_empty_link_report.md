Fixed the symlink fixture by unlinking empty after its empty-target attempt, with an empty catch, so Darwin leaves no dangling empty link.
Parent: e32a414a630de25a372b0855a09f73d7f8153704; pushed SHA is supplied in the handoff.
Checks on Linux: uncached dual-backend symlink oracle PASS 2.434s; complete Linux counts regeneration PASS 49.992s; edited-source whitespace check passed.
Mutants: wrong target, code and message caught only by Node stdout; all exited zero with clean ASan, UBSan and leak checks.
Uncovered: macOS execution was not available; Darwin empty-target behavior is the user's official Node 24.19.0 observation.

The first counts check correctly failed because the extra caught ENOENT cleanup on Linux adds five allocations/frees and two releases. Full Linux regeneration changed exactly the symlink fixture row: 162/162/90/177/34/0 to 167/167/90/179/34/0. An explicit comparison verified every other row, including process_observations.a and every other node_fs_directory fixture, stayed byte-for-byte unchanged. No Mac count record was used.
