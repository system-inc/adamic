# The front door to main

Every change reaches main through Loom's queue at `https://loom.system.inc`. Nobody else pushes
main. Loom's contracts are in system-inc/loom, `docs/contracts.md`.

## Submit

Push your commit to any branch on system-inc/adamic, then post it with your submit token (scope `submit`, minted for
your username):

```sh
curl -sS -X POST https://loom.system.inc/changes \
  -H "Authorization: Bearer $LOOM_SUBMIT_TOKEN" -H 'Content-Type: application/json' \
  -d '{"sha":"<40 hex>","base":"<main sha it starts from>","owner":"<your username>","paths":["<every path the diff touches>"]}'
```

- `base` is an ancestor of `sha` and on main. Start from main's tip when you can: a change on an older main is tested
  merged onto the newest one.
- `paths` lists every path in `git diff --name-only base sha`, and nothing else.
- A change under `stage3/fixtures/` or `stage3/meter/` (other than `_test.go` or `testdata`) carries its mutant evidence:
  one of its paths names `mutant`.
- Optional: `parent` (the change id yours stacks on, at most 5 deep) and `fixesRed` (the red main sha it fixes).

The answer is `201 {"change": "chg_...", "state": "queued", "position": n}`, or `422 {"reason": ...}` when the line or
git rules it out. Keep the change id.

## What you hear

Only what you act on, in your inbox: `landed` (the main sha), `red` (the failing test, the diff, and
`loom repro <unitKey>`), or `parked` (why it can't land as it is; resubmit on main). Voids and retries are Loom's and
never reach you. Read your change any time with `GET /changes/<change>` and its events with
`GET /changes/<change>/events`.

## Resubmit

A red or parked change moves to a new sha and keeps its id: post the same body as a submit to
`POST /changes/<change>/sha` with your submit token. Only the sha, base and paths move, and a sha the queue already
tested is refused, so push a new commit first.

## Who lands it

The lander, on workshop with its own deploy key. It moves main only by a fast-forward to the exact tree that was tested,
never a force. A tested tree that isn't a fast-forward of main when its turn comes isn't pushed: the change parks, and
you resubmit it on main's tip.
