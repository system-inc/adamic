# New refusals

Comparison: `ff41896cee93f832a4a8fa68c4c1822ddef5885e` to `550acd97678cf5433a29649ded1110b6a92b4c08`.

Count: **3**. Admission tool verdict: **pass**. Complete corpus: `True`.

Flags identify override diagnostics originating in `internal/lower/class_inheritance.go`, and constrained generic method views that require inspection as potentially honest programs. The complete diagnostic is retained for each refusal.

## internal/oracle/testdata/host_watch/tsc_shape.a

```text
adamic: /tmp/adamic-gate/admission-delta-496418865/head/internal/oracle/testdata/host_watch/tsc_shape.a:15:9: stage 0 can't lower a generic method on of type { <E extends keyof FSWatcherEventMap>(eventName: E, listener: (...args: FSWatcherEventMap[E]) => void): FSWatcher; (eventName: string | symbol, listener: (...args: any[]) => void): FSWatcher; } seen as (eventName: string, listener: () => void) => void yet
```

## internal/oracle/testdata/host_watch/view_after_close.a

```text
adamic: /tmp/adamic-gate/admission-delta-496418865/head/internal/oracle/testdata/host_watch/view_after_close.a:5:39: stage 0 can't lower a generic method on of type { <E extends keyof FSWatcherEventMap>(eventName: E, listener: (...args: FSWatcherEventMap[E]) => void): FSWatcher; (eventName: string | symbol, listener: (...args: any[]) => void): FSWatcher; } seen as (eventName: string, listener: () => void) => void yet
```

## internal/oracle/testdata/host_watch/view_other_event.a

```text
adamic: /tmp/adamic-gate/admission-delta-496418865/head/internal/oracle/testdata/host_watch/view_other_event.a:5:39: stage 0 can't lower a generic method on of type { <E extends keyof FSWatcherEventMap>(eventName: E, listener: (...args: FSWatcherEventMap[E]) => void): FSWatcher; (eventName: string | symbol, listener: (...args: any[]) => void): FSWatcher; } seen as (eventName: string, listener: () => void) => void yet
```


None of these three source programs declares a class or a constrained `handle<T extends string>` method. Thus no observed new refusal is flagged as a generic override refused by `class_inheritance.go`, or as that honest constrained-handle view. This conclusion applies to this manifest; inline Go-test snippets and the externally fetched probes are separate evidence.
