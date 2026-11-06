# Stopped at check 1

Command, run at repository root before creating this evidence directory:

```sh
python3 stage3/apply.py /workspace/scratch/maybebind-before --build-product > /tmp/maybebind-check1-before.log 2>&1
```

Exit: 1. The complete log is check1-before.log. The final exception is:

```text
RuntimeError: set ADAMIC_TYPESCRIPT_SOURCE to the gate inputs' TypeScript v6.0.3 checkout at 050880ce59e30b356b686bd3144efe24f875ebc8
```

The failure occurred in build_product -> typescript_source, before adapting any
source. There is no before or after tree. No adapter was implemented.

| Check | Result |
| --- | --- |
| 1 baseline full apply | Failed: required gate source is not configured |
| 1 after apply and ordered replay | Not run: stopped at baseline |
| 2 Adamic types, 79 roots | Not run |
| 3 emitted JavaScript diff | Not run; no diff hunks exist |
| 4 every upstream oracle shard | Not run |
| 5 restored-site negative witness | Not run |
| 6 latent and ordinary census | Not run |

CW051's six unit identities were not loaded. No before or after stop was
measured for any unit; a unit table would invent results. The requested
before/after table is therefore unavailable.

No long-lived process was started. All invoked commands exited. No other
adaptation directory was edited. No Go file was added.
