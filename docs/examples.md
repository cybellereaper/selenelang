---
layout: default
title: Example Scripts
---

# Example scripts

Selene ships with runnable scripts that illustrate different language features. Execute any of them with `selene path/to/script.selene`.

## Greeting

**File:** `examples/hello.selene`

Introduces bindings, expression-bodied functions, and printing.

```selene
let greeting: String = "Hello";

fn greet(name: String): String => greeting + ", " + name;

fn main() {
    print(greet("Selene"));
}

main();
```

## Math utilities

**File:** `examples/math.selene`

Shows how to organize reusable arithmetic helpers and compose functions.

```selene
let pi: Number = 314159 / 100000;

fn circleArea(radius: Number): Number => radius * radius * pi;
fn circumference(radius: Number): Number => 2 * pi * radius;

fn report(name: String, radius: Number) {
    let area = circleArea(radius);
    let edge = circumference(radius);
    print(name + " => area=" + area + ", circumference=" + edge);
}

report("small", 2);
report("medium", 5);
report("large", 9);
```

## Collections and objects

**File:** `examples/collections.selene`

Demonstrates arrays, objects, optional chaining, Elvis expressions, and string indexing.

```selene
let toolkit = {
    name: "Selene",
    version: "0.1.0",
    features: ["lexer", "parser", "runtime", "docs"],
    maintainer: null
};

fn describeToolkit() {
    print("name => " + toolkit.name);
    print("version => " + toolkit.version);
    print("first feature => " + toolkit.features[0]);
    let maintainer = toolkit.maintainer?.name ?: "Community";
    print("maintainer => " + maintainer);
    print("version major => " + toolkit.version[0]);
}

describeToolkit();
```

## Optional access patterns

**File:** `examples/options.selene`

Highlights optional chaining, Elvis defaults, and the non-null assertion when navigating nested data.

```selene
let profile = {
    name: "Nova",
    contact: {
        email: "nova@example.com",
        address: null
    }
};

let email = profile.contact?.email ?: "unknown@selene.dev";
print("Email: " + email);

let city = profile.contact?.address?.city ?: "Unknown";
print("City: " + city);

let safeContact = profile.contact!!;
print("Forced contact email: " + safeContact.email);
```

## Pattern matching

**File:** `examples/patterns.selene`

Uses `match` statements with object destructuring to branch on structured data.

```selene
let shapes = [
    { kind: "circle", radius: 2 },
    { kind: "square", side: 4 },
    { kind: "triangle", base: 3, height: 4 }
];

fn describe(shape: Any) {
    match shape {
        { kind: "circle", radius: r } => print("Circle with radius " + r);
        { kind: "square", side: s } => print("Square with side " + s);
        other => print("Unknown shape kind " + (other.kind ?: "n/a"));
    }
}

print("Cataloging shapes:");
describe(shapes[0]);
describe(shapes[1]);
describe(shapes[2]);
```

## Recursive helpers

**File:** `examples/recursion.selene`

Demonstrates using `match` clauses to implement recursion-driven control flow without loops.

```selene
fn factorial(n: Number) {
    match n {
        0 => 1;
        1 => 1;
        value => value * factorial(value - 1);
    }
}

print("Factorial results:");
print("5! = " + factorial(5));
print("7! = " + factorial(7));
```

## Control flow in action

**File:** `examples/control.selene`

Walks through `if`/`else`, `for`/`while`, and the `return`/`break`/`continue` statements.

```selene
var total = 0;

for (let i = 0; i < 10; i = i + 1) {
    if i % 2 == 0 {
        continue;
    }
    if i > 7 {
        break;
    }
    total = total + i;
}

var countdown = 3;
while countdown > 0 {
    print("tick", countdown);
    countdown = countdown - 1;
}

fn fib(n: Number): Number {
    if n <= 1 {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

print("total", total);
print("fib(6)", fib(6));

var attempts = 0;
while true {
    attempts = attempts + 1;
    if attempts == 3 {
        print("finished after", attempts, "tries");
        break;
    }
}
```

## Custom types

**File:** `examples/types.selene`

Introduces struct methods, simple classes, enums, and matching on enum cases.

```selene
struct Point(x: Number, y: Number) {
    fn describe(): String {
        return "Point(" + self.x + ", " + self.y + ")";
    }
}

class Greeter(name: String) {
    fn greet(target: String): String {
        return self.name + ", " + target;
    }
}

enum Option {
    Some(value: Number);
    None;
}

let origin = Point(0, 0);
let greeter = Greeter("Selene");
let value = Option.Some(41);

print(origin.describe());
print(greeter.greet("friend"));

match value {
    Some(v) => print("value + 1 =", v + 1);
    None => print("no value");
}
```

## Modules and imports

**File:** `examples/modules.selene`

Shows how to export helpers from a module and import them elsewhere in the same script.

```selene
module math_utils {
    fn square(n: Number): Number => n * n;
    let constants = { tau: 6.28318 };
}

import math_utils.square;
import math_utils.constants as consts;

print(square(5));
print("tau ≈" , consts.tau);
```

## Contracts and postconditions

**File:** `examples/contracts.selene`

Combines standalone contract declarations with inline function contracts to validate results.

```selene
contract validators {
    fn withinRange(value: Number, min: Number, max: Number): Boolean {
        return value >= min && value <= max;
    }
}

fn clamp(value: Number, min: Number, max: Number): Number
    contract {
        returns(result) => result >= min;
        returns(result) => result <= max;
    }
{
    if value < min {
        return min;
    }
    if value > max {
        return max;
    }
    return value;
}

let values = [ -5, 3, 42 ];
for (let i = 0; i < values.length; i = i + 1) {
    let original = values[i];
    let clamped = clamp(original, 0, 10);
    print("clamp(" + original + ") => " + clamped);
    print("within range? => " + validators.withinRange(clamped, 0, 10));
}
```

## Try your own

Create a new file under `examples/` and run it with the CLI. With modules, functions, structs/classes/enums, pattern matching, and imperative loops, Selene gives you enough building blocks to sketch real control flow.
