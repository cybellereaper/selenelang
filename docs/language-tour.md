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

## Strings and formatting

Selene offers several string literal forms:

- Standard quoted strings support escapes and interpolation with `${ ... }`.
- Format strings start with `f"..."` and accept inline format specifiers like `${value | upper}`.
- Triple-quoted strings (`""" ... """`) preserve indentation and newlines.
- Raw strings are prefixed with `r"..."` or `r"""..."""` and do not treat escapes specially.

```selene
let name = "Selene";
let standard = "Hello, ${name}!";
let formatted = f"{name | upper} => ${3.14159 | %.2f}";
let multiline = """
Line one
Line two with ${name}
""";
let raw = r"C:\\Users\\demo";

print(standard);
print(formatted);
print(multiline);
print(raw);
```

`f""` format specifiers understand transformations such as `upper`, `lower`, `title`, `trim`, and printf-style numeric codes.

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

Selene offers familiar imperative control flow. Use `if` for branching and `for`/`while` for iteration. `return`, `break`, and `continue` behave as you would expect from other languages. Mutable bindings can take advantage of the augmented assignment operators (`+=`, `-=`, `*=`, `/=`, `%=`) to update values concisely:

```selene
var total = 0;

for (let i = 0; i < 5; i = i + 1) {
    if i % 2 == 0 {
        continue;
    }
    total += i;
}

while total > 4 {
    total -= 1;
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

## Extension functions

Use `ext fn` to add behavior to existing types without modifying their original declarations. Extension methods receive the
receiver as `this` inside the body and can leverage other helpers in scope:

```selene
ext fn String.reverse(): String = this.chars().reverse().join("");

ext fn String.isPalindrome(): Boolean = this == this.reverse();

ext fn Int.squared(): Int = this * this;

print("level".isPalindrome());
print(9.squared());
```

## Pointers and references

The runtime models pointers explicitly. Use `&` to capture the address of an identifier in scope and `*` to read or assign
through a pointer. This enables by-reference helpers without introducing classes:

```selene
var left = 1;
var right = 2;

fn swap(a: Pointer, b: Pointer) {
    let temp = *a;
    *a = *b;
    *b = temp;
}

swap(&left, &right);
print("left => " + left + ", right => " + right);
```

Pointers can only be created for identifiers that exist in the current environment, preventing dangling references.

## Interfaces and type checks

Interfaces describe structural requirements. A value conforms when it supplies the listed members. Use `is`/`!is` to test
conformance at runtime:

```selene
interface Greets {
    fn greet(name: String): String;
}

class Robot(id: String) {
    fn greet(name: String): String {
        return f"#{id} greets ${name}";
    }
}

let bot = Robot("rx-1");
print("is interface => " + (bot is Greets));
if bot is Greets {
    print(bot.greet("Selene"));
}
```

Extension functions may also satisfy interfaces by contributing the required members to existing types.

## Resource safety with `using`

`using` ensures resources are closed automatically. The target expression must expose a `close()` function (or Go-style `Close()`
method when embedding):

```selene
struct Trace(name: String) {
    fn close() {
        print("closing", self.name);
    }
}

fn open(name: String): Trace {
    print("opening", name);
    return Trace(name);
}

using scope = open("session.log") {
    print("work inside using");
}
```

When execution leaves the block—because of a return, throw, or normal completion—the resource is closed exactly once.

## Error handling

Handle failure paths explicitly with `try`/`catch`/`finally` and `throw`. The runtime propagates errors until a matching `catch`
intercepts them. `finally` always runs, even if a catch rethrows a new error:

```selene
fn parseNumber(text: String): Number {
    try {
        return text.toNumber();
    } catch (err) {
        print("failed: " + err.message);
        throw err;
    } finally {
        print("parse attempt complete");
    }
}
```

Wrap the identifier in parentheses to bind the thrown value inside a `catch` clause.

## Concurrency primitives

`spawn` launches a function asynchronously and returns a task handle. Create channels with `channel()` and coordinate producers
and consumers using `send`/`recv`. Await the result of a spawned task or the next value from a channel with `await`:

```selene
fn produce(out: Channel) {
    var i = 0;
    while i < 3 {
        out.send(f"tick #${i}");
        i += 1;
    }
    out.close();
}

let ch = channel();
let worker = spawn(produce, ch);

try {
    while true {
        print(await ch);
    }
} catch (err) {
    print("channel closed");
}

await worker;
```

Channels signal completion by raising an error from `recv()` (and therefore `await channel`) when closed, making them easy to
integrate with `try` blocks.

## Condition dispatch

`condition` blocks offer rule-based, object-oriented dispatch. Each `when` guard checks a predicate; the first truthy guard runs
its body. An optional `else` clause handles the fallback:

```selene
let order = { total: 120 };

condition {
    when order.total > 100 => print("VIP shipping");
    when order.total > 0 => print("Standard shipping");
    else => print("Empty order");
}
```

This pattern keeps complex branching logic declarative and readable. Guards execute in the surrounding scope, so any identifier
referenced within `when` clauses must already be defined.

## Packages, modules, and imports

Start a script with an optional `package` header to label its namespace. Group related helpers inside a `module` and import them elsewhere in the file. Everything defined inside the module body is exported automatically, and you can import exports either by identifier path or via Go-style string syntax:

```selene
package analytics;

module math_utils {
    fn square(n: Number): Number => n * n;
    let constants = { tau: 6.28318 };
}

import math_utils.square;
import utils "math_utils";

print(square(5));
print("tau => " + utils.constants.tau);
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

Selene remains intentionally small: numbers are all 64-bit floats, property assignment on objects and instances is still
restricted, and the standard library only includes a handful of helpers (`print`, `format`, `spawn`, `channel`, etc.). The
concurrency runtime is lightweight and best suited for demos rather than high-performance workloads. Future iterations may grow
the builtin ecosystem, expand mutation helpers for structured data, and optimize the interpreter pipeline.
