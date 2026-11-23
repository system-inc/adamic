# Mutant for the bind receiver's words (#d6jncyx)

## M1: the old words

In `internal/lower/function_bind.go`, the stop under `!l.stableBindReceiver(receiver)` says "method bind with an assigned receiver identifier" again.

Expected: `TestBindReceiverWordsNameWhatTheReceiverIs` goes red on both rows, and internal/oracle's `TestMethodBindRefusals/assigned_receiver` goes red too.
