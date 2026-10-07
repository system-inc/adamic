# Wave 01 listener manifests

Each child directory contains rule.json with numeric typescript-go SyntaxKind
values in kinds. The directory basename matches the existing native file:
no_implicit_return, no_deprecated and no_else_return are at the typeaware root;
child_process_error_listener, response_status_check and independent_await_in_loop
are in wave_01_next. These manifests describe the six completed rules only.
The shared driver does not yet load this directory on this branch.

listener_check.py checks manifests and native listenerKinds against production
Go listener registrations and independently measured enum values. It also
rejects an in-memory +1 kind mutation of each manifest.

These declarations prepare integration; existing string-kind helpers and
per-rule scans still need migration to shared numeric node delivery. No new
speed or full compliance claim is made here.
