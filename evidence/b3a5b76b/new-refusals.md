# New refusals at b3a5b76b

Base: ff41896cee93f832a4a8fa68c4c1822ddef5885e. Head: b3a5b76b91e61baea9db707dd121dc5664c473cc.

Count: **3**, rather than the expected zero. Verdict: **pass**; complete corpus: True; programs: 4957; omitted: 0. New refusals are informational to this gate.

## internal/oracle/testdata/host_watch/tsc_shape.a

```text
adamic: /tmp/adamic-gate/admission-delta-4232541155/head/internal/oracle/testdata/host_watch/tsc_shape.a:15:9: stage 0 can't lower a generic method on of type { <E extends keyof FSWatcherEventMap>(eventName: E, listener: (...args: FSWatcherEventMap[E]) => void): FSWatcher; (eventName: string | symbol, listener: (...args: any[]) => void): FSWatcher; } seen as (eventName: string, listener: () => void) => void yet
```

## internal/oracle/testdata/host_watch/view_after_close.a

```text
adamic: /tmp/adamic-gate/admission-delta-4232541155/head/internal/oracle/testdata/host_watch/view_after_close.a:5:39: stage 0 can't lower a generic method on of type { <E extends keyof FSWatcherEventMap>(eventName: E, listener: (...args: FSWatcherEventMap[E]) => void): FSWatcher; (eventName: string | symbol, listener: (...args: any[]) => void): FSWatcher; } seen as (eventName: string, listener: () => void) => void yet
```

## internal/oracle/testdata/host_watch/view_other_event.a

```text
adamic: /tmp/adamic-gate/admission-delta-4232541155/head/internal/oracle/testdata/host_watch/view_other_event.a:5:39: stage 0 can't lower a generic method on of type { <E extends keyof FSWatcherEventMap>(eventName: E, listener: (...args: FSWatcherEventMap[E]) => void): FSWatcher; (eventName: string | symbol, listener: (...args: any[]) => void): FSWatcher; } seen as (eventName: string, listener: () => void) => void yet
```

None of these three is an override refusal from class_inheritance.go or the constrained-string handle view from the earlier request.
