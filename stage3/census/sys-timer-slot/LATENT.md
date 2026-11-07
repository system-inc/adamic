Full latent site comparison

| Tree | Checker diagnostics | Units | Distinct NotYet/Refused sites | Common with main | New | Not observed from main |
| --- | --- | --- | --- | --- | --- | --- |
| main | 0 | 140 | 123 | N/A | 0 | 0 |
| A | 2 | 140 | 125 | 122 | 3 | 1 |
| B | 5 | 140 | 102 | 102 | 0 | 21 |

Deduplication key is (location, kind, exact reason, phase), matching the cited core report. Boundary events are not compiler stops. Raw streams retain every event and unit. Main has zero checker diagnostics despite the preserved legacy measurement label. A has one split_checker_body; B has four. Deferred generic units are not proven clean.

The timer contract produces one observed main stop, not four separate declaration-token stops: `src/compiler/sys.ts:384:79`, NotYet / refusal_scan, `a member typed any ((0)) related to string at the argument site`. The System declarations at 1447 and 1448 have no independent recorded site. The emission line is:

`internal/lower/any_relations.go:43`:
```go
l.refusal.deferred = l.notYet(node, "a member typed any ("+found.path+") related to "+l.checker.TypeToString(found.target)+" at the "+name+" site")
```

That callback stop is absent in both candidate streams. This is not a successful-lowering claim: the registering function at sys.ts:383 is checker-split in both candidates. A additionally records these new sites:

`src/compiler/sys.ts:1502:13`, Refused / refusal_scan: `a function taking Timeout | undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take`.

`src/compiler/sys.ts:1519:36`, NotYet / refusal_scan: `a member typed any (require()module) related to {} at the initializer site`.

`src/compiler/sys.ts:1600:13`, Refused / refusal_scan: `a function taking Timeout | undefined seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take`.

The two Timeout-to-unknown findings at sys.ts:1502 and 1600 are direct new clearTimeout-contract refusals. Stock tsc accepts these method assignments by bivariance; Adamic deliberately refuses the wider input. The message and raising return are:

`internal/lower/invariance.go:1042`:
```go
what = "a function taking " + l.checker.TypeToString(found.source) + " seen as one taking " + l.checker.TypeToString(found.target) + " (tsc relates a method's parameters both ways), so it can be handed what it can't take"
```

`internal/lower/invariance.go:1053`:
```go
return &Refused{Where: l.program.Where(node), What: what, Fix: fix}
```

The new any-related initializer finding at sys.ts:1519:36 is collateral scanner exposure, not a new timer declaration token. A retains 122 main sites. B records no new site, but its getNodeSystem body is checker-split by the two failed generic timer assignments; the lack of A's new cancellation refusals in B does not prove them safe. B also splits the two zero-parameter callback registration bodies and the dynamic polling registration body, so its 21 absent observations cannot be credited as fixes.

The private `pollScheduled: any` finding at sys.ts:470:9 remains under both A and B: `a value of type any`, raised by `internal/lower/expression.go:53`. Retyping the public slot does not retype private storage. All 123/125/102 sites, their recording units, and every added/absent row are in latent-sites.json.
