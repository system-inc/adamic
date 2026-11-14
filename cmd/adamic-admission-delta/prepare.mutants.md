# Mutants for stage3/api's two provisioning roads (#zvaxp1x)

The delta provisions a checkout's stage3/api from the install laid in the invoking repository when it digests to that checkout's own pin (`stage3/apply.py --packages-digest`), and only otherwise by npm. Each mutant is one edit to prepare.go, run on Cloud (job 01M6YMMCZX8HEGFT4ER8C15N8F, the box's toolchain), then put back; the tree read clean after.

| Mutant | Edit | Red |
|---|---|---|
| A link instead of a copy | provisionCheckout: `copyTree(laid, ...)` to `os.Symlink(laid, ...)` | TestProvisioningCopiesThePinnedLaidInstall, prepare_test.go:215: the checkout's node_modules is not a directory of its own |
| Copy before the digest decides | provisionCheckout: `if unused == "" {` to `if true {` | TestProvisioningWithoutNpmNamesBothRoads, prepare_test.go:270: a mismatched laid install is copied instead of naming both roads |
| The copy unchecked | provisionCheckout: `if copied != pin {` to `if false && copied != pin {` | TestProvisioningVerifiesTheCopy, prepare_test.go:291: a copy without node_modules/.bin's link passes |

## Through Loom's door (#t7b148j)

The door lays stage3/api's install as read-only hard links, and a door worktree's laid submodules hold no repository of their own. Each mutant was one edit, run in door job J-46kexg7 at d3b54d4c (Lunatic, Nov 14, 2026), then put back, and both pins read green again.

| Mutant | Edit | Red |
|---|---|---|
| The copy keeps the lay's modes | prepare.go, copyInstall: `info.Mode().Perm()\|0200` to `info.Mode().Perm()` | TestProvisioningCorrectsACopyOfAReadOnlyLay, prepare_test.go:294: correcting @types/node fails, "open .../tty.d.ts: permission denied" |
| No mirror holds the pin | build.go, checkoutPinnedFrom: `range mirrors` to `range []string{os.TempDir() + "/adamic-no-mirrors-here"}` | TestPinnedSubmoduleComesFromTheBoxsMirror, command_test.go:384: "cohere at <sha> is in no repository this box holds (looked in ... (not a repository of its own), .../cohere.git (absent))" |
| A mirror that can't check the pin out stops the search | build.go, checkoutPinnedFrom: `looked = append(looked, mirror+" (holds it, but didn't check it out: "+...+")")` then `continue`, made `return nil, err` | TestPinnedSubmoduleSkipsAMirrorThatCannotCheckItOut: the first mirror's lost tree fails the checkout, "reference is not a tree", as Home's shallow TypeScript mirror did (J-dmtb95y); to be run on the door |
