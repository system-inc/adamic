# to_json_view.refused's row, #sbvnh96

The cycle finder's shape now holds a non-scalar object literal's regular widened type beside its fresh one
(#sbvnh96), so the JSON reach walk meets `{ value: 1, toJSON }`'s own type before `Hooked` and names it. The
refusal is the same rule on the same program, so only the type it names moves.

**Red without this row:** Loom job J-25zxw83, at the head before this row (81e15400, proof e324e9c5), went red
on `TestJSONObjectsRefused/to_json_view.a` alone in internal/oracle:

    diagnostic: got "...to_json_view.a:12:28: stage 0 can't lower JSON.stringify a value that can reach
    { value: number; toJSON: () => string; }, which has a toJSON member yet", want "...can reach Hooked,
    which has a toJSON member yet"

Native whole was green in the same job.
