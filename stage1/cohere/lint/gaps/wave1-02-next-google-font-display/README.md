`@next/next/google-font-display` is claimed by wave1 slot 02 but not implemented or registered. Its Go cohere tests use JSX opening/self-closing elements and decoded JSX attributes. The shared Adamic parser does not construct those nodes.

`jsx-probe.a` uses the first real failing Go fixture. Source Node, emitted JavaScript and ASan/UBSan native all exit 70 with exactly:

```
adamic: panic: parser slice expected GreaterThanToken, got Identifier at 32 in Component.tsx
```

The 16 Go fixture rows pass in cohere. The isolated `TestWave02JSXBlocker` proves the shared prerequisite fails consistently; it does not claim rule parity. No placeholder returning zero findings, descriptor, throughput claim or semantic rule mutant is provided. JSX parsing/attribute decoding belongs to the shared parser integrator under `docs/parallel-work.md`, which was read on `origin/codex/no-shared-lists`.

The owned validation overlay in `rules/structure-tailwind-no-physical-direction` reproduces this probe. The claim remains reserved and explicitly blocked.
