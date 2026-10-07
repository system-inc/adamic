export const messageNew =
    "This constructs an object and then throws it away, so the only thing the line can accomplish is whatever the constructor does on the side. A constructor that works by side effect is doing a function's job under a name that promises a value, and the next reader cannot tell the discarded result from a mistake. Call a function instead, or keep the object and use it.";
