# Mutant for the locale Date form's refusal (#32e763w)

## M1: the locale form NotYet again

In `internal/lower/library_node_fs_file.go`, `localeDateForm` answers false for every name, so a locale or local-timezone Date form stops as the unbuilt NotYet again.

Expected: `TestDateLocaleFormIsRefused` goes red on its first method (toLocaleTimeString), and `TestDateISOStringOtherMethodStopsByName` stays green, since toUTCString is no locale form.
