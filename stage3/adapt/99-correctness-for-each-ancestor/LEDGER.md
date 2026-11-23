- **99-correctness-for-each-ancestor**: forEachAncestorDirectory drops its internal
  Path overload and uses the string overload/callback contract in a cast-free
  string implementation. The watchUtilities.ts:296 cache callback and
  typingsInstaller.ts:258 installPackage callback require only string; their
  bodies and runtime JavaScript are unchanged. Whole-owner exact text/SHA pins,
  drift/no-write contracts and zero-edit replay guard the adaptation. Evidence:
  typescript/for-each-ancestor-evidence, stage3/census/for-each-ancestor-adapt.
