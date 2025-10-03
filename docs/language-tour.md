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

Functions return the value of their last expression. Explicit `return` statements are not yet available.

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

## Optional chaining and Elvis operator

Member lookups can be made optional with `?.`. Combine this with the Elvis operator `?:` to provide defaults when values are
null (or become null after an optional access):

```selene
let config = { name: "Selene", owner: null };
let owner = config.owner?.name ?: "Unknown";
print("owner => " + owner);
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

The runtime currently focuses on expressions. Statements such as `if`, `while`, and pattern matching are parsed but not executed
by the interpreter yet. When designing scripts, structure control flow with function decomposition and Elvis operators. Future
iterations aim to add richer control flow, user-defined modules, and more built-in functions.
