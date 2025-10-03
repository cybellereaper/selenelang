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

## Try your own

Create a new file under `examples/` and run it with the CLI. As long as you stick to expressions, function definitions, and
variable declarations, the interpreter will execute the script and print diagnostics for unsupported constructs.
