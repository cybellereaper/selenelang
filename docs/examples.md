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

## Try your own

Create a new file under `examples/` and run it with the CLI. Expressions, function definitions, variable declarations, and `match` statements are supported today, giving you enough building blocks to sketch real control flow.
