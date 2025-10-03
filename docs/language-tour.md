---
layout: default
title: Language Tour
---

# Language tour

Selene emphasizes expression-oriented programming with a concise syntax. This tour highlights the features implemented in the
current interpreter so you can design scripts and DSLs with confidence.

## Literals and bindings

Selene supports numbers, strings, booleans, and null. Bindings are introduced with `let` and are immutable by default:

```selene
let greeting: String = "Hello";
let version: Number = 1;
let debug: Boolean = true;
let missing = null;
```

Bindings are looked up dynamically at runtime. Type annotations are optional but help document intent.

## Arithmetic and comparison

Numbers participate in the standard arithmetic and comparison operators:

```selene
let radius = 5;
let area = radius * radius * (314159 / 100000);
let isLarge = area > 50;
print("area => " + area);
print("large? " + isLarge);
```

## Functions

Define functions with `fn`. They close over the lexical environment and may use expression bodies (`=>`) or block bodies:

```selene
fn double(value: Number): Number => value * 2;

fn describe(name: String, value: Number) {
    let doubled = double(value);
    print(name + " => " + doubled);
}

describe("sample", 21);
```

Functions return the value of their last expression, or you can use `return` to exit early from a block-bodied function.

## Arrays and objects

Use brackets for arrays and braces for objects. Indexing works on arrays and strings:

```selene
let features = ["lexer", "parser", "runtime"];
let project = {
    name: "Selene",
    items: features
};

print(project.name);
print("first feature => " + project.items[0]);
print("first letter => " + project.name[0]);
```

Arrays and strings expose a `length` property for quick sizing.

## Optional chaining and Elvis operator

Member lookups can be made optional with `?.`. Combine this with the Elvis operator `?:` to provide defaults when values are
null (or become null after an optional access):

```selene
let config = { name: "Selene", owner: null };
let owner = config.owner?.name ?: "Unknown";
print("owner => " + owner);
```

## Pattern matching

`match` statements provide a flexible way to branch on literals, bind values, and destructure objects. Each clause pattern is tested in order until one matches, and the body of the matching clause produces the statement result:

```selene
fn describe(shape: Any) {
    match shape {
        { kind: "circle", radius: r } => "circle => " + r;
        { kind: "square", side: s } => "square => " + s;
        other => "unknown => " + (other.kind ?: "n/a");
    }
}

print(describe({ kind: "circle", radius: 3 }));
```

Patterns may be simple literals (matched by value), identifiers (which bind to the target), object destructuring clauses that recursively match nested shapes, or struct patterns such as `Point(x, y)` that match constructor calls (including enum cases) by position.

## Control flow

Selene offers familiar imperative control flow. Use `if` for branching and `for`/`while` for iteration. `return`, `break`, and `continue` behave as you would expect from other languages:

```selene
var total = 0;

for (let i = 0; i < 5; i = i + 1) {
    if i % 2 == 0 {
        continue;
    }
    total = total + i;
}

while total > 4 {
    total = total - 1;
    if total == 6 {
        break;
    }
}

fn describe(value: Number): String {
    if value > 4 {
        return "large";
    }
    return "small";
}

print("summary => " + describe(total));
```

## Modules and imports

Group related helpers inside a `module` and import them elsewhere in the file. Everything defined inside the module body is exported automatically:

```selene
module math_utils {
    fn square(n: Number): Number => n * n;
    let constants = { tau: 6.28318 };
}

import math_utils.square;
import math_utils.constants as consts;

print(square(5));
print("tau => " + consts.tau);
```

## Contracts

Use `contract { ... }` blocks to attach postconditions to functions. Each `returns(condition)` clause evaluates after the
function body with a special `result` binding that contains the returned value. If a condition is falsy, Selene raises a
runtime error. Contracts can also capture parameters or other locals from the definition site:

```selene
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

print("clamp(42, 0, 10) => " + clamp(42, 0, 10));
```

Separate `contract` declarations create reusable bundles of helpers that can be accessed via dot syntax similar to modules.

## Structs, classes, and enums

Define your own data types with `struct`, `class`, and `enum`. Struct and class constructors behave like functions, and methods gain access to the current instance through `self`:

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

let value = Option.Some(41);
print(Point(2, 3).describe());
print(Greeter("Selene").greet("friend"));

match value {
    Some(v) => print("value + 1 =" + (v + 1));
    None => print("no value");
}
```

## Function application and scoping

Functions capture the environment in which they are defined:

```selene
let prefix = "[Selene]";

fn makeLogger(label: String) {
    fn log(message: String) {
        print(prefix + " " + label + ": " + message);
    }

    log("logger ready");
}

makeLogger("demo");
```

Each function call receives a fresh scope for its parameters, ensuring predictable lexical scoping.

## Limitations and roadmap

The interpreter now covers control flow, modules, and user-defined data types, but it is still intentionally small. `async`
functions and `await` execute synchronously, property assignment on objects and instances is not yet supported, and the
standard library only provides `print`. Future iterations aim to grow the builtin ecosystem, add mutation helpers for
structured data, and experiment with asynchronous execution.
