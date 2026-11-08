# Mutants for stage3/api's two provisioning roads (#zvaxp1x)

The delta provisions a checkout's stage3/api from the install laid in the invoking repository when it digests to that checkout's own pin (`stage3/apply.py --packages-digest`), and only otherwise by npm. Each mutant is one edit to prepare.go, run on Cloud (job 01M6YMMCZX8HEGFT4ER8C15N8F, the box's toolchain), then put back; the tree read clean after.

| Mutant | Edit | Red |
|---|---|---|
| A link instead of a copy | provisionCheckout: `copyTree(laid, ...)` to `os.Symlink(laid, ...)` | TestProvisioningCopiesThePinnedLaidInstall, prepare_test.go:215: the checkout's node_modules is not a directory of its own |
| Copy before the digest decides | provisionCheckout: `if unused == "" {` to `if true {` | TestProvisioningWithoutNpmNamesBothRoads, prepare_test.go:270: a mismatched laid install is copied instead of naming both roads |
| The copy unchecked | provisionCheckout: `if copied != pin {` to `if false && copied != pin {` | TestProvisioningVerifiesTheCopy, prepare_test.go:291: a copy without node_modules/.bin's link passes |
