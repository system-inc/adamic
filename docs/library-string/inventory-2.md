# Baseline String refusal inventory

All 267 first refusals, from the observation-only audit build of the mandated runner. Its table exactly matches the unmodified runner. Ownership is inferred from the first obstruction; a test can have further dependencies. See claim-2.md for language reproducers. `new` splits into 42 String tests and seven other constructors.

| Test262 path | Observed refusal |
|---|---|
| built-ins/String/15.5.5.5.2-3-3.js | not yet: new an Identifier |
| built-ins/String/15.5.5.5.2-3-4.js | not yet: new an Identifier |
| built-ins/String/15.5.5.5.2-3-5.js | not yet: new an Identifier |
| built-ins/String/15.5.5.5.2-7-1.js | not yet: new an Identifier |
| built-ins/String/15.5.5.5.2-7-3.js | not yet: new an Identifier |
| built-ins/String/S15.5.1.1_A1_T1.js | not yet: a void call used as a value |
| built-ins/String/S15.5.1.1_A1_T13.js | not yet: reading Boolean |
| built-ins/String/S15.5.1.1_A1_T19.js | not yet: new an Identifier |
| built-ins/String/S15.5.1.1_A1_T3.js | refuses the void operator |
| built-ins/String/S15.5.1.1_A1_T5.js | not yet: reading x |
| built-ins/String/S15.5.1.1_A1_T7.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/S15.5.1.1_A1_T8.js | refuses a method read as a value (toString would lose its object, and this with it) |
| built-ins/String/S15.5.1.1_A1_T9.js | refuses var |
| built-ins/String/S15.5.2.1_A1_T1.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T11.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T12.js | refuses a method read as a value (valueOf would lose its object, and this with it) |
| built-ins/String/S15.5.2.1_A1_T13.js | refuses a method read as a value (valueOf would lose its object, and this with it) |
| built-ins/String/S15.5.2.1_A1_T16.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T17.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T18.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T19.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T2.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T3.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T4.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T6.js | refuses != |
| built-ins/String/S15.5.2.1_A1_T7.js | refuses a method read as a value (toString would lose its object, and this with it) |
| built-ins/String/S15.5.2.1_A1_T8.js | refuses a method read as a value (toString would lose its object, and this with it) |
| built-ins/String/S15.5.2.1_A1_T9.js | refuses != |
| built-ins/String/S15.5.2.1_A2_T1.js | not yet: new an Identifier |
| built-ins/String/S15.5.2.1_A3.js | refuses a method read as a value (toString would lose its object, and this with it) |
| built-ins/String/S15.5.3_A1.js | not yet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) |
| built-ins/String/S15.5.3_A2_T1.js | refuses isPrototypeOf |
| built-ins/String/S15.5.5.1_A1.js | not yet: new an Identifier |
| built-ins/String/S15.5.5.1_A2.js | not yet: new an Identifier |
| built-ins/String/S15.5.5.1_A5.js | refuses a method read as a value (valueOf would lose its object, and this with it) |
| built-ins/String/S8.12.8_A1.js | not yet: new an Identifier |
| built-ins/String/S8.12.8_A2.js | not yet: new an Identifier |
| built-ins/String/S9.8_A5_T1.js | refuses the void operator |
| built-ins/String/fromCharCode/S15.5.3.2_A1.js | refuses a method read as a value (fromCharCode would lose its object, and this with it) |
| built-ins/String/fromCharCode/S9.7_A1.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/fromCharCode/S9.7_A2.1.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/fromCharCode/S9.7_A2.2.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/fromCharCode/S9.7_A3.2_T1.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/S15.5.3.1_A1.js | not yet: hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow) |
| built-ins/String/prototype/S15.5.3.1_A2.js | not yet: hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow) |
| built-ins/String/prototype/S15.5.4_A2.js | refuses != |
| built-ins/String/prototype/charAt/S15.5.4.4_A11.js | refuses a method read as a value (charAt would lose its object, and this with it) |
| built-ins/String/prototype/charAt/S15.5.4.4_A3.js | not yet: new an Identifier |
| built-ins/String/prototype/charAt/S15.5.4.4_A4_T1.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/charAt/S15.5.4.4_A4_T2.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/charAt/S15.5.4.4_A4_T3.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/charAt/S15.5.4.4_A6.js | refuses a method read as a value (charAt would lose its object, and this with it) |
| built-ins/String/prototype/charAt/S15.5.4.4_A8.js | refuses a method read as a value (charAt would lose its object, and this with it) |
| built-ins/String/prototype/charAt/pos-coerce-err.js | not yet: a value of type any |
| built-ins/String/prototype/charAt/this-value-not-obj-coercible.js | refuses a method read as a value (charAt would lose its object, and this with it) |
| built-ins/String/prototype/charCodeAt/S15.5.4.5_A11.js | refuses a method read as a value (charCodeAt would lose its object, and this with it) |
| built-ins/String/prototype/charCodeAt/S15.5.4.5_A3.js | not yet: new an Identifier |
| built-ins/String/prototype/charCodeAt/S15.5.4.5_A6.js | refuses a method read as a value (charCodeAt would lose its object, and this with it) |
| built-ins/String/prototype/charCodeAt/S15.5.4.5_A8.js | refuses a method read as a value (charCodeAt would lose its object, and this with it) |
| built-ins/String/prototype/charCodeAt/pos-coerce-err.js | not yet: a value of type any |
| built-ins/String/prototype/charCodeAt/this-value-not-obj-coercible.js | refuses a method read as a value (charCodeAt would lose its object, and this with it) |
| built-ins/String/prototype/codePointAt/return-abrupt-from-this.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/codePointAt/this-is-null-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/codePointAt/this-is-undefined-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/concat/S15.5.4.6_A11.js | refuses a method read as a value (concat would lose its object, and this with it) |
| built-ins/String/prototype/concat/S15.5.4.6_A3.js | refuses != |
| built-ins/String/prototype/concat/S15.5.4.6_A6.js | refuses a method read as a value (concat would lose its object, and this with it) |
| built-ins/String/prototype/concat/S15.5.4.6_A8.js | refuses a method read as a value (concat would lose its object, and this with it) |
| built-ins/String/prototype/concat/this-value-not-obj-coercible.js | refuses a method read as a value (concat would lose its object, and this with it) |
| built-ins/String/prototype/constructor/S15.5.4.1_A1_T1.js | refuses inherited library member constructor read as an own field |
| built-ins/String/prototype/endsWith/String.prototype.endsWith_Fail_2.js | not yet: endsWith with these arguments |
| built-ins/String/prototype/endsWith/String.prototype.endsWith_Success_2.js | not yet: endsWith with these arguments |
| built-ins/String/prototype/endsWith/String.prototype.endsWith_Success_3.js | not yet: endsWith with these arguments |
| built-ins/String/prototype/endsWith/String.prototype.endsWith_Success_4.js | not yet: endsWith with these arguments |
| built-ins/String/prototype/endsWith/return-abrupt-from-this.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/endsWith/return-false-if-search-start-is-less-than-zero.js | not yet: endsWith with these arguments |
| built-ins/String/prototype/endsWith/return-true-if-searchstring-is-empty.js | not yet: endsWith with these arguments |
| built-ins/String/prototype/endsWith/searchstring-found-with-position.js | not yet: endsWith with these arguments |
| built-ins/String/prototype/endsWith/searchstring-not-found-with-position.js | not yet: endsWith with these arguments |
| built-ins/String/prototype/endsWith/this-is-null-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/endsWith/this-is-undefined-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/includes/String.prototype.includes_lengthProp.js | refuses a method read as a value (includes would lose its object, and this with it) |
| built-ins/String/prototype/includes/return-abrupt-from-this.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/includes/this-is-null-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/includes/this-is-undefined-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/indexOf/S15.5.4.7_A11.js | refuses a method read as a value (indexOf would lose its object, and this with it) |
| built-ins/String/prototype/indexOf/S15.5.4.7_A1_T12.js | not yet: new an Identifier |
| built-ins/String/prototype/indexOf/S15.5.4.7_A2_T1.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/indexOf/S15.5.4.7_A2_T2.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/indexOf/S15.5.4.7_A2_T3.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/indexOf/S15.5.4.7_A2_T4.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/indexOf/S15.5.4.7_A3_T1.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/indexOf/S15.5.4.7_A3_T3.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/indexOf/S15.5.4.7_A5_T1.js | refuses the comma operator |
| built-ins/String/prototype/indexOf/S15.5.4.7_A5_T2.js | refuses the comma operator |
| built-ins/String/prototype/indexOf/S15.5.4.7_A5_T3.js | refuses the comma operator |
| built-ins/String/prototype/indexOf/S15.5.4.7_A5_T4.js | refuses the comma operator |
| built-ins/String/prototype/indexOf/S15.5.4.7_A5_T5.js | refuses the comma operator |
| built-ins/String/prototype/indexOf/S15.5.4.7_A5_T6.js | refuses the comma operator |
| built-ins/String/prototype/indexOf/S15.5.4.7_A6.js | refuses a method read as a value (indexOf would lose its object, and this with it) |
| built-ins/String/prototype/indexOf/S15.5.4.7_A8.js | refuses a method read as a value (indexOf would lose its object, and this with it) |
| built-ins/String/prototype/indexOf/this-value-not-obj-coercible.js | refuses a method read as a value (indexOf would lose its object, and this with it) |
| built-ins/String/prototype/isWellFormed/return-abrupt-from-this.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/isWellFormed/returns-boolean.js | refuses inherited library member isWellFormed read as an own field |
| built-ins/String/prototype/isWellFormed/to-string.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/lastIndexOf/S15.5.4.8_A11.js | refuses a method read as a value (lastIndexOf would lose its object, and this with it) |
| built-ins/String/prototype/lastIndexOf/S15.5.4.8_A1_T12.js | not yet: new an Identifier |
| built-ins/String/prototype/lastIndexOf/S15.5.4.8_A6.js | refuses a method read as a value (lastIndexOf would lose its object, and this with it) |
| built-ins/String/prototype/lastIndexOf/S15.5.4.8_A8.js | refuses a method read as a value (lastIndexOf would lose its object, and this with it) |
| built-ins/String/prototype/lastIndexOf/this-value-not-obj-coercible.js | refuses a method read as a value (lastIndexOf would lose its object, and this with it) |
| built-ins/String/prototype/localeCompare/S15.5.4.9_A11.js | refuses a method read as a value (localeCompare would lose its object, and this with it) |
| built-ins/String/prototype/localeCompare/S15.5.4.9_A1_T2.js | not yet: localeCompare (Node uses locale collation, not ordinal UTF-16 order) |
| built-ins/String/prototype/localeCompare/S15.5.4.9_A6.js | refuses a method read as a value (localeCompare would lose its object, and this with it) |
| built-ins/String/prototype/localeCompare/S15.5.4.9_A8.js | refuses a method read as a value (localeCompare would lose its object, and this with it) |
| built-ins/String/prototype/localeCompare/this-value-not-obj-coercible.js | refuses a method read as a value (localeCompare would lose its object, and this with it) |
| built-ins/String/prototype/match/S15.5.4.10_A6.js | refuses a method read as a value (match would lose its object, and this with it) |
| built-ins/String/prototype/match/S15.5.4.10_A8.js | refuses a method read as a value (match would lose its object, and this with it) |
| built-ins/String/prototype/match/this-value-not-obj-coercible.js | refuses a method read as a value (match would lose its object, and this with it) |
| built-ins/String/prototype/normalize/return-abrupt-from-this.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/normalize/this-is-null-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/normalize/this-is-undefined-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/repeat/count-is-infinity-throws.js | not yet: a try around repeat, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |
| built-ins/String/prototype/repeat/count-less-than-zero-throws.js | not yet: a try around repeat, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |
| built-ins/String/prototype/replace/S15.5.4.11_A3_T1.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/replace/S15.5.4.11_A4_T1.js | refuses arguments |
| built-ins/String/prototype/replace/S15.5.4.11_A4_T2.js | refuses arguments |
| built-ins/String/prototype/replace/S15.5.4.11_A4_T3.js | refuses arguments |
| built-ins/String/prototype/replace/S15.5.4.11_A4_T4.js | refuses arguments |
| built-ins/String/prototype/replace/S15.5.4.11_A6.js | refuses a method read as a value (replace would lose its object, and this with it) |
| built-ins/String/prototype/replace/regexp-capture-by-index.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/replaceAll/getSubstitution-0x0024-0x0024.js | not yet: a value of type any |
| built-ins/String/prototype/replaceAll/getSubstitution-0x0024-0x0026.js | not yet: a value of type any |
| built-ins/String/prototype/replaceAll/getSubstitution-0x0024-0x0027.js | not yet: a value of type any |
| built-ins/String/prototype/replaceAll/getSubstitution-0x0024-0x0060.js | not yet: a value of type any |
| built-ins/String/prototype/replaceAll/getSubstitution-0x0024.js | not yet: a value of type any |
| built-ins/String/prototype/replaceAll/replaceValue-fn-skip-toString.js | refuses Object.defineProperty |
| built-ins/String/prototype/replaceAll/searchValue-empty-string.js | not yet: a value of type any |
| built-ins/String/prototype/search/S15.5.4.12_A11.js | refuses a method read as a value (search would lose its object, and this with it) |
| built-ins/String/prototype/search/S15.5.4.12_A1_T14.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/search/S15.5.4.12_A2_T1.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A2_T2.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A2_T3.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A2_T4.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A2_T5.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A2_T6.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A2_T7.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A3_T1.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A3_T2.js | not yet: new an Identifier |
| built-ins/String/prototype/search/S15.5.4.12_A6.js | refuses a method read as a value (search would lose its object, and this with it) |
| built-ins/String/prototype/search/S15.5.4.12_A8.js | refuses a method read as a value (search would lose its object, and this with it) |
| built-ins/String/prototype/search/this-value-not-obj-coercible.js | refuses a method read as a value (search would lose its object, and this with it) |
| built-ins/String/prototype/slice/S15.5.4.13_A11.js | refuses a method read as a value (slice would lose its object, and this with it) |
| built-ins/String/prototype/slice/S15.5.4.13_A1_T14.js | not yet: a function returning undefined |
| built-ins/String/prototype/slice/S15.5.4.13_A1_T6.js | refuses inherited library member slice read as an own field |
| built-ins/String/prototype/slice/S15.5.4.13_A1_T8.js | refuses the void operator |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T1.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T2.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T3.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T4.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T5.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T6.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T7.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T8.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A2_T9.js | not yet: new an Identifier |
| built-ins/String/prototype/slice/S15.5.4.13_A6.js | refuses a method read as a value (slice would lose its object, and this with it) |
| built-ins/String/prototype/slice/S15.5.4.13_A8.js | refuses a method read as a value (slice would lose its object, and this with it) |
| built-ins/String/prototype/slice/this-value-not-obj-coercible.js | refuses a method read as a value (slice would lose its object, and this with it) |
| built-ins/String/prototype/slice/this-value-tostring-throws-toprimitive.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/split/checking-if-enumerating-the-string-prototype-split-length-property-fails.js | refuses a method read as a value (split would lose its object, and this with it) |
| built-ins/String/prototype/split/checking-string-prototype-split-length.js | refuses a method read as a value (split would lose its object, and this with it) |
| built-ins/String/prototype/split/checking-string-prototype-split-prototype.js | refuses a method read as a value (split would lose its object, and this with it) |
| built-ins/String/prototype/split/separator-regexp.js | error TS2345: Argument of type '…' is not assignable to parameter of type '…'. (tsc: TS1125,TS1532) |
| built-ins/String/prototype/startsWith/out-of-bounds-position.js | not yet: startsWith with these arguments |
| built-ins/String/prototype/startsWith/return-abrupt-from-this.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/startsWith/return-true-if-searchstring-is-empty.js | not yet: startsWith with these arguments |
| built-ins/String/prototype/startsWith/searchstring-found-with-position.js | not yet: startsWith with these arguments |
| built-ins/String/prototype/startsWith/searchstring-not-found-with-position.js | not yet: startsWith with these arguments |
| built-ins/String/prototype/startsWith/this-is-null-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/startsWith/this-is-undefined-throws.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/substring/S15.5.4.15_A11.js | refuses a method read as a value (substring would lose its object, and this with it) |
| built-ins/String/prototype/substring/S15.5.4.15_A1_T8.js | refuses the void operator |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T10.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T2.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T3.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T4.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T5.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T6.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T7.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T8.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A2_T9.js | not yet: new an Identifier |
| built-ins/String/prototype/substring/S15.5.4.15_A6.js | refuses a method read as a value (substring would lose its object, and this with it) |
| built-ins/String/prototype/substring/S15.5.4.15_A8.js | refuses a method read as a value (substring would lose its object, and this with it) |
| built-ins/String/prototype/substring/this-value-not-obj-coercible.js | refuses a method read as a value (substring would lose its object, and this with it) |
| built-ins/String/prototype/toLocaleLowerCase/Final_Sigma_U180E.js | refuses inherited library member toLocaleLowerCase read as an own field |
| built-ins/String/prototype/toLocaleLowerCase/S15.5.4.17_A11.js | refuses a method read as a value (toLocaleLowerCase would lose its object, and this with it) |
| built-ins/String/prototype/toLocaleLowerCase/S15.5.4.17_A1_T5.js | refuses inherited library member toLocaleLowerCase read as an own field |
| built-ins/String/prototype/toLocaleLowerCase/S15.5.4.17_A2_T1.js | refuses inherited library member toLocaleLowerCase read as an own field |
| built-ins/String/prototype/toLocaleLowerCase/S15.5.4.17_A6.js | refuses a method read as a value (toLocaleLowerCase would lose its object, and this with it) |
| built-ins/String/prototype/toLocaleLowerCase/S15.5.4.17_A8.js | refuses a method read as a value (toLocaleLowerCase would lose its object, and this with it) |
| built-ins/String/prototype/toLocaleLowerCase/special_casing.js | refuses inherited library member toLocaleLowerCase read as an own field |
| built-ins/String/prototype/toLocaleLowerCase/special_casing_conditional.js | refuses inherited library member toLocaleLowerCase read as an own field |
| built-ins/String/prototype/toLocaleLowerCase/supplementary_plane.js | refuses inherited library member toLocaleLowerCase read as an own field |
| built-ins/String/prototype/toLocaleLowerCase/this-value-not-obj-coercible.js | refuses a method read as a value (toLocaleLowerCase would lose its object, and this with it) |
| built-ins/String/prototype/toLocaleUpperCase/S15.5.4.19_A11.js | refuses a method read as a value (toLocaleUpperCase would lose its object, and this with it) |
| built-ins/String/prototype/toLocaleUpperCase/S15.5.4.19_A1_T5.js | refuses inherited library member toLocaleUpperCase read as an own field |
| built-ins/String/prototype/toLocaleUpperCase/S15.5.4.19_A2_T1.js | refuses inherited library member toLocaleUpperCase read as an own field |
| built-ins/String/prototype/toLocaleUpperCase/S15.5.4.19_A6.js | refuses a method read as a value (toLocaleUpperCase would lose its object, and this with it) |
| built-ins/String/prototype/toLocaleUpperCase/S15.5.4.19_A8.js | refuses a method read as a value (toLocaleUpperCase would lose its object, and this with it) |
| built-ins/String/prototype/toLocaleUpperCase/special_casing.js | refuses inherited library member toLocaleUpperCase read as an own field |
| built-ins/String/prototype/toLocaleUpperCase/supplementary_plane.js | refuses inherited library member toLocaleUpperCase read as an own field |
| built-ins/String/prototype/toLocaleUpperCase/this-value-not-obj-coercible.js | refuses a method read as a value (toLocaleUpperCase would lose its object, and this with it) |
| built-ins/String/prototype/toLowerCase/S15.5.4.16_A11.js | refuses a method read as a value (toLowerCase would lose its object, and this with it) |
| built-ins/String/prototype/toLowerCase/S15.5.4.16_A2_T1.js | not yet: new an Identifier |
| built-ins/String/prototype/toLowerCase/S15.5.4.16_A6.js | refuses a method read as a value (toLowerCase would lose its object, and this with it) |
| built-ins/String/prototype/toLowerCase/S15.5.4.16_A8.js | refuses a method read as a value (toLowerCase would lose its object, and this with it) |
| built-ins/String/prototype/toLowerCase/this-value-not-obj-coercible.js | refuses a method read as a value (toLowerCase would lose its object, and this with it) |
| built-ins/String/prototype/toLowerCase/this-value-tostring-throws-toprimitive.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/toString/string-object.js | refuses a method read as a value (toString would lose its object, and this with it) |
| built-ins/String/prototype/toString/string-primitive.js | refuses a method read as a value (toString would lose its object, and this with it) |
| built-ins/String/prototype/toUpperCase/S15.5.4.18_A11.js | refuses a method read as a value (toUpperCase would lose its object, and this with it) |
| built-ins/String/prototype/toUpperCase/S15.5.4.18_A2_T1.js | not yet: new an Identifier |
| built-ins/String/prototype/toUpperCase/S15.5.4.18_A6.js | refuses a method read as a value (toUpperCase would lose its object, and this with it) |
| built-ins/String/prototype/toUpperCase/S15.5.4.18_A8.js | refuses a method read as a value (toUpperCase would lose its object, and this with it) |
| built-ins/String/prototype/toUpperCase/this-value-not-obj-coercible.js | refuses a method read as a value (toUpperCase would lose its object, and this with it) |
| built-ins/String/prototype/toWellFormed/return-abrupt-from-this.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/toWellFormed/returns-well-formed-string.js | refuses inherited library member toWellFormed read as an own field |
| built-ins/String/prototype/toWellFormed/to-string.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-0-1.js | refuses a method read as a value (trim would lose its object, and this with it) |
| built-ins/String/prototype/trim/15.5.4.20-0-2.js | refuses a method read as a value (trim would lose its object, and this with it) |
| built-ins/String/prototype/trim/15.5.4.20-1-1.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/trim/15.5.4.20-1-2.js | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |
| built-ins/String/prototype/trim/15.5.4.20-1-5.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-1-6.js | not yet: new an Identifier |
| built-ins/String/prototype/trim/15.5.4.20-1-9.js | not yet: new an Identifier |
| built-ins/String/prototype/trim/15.5.4.20-2-34.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-2-35.js | not yet: new an Identifier |
| built-ins/String/prototype/trim/15.5.4.20-2-36.js | not yet: new an Identifier |
| built-ins/String/prototype/trim/15.5.4.20-2-37.js | not yet: new an Identifier |
| built-ins/String/prototype/trim/15.5.4.20-2-38.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-2-39.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-2-40.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-2-41.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-2-42.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-2-45.js | not yet: a BinaryExpression with a string and a number |
| built-ins/String/prototype/trim/15.5.4.20-2-46.js | refuses arguments |
| built-ins/String/prototype/trim/15.5.4.20-2-47.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trim/15.5.4.20-2-49.js | not yet: RegExp with a nonconstant pattern |
| built-ins/String/prototype/trim/15.5.4.20-2-50.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/prototype/trimEnd/this-value-boolean.js | refuses a method read as a value (trimEnd would lose its object, and this with it) |
| built-ins/String/prototype/trimEnd/this-value-line-terminator.js | refuses a method read as a value (trimEnd would lose its object, and this with it) |
| built-ins/String/prototype/trimEnd/this-value-not-obj-coercible.js | refuses a method read as a value (trimEnd would lose its object, and this with it) |
| built-ins/String/prototype/trimEnd/this-value-number.js | refuses a method read as a value (trimEnd would lose its object, and this with it) |
| built-ins/String/prototype/trimEnd/this-value-whitespace.js | refuses a method read as a value (trimEnd would lose its object, and this with it) |
| built-ins/String/prototype/trimStart/this-value-boolean.js | refuses a method read as a value (trimStart would lose its object, and this with it) |
| built-ins/String/prototype/trimStart/this-value-line-terminator.js | refuses a method read as a value (trimStart would lose its object, and this with it) |
| built-ins/String/prototype/trimStart/this-value-not-obj-coercible.js | refuses a method read as a value (trimStart would lose its object, and this with it) |
| built-ins/String/prototype/trimStart/this-value-number.js | refuses a method read as a value (trimStart would lose its object, and this with it) |
| built-ins/String/prototype/trimStart/this-value-whitespace.js | refuses a method read as a value (trimStart would lose its object, and this with it) |
| built-ins/String/prototype/valueOf/string-object.js | refuses a method read as a value (valueOf would lose its object, and this with it) |
| built-ins/String/prototype/valueOf/string-primitive.js | refuses a method read as a value (valueOf would lose its object, and this with it) |
| built-ins/String/raw/return-empty-string-from-empty-array-length.js | not yet: String.raw with raw elements that are not strings |
| built-ins/String/raw/return-empty-string-if-length-is-negative-infinity.js | not yet: String.raw without a present array of strings in raw |
| built-ins/String/raw/return-empty-string-if-length-is-zero-NaN.js | not yet: String.raw without a present array of strings in raw |
| built-ins/String/raw/return-empty-string-if-length-is-zero-or-less-number.js | not yet: String.raw without a present array of strings in raw |
| built-ins/String/raw/returns-abrupt-from-substitution.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/raw/substitutions-are-limited-to-template-raw-length.js | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |
| built-ins/String/raw/template-raw-throws.js | error TS2741: Property '…' is missing in type '…' but required in type '…'. (tsc: TS2345) |
