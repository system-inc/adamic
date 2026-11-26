# structuralViewCast never answers nil (#bgm34gh)

**N1, the nil back.** In internal/lower/view_objects.go's structuralViewCast, the first guard's NotYet ("a checked cast
from ... whose value isn't a plain object") made `return nil, nil`, main's code before this change. A value held
otherwise than as a plain object answers no value and no error, which castRoadValue takes as the cast's value and
expression's noteViewArrayStorage dereferences. Red: `TestStructuralViewCastNeverAnswersNil/a_value_held_as_a_Union`
and `/a_value_held_as_a_number`.
