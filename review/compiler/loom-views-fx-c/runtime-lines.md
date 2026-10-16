C runtime changes against bd90a433, object.c lines 177-178:

```c
        value.payload = (adamic_value){0};
        if (boolean.present) { value.payload.boolean = boolean.boolean; }
```

Replaces: `value.payload.boolean = boolean.boolean;`
No other runtime C lines change.
