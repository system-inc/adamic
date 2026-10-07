These packages are the independent Node oracle for stage1/cohere/markdownblocks's
TestMarkdownUnicodeWidths. The test's width_library.mjs imports their index.js
files directly from ADAMIC_MARKDOWNWIDTH_DEPS/node_modules. Without that variable,
it looks in /tmp/adamic-markdown-width; setup deliberately uses its tool directory
instead and exports the variable in env.sh. Source the printed env.sh before tests.

package.json pins emoji-regex 10.6.0, get-east-asian-width 1.6.0 and narrow-emojis
0.0.3. package-lock.json pins the registry tarballs and SHA-512 integrity values.
setup-markdown-width.py runs npm ci with scripts disabled and an empty temporary
npm cache, then publishes a completed installation. npm-bootstrap.json separately
pins npm 11.9.0 and its SHA-512 integrity because fresh setup may have only the Node
executable, without a global npm. The bootstrap is downloaded and verified only
when installation is needed; no system npm or global npm cache is required.

Preparation runs in the Node background task after Node is ready, overlapping Go,
LLVM and submodules. The install directory is outside /root, including the fallback
when ADAMIC_TOOLS is under /root. Each cache hit validates the lockfile, package
manifest, bootstrap manifest, installer source, Node version, and installed file
bytes, paths and permissions. ADAMIC_GATE_UNCACHED=1 always reinstalls with npm ci.
The outer Go warming stamp also includes all three manifests and the installer.

Run the installer proofs and independent cache omissions with:

    ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_markdown_setup.py > /tmp/markdown-install-tests.log 2>&1
    python3 cloud/markdown-setup-mutants.py > /tmp/markdown-install-mutants.log 2>&1

The integration retains its scratch directory and logs, corrupts a package's
expected integrity to require npm's EINTEGRITY failure, and compares cached and
forced-uncached installation bytes. Failed attempts keep the previous installation.
