# Phantom string brands

The representation proof is structural, rather than keyed to the name `__String`.
It admits a string intersected with finite object types whose members are all
phantom `void` fields, including optional fields with undefined, and unions of
such brands with strings, string literals, string enums, undefined, and phantom
void intersections. At least one string member is required. Callable,
constructable and indexed brands, real fields, and names that collide with
JavaScript primitive properties do not pass the proof. Ordinary object fields
and number or boolean brands keep their existing representation rules.

Casts between these proven string shapes may erase phantom fields when their
underlying string literal types are assignable. A cast removing possible
undefined inserts the existing Defined check in both backends. It does not
authorize unknown values, real fields, or string literal refinements.

An admitted string brand uses the ordinary string pointer and ownership rules.
Undefined uses the missing string pointer. A phantom void intersection is
conservatively treated as undefined, rather than assumed uninhabited: main's
view descriptors already name this arm undefined. A separate review probe
shows that the current read boundary rejects an undefined payload asserted
as __String. This unit does not change that runtime view boundary. Unknown
to brand casts remain refused.

A brand member accessed by a property read, an indexed read, or object
binding destructuring is refused by name, as is a real brand field even when
unread, when a runtime expression or implemented signature carries the brand.
Unused type declarations and overload signatures retain their existing handling.
The refusal applies to writes through the same access syntax too.
Mixed unions containing a proven string brand use the existing union rules;
operation-specific admission checks continue to apply.

The tsc-derived enum witnesses use an ordinary enum with selected real string
members. The upstream enum is const; reducing its declaration avoids the
separate const-enum admission rule. The alias shape itself is exact. `Symbol`
is tsc's object interface, reduced to a readonly name, not JavaScript's symbol
primitive.
