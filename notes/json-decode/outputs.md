# Observed program outputs

These are final explicit build/run observations, separate from the three-way oracle.

## json_decode_coverage_calls.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
input 2
input wrong
1
first|at $.value: expected string, found number
Ok
local
Ok
```

## json_decode_coverage_depth_extra.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
1
1
invalid JSON at line 1 column 145: nesting depth exceeds 128
```

## json_decode_coverage_empty.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text


at $: expected {}, found array
at $: expected {}, found null
invalid JSON at line 1 column 15: expected JSON value
true
true,false,true
at $[1]: expected boolean, found number
at $: expected boolean[], found boolean
```

## json_decode_coverage_grammar.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
invalid JSON at line 1 column 6: expected end of input
invalid JSON at line 1 column 4: invalid keyword
invalid JSON at line 1 column 5: invalid keyword
invalid JSON at line 1 column 2: expected digit
invalid JSON at line 1 column 4: expected digit
invalid JSON at line 1 column 4: expected digit
invalid JSON at line 1 column 4: expected digit
invalid JSON at line 1 column 4: expected four hexadecimal digits
invalid JSON at line 1 column 7: expected four hexadecimal digits
invalid JSON at line 1 column 3: unterminated string
invalid JSON at line 1 column 5: expected ':'
invalid JSON at line 1 column 10: expected ',' or '}'
invalid JSON at line 1 column 7: expected ',' or ']'
invalid JSON at line 1 column 10: expected ',' or '}'
at $: expected string, found object
1
43981
3
55296
55296
56320
2
56320
55296
```

## json_decode_coverage_layouts.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
count,after - 1 a -1
before,count,after b 2 c -1
count,middle,after - 3 d -1
count,after,end - 4 e 1
before,count,middle,after,end g 5 h 0
at $.count: expected number, found string
at $.end[1]: expected string, found boolean
count,after - 1 a -1
before,count,after b 2 c -1
count,middle,after - 3 d -1
count,after,end - 4 e 1
before,count,middle,after,end g 5 h 0
at $.count: expected number, found string
at $.end[1]: expected string, found boolean
count,after - 1 a -1
before,count,after b 2 c -1
count,middle,after - 3 d -1
count,after,end - 4 e 1
before,count,middle,after,end g 5 h 0
at $.count: expected number, found string
at $.end[1]: expected string, found boolean
```

## json_decode_coverage_nested.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
2 2 2 absent
false 2
55296
10
at $.rows[0][1]: missing field enabled
at $.rows[0][0].text: expected string, found number
at $.rows: expected readonly (readonly Leaf[])[], found object
at $.pair[1][1]: expected number, found string
at $.optional: expected Leaf, found null
0 2 0 absent
invalid JSON at line 1 column 23: expected end of input
```

## json_decode_coverage_object_scalar.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
Ok
Ok
at $: missing field value
at $.value: expected number, found string
at $: expected Value, found number
at $: expected Value, found null
```

## json_decode_coverage_tags.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
4
8
at $.tag: expected N, found number
at $.tag: expected N, found string
at $: missing field tag
at $: missing field value
at $.value: expected number, found string
at $: expected N, found array
yes
0
at $.flag: expected B, found null
at $.count: expected number, found string
```

## json_decode_coverage_unicode_tag.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
Ok
Ok
at $.�: expected Tagged, found string
at $: missing field �
at $.count: expected number, found boolean
```

## json_decode_coverage_unions.a

Native and source Node stdout (identical), both exit 0; stderr empty:

```text
Ok
Ok
Ok
Ok
at $: expected string | number | boolean, found null
at $: expected string | number | boolean, found array
at $: expected string | number | boolean, found object
```

## json_decode_coverage_shapes.a

Native, source Node and JavaScript backend stdout (identical), all exit 0; stderr empty:

```text
2 allocated false 2 different name
false
2 allocated false 2 different name
false
2 allocated false 2 different name
false
```

## Compile-time limits

These failed to produce native binaries. Node results below are from the source runner, whose descriptor generator also refuses non-JSON types.

### mixed_array.a

Compiler output:

```text
adamic: /workspace/adamic/notes/json-decode/mixed_array.a:5:23: stage 0 can't lower an array of string | number | boolean yet
exit status 1
```

Node stdout:

```text
number 1
string two
boolean false
at $[2]: expected string | number | boolean, found null
```

Node stderr:

```text

```

### mixed_read.a

Compiler output:

```text
adamic: /workspace/adamic/notes/json-decode/mixed_read.a:2:81: stage 0 can't lower a field of type string | number yet
exit status 1
```

Node stdout:

```text
1
```

Node stderr:

```text

```

### number_undefined.a

Compiler output:

```text
adamic: /workspace/adamic/notes/json-decode/number_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
```

Node stdout:

```text

```

Node stderr:

```text
/workspace/adamic/notes/json-decode/number_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
adamic: panic: Error: Command failed: go run ./oracle/json_types.go /workspace/adamic/notes/json-decode/number_undefined.a
/workspace/adamic/notes/json-decode/number_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
```

### optional_boolean_read.a

Compiler output:

```text
adamic: /workspace/adamic/notes/json-decode/optional_boolean_read.a:2:94: stage 0 can't lower a field of type boolean | undefined yet
exit status 1
```

Node stdout:

```text
false
```

Node stderr:

```text

```

### optional_undefined.a

Compiler output:

```text
adamic: /workspace/adamic/notes/json-decode/optional_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove number | undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
```

Node stdout:

```text

```

Node stderr:

```text
/workspace/adamic/notes/json-decode/optional_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove number | undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
adamic: panic: Error: Command failed: go run ./oracle/json_types.go /workspace/adamic/notes/json-decode/optional_undefined.a
/workspace/adamic/notes/json-decode/optional_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove number | undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
```

### required_undefined.a

Compiler output:

```text
adamic: /workspace/adamic/notes/json-decode/required_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
```

Node stdout:

```text

```

Node stderr:

```text
/workspace/adamic/notes/json-decode/required_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
adamic: panic: Error: Command failed: go run ./oracle/json_types.go /workspace/adamic/notes/json-decode/required_undefined.a
/workspace/adamic/notes/json-decode/required_undefined.a:2:1: Adamic 0.1 refuses decodeJson cannot prove undefined is JSON data; name a data type made of JSON scalars, arrays, tuples and plain fields; give object unions one distinct literal discriminant
exit status 1
```


## json_decode_coverage_arrays.a

Node and native agree, exit 0 and empty stderr:

```text
7
one -
two yes
count 3
at $[1].value: expected number, found string
at $[0]: missing field kind
at $[0].kind: expected Item, found string
at $[0]: missing field value
at $: expected readonly Item[], found object
Ok
Ok
at $[1]: expected string | number | boolean, found null
at $[0]: expected string | number | boolean, found object
at $: expected readonly (string | number | boolean)[], found number
```
