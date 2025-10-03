---
layout: default
title: Language Reference
---

# Selene language reference

This reference catalogs the syntax recognized by the Selene lexer and parser, along with notes on what the runtime executes today. Use it to cross-check constructs while writing scripts or building tooling on top of the Selene frontend.

## Lexical structure

- **Whitespace** – spaces, tabs, and newlines separate tokens but are otherwise ignored.
- **Comments** – `//` starts a line comment that runs to the end of the line. `/* ... */` forms a block comment and may span multiple lines. Comments are skipped by the lexer and do not nest.
- **Identifiers** – start with an ASCII letter or underscore and may contain ASCII letters, digits, or underscores. Identifiers are case-sensitive.
- **Keywords** – `let`, `var`, `fn`, `async`, `contract`, `returns`, `class`, `struct`, `enum`, `match`, `module`, `import`, `as`, `if`, `else`, `while`, `for`, `return`, `break`, `continue`, `true`, `false`, `null`, `is`, and `await`.
- **Operators** – arithmetic (`+`, `-`, `*`, `/`, `%`), comparison (`==`, `!=`, `<`, `<=`, `>`, `>=`), logical (`&&`, `||`, `!`), assignment (`=`), Elvis (`?:`), member access (`.`), optional chaining (`?.`), non-null assertion (`!!`), indexing (`[]`), call application (`()`), and the pattern arrow (`=>`).

## Literals

- **Numbers** – sequences of digits optionally containing a single decimal point (e.g. `42`, `3.14`). All numbers are stored as 64-bit floating point values at runtime.
- **Strings** – delimited by double quotes and supporting standard escape sequences (e.g. `"\n"`).
- **Booleans** – the keywords `true` and `false`.
- **Null** – represented by the keyword `null`.
- **Arrays** – bracketed collections such as `[1, 2, 3]`. Elements evaluate from left to right.
- **Objects** – brace-delimited key/value maps like `{ name: "Selene", version: "0.1.0" }`. Keys may be bare identifiers or string literals.

## Declarations

### Variable bindings

- `let name = expression;` creates an immutable binding.
- `var name = expression;` creates a mutable binding that may be reassigned later with `name = newValue`.
- Optional type annotations may appear after the identifier: `let count: Number = 3;`. Types are parsed but not enforced by the runtime yet.

### Functions

```
fn name(parameters) [ : ReturnType ] [async] [contract { ... }] { ... }
fn name(parameters) [ : ReturnType ] [async] [contract { ... }] => expression;
```

- Parameters are comma-separated identifiers that may also include type annotations (parsed only).
- Bodies can be expression-bodied (`=>`) or block-bodied (`{ ... }`). Expression bodies implicitly become the function result.
- Functions close over the lexical environment in which they are defined.
- `async` marks a function for future asynchronous execution but currently runs synchronously. Contract clauses are enforced
  after the function body completes: each `returns(condition)` entry evaluates the optional guard/condition with a `result`
  binding that contains the function's return value. A falsy condition triggers a runtime error.

### Modules and imports

```
module Identifier {
    // nested declarations and statements
}

import foo.bar as baz;
```

- Modules execute in their own lexical scope and export every binding defined in the body. `import` pulls values from a module (optionally drilling into nested properties) and binds them into the current scope, with `as` supplying an alias.

### Classes, structs, enums, and contracts

```
class Name(params) [ : Super ] { ... }
struct Name(params) { ... }
enum Name { CaseOne; CaseTwo(value); }
contract Name { ... }
```

- Struct and class declarations introduce callable constructors. Struct instances expose their fields and methods via `self`. Class declarations optionally inherit methods and static values from a superclass. Enum declarations generate constructor functions for each case that produce tagged union values. Contract declarations produce reusable bundles of declarations that can be accessed with dot syntax (similar to modules), and inline function contracts are enforced when the function returns.

## Statements

- **Expression statement** – any expression followed by an optional semicolon. The value of the expression becomes the statement result.
- **Block** – `{ statement* }` introduces a new lexical scope and returns the value of the last statement inside the block.
- **If statement** – `if condition { ... } [else statement]` executes the first branch whose condition is truthy. `else if` chains are written as `else` followed by another `if` statement.
- **While loop** – `while condition { ... }` repeats the body while the condition evaluates to a truthy value.
- **For loop** – `for (initializer; condition; post) { ... }` executes the initializer once, evaluates the condition before each iteration, and runs the post expression after each iteration.
- **Match statement** – `match expression { pattern => statement; ... }` evaluates the target expression, tries each pattern in order, and executes the body of the first successful match. The value produced by the body becomes the statement result. If no patterns match, the statement yields `null`.
- **Return statement** – `return expression?;` exits the innermost function. Without an expression the function returns `null`.
- **Break/continue** – `break;` exits the nearest loop; `continue;` skips directly to the next iteration.

## Patterns

Selene patterns appear inside `match` statements.

- **Identifier pattern** – `name` binds the matched value to a fresh identifier within the case body.
- **Literal pattern** – any literal expression (`0`, `"text"`, `true`, `null`, etc.) that matches by value.
- **Object pattern** – `{ key: subpattern, other: anotherPattern }` destructures object properties recursively. Keys may use identifier or string syntax.
- **Struct pattern** – `Point(x, y)` matches struct and class instances created by `Point` and binds their positional fields. It also matches enum cases with the same name, binding the case parameters.

## Expressions

- **Primary expressions** – identifiers, literals, array/object literals, and grouped expressions `( ... )`.
- **Unary operators** – `-expr`, `+expr`, and `!expr`.
- **Binary operators** – addition, subtraction, multiplication, division, modulo, comparisons, equality, logical `&&`/`||`, and Elvis `?:`.
- **Assignments** – `name = expression` updates an existing binding created with `var`.
- **Function calls** – `callee(arg1, arg2)` evaluate the callee then its arguments left to right.
- **Indexing** – `array[index]` or `string[index]`.
- **Member access** – `object.property`, optional chaining `object?.property`, and non-null assertions `expression!!`. Arrays and strings expose a read-only `length` property.
- **Await expression** – the `await` keyword parses and simply evaluates its operand. `async` functions currently execute synchronously.

## Runtime support summary

- The interpreter executes:

- Variable declarations, assignments, and lexical scoping.
- Modules, imports, and nested scopes.
- Function declarations, first-class closures, expression/block bodies, and inline contracts.
- Struct, class, and enum constructors with instance methods.
- Arrays, objects, arithmetic, comparisons, logical operators, Elvis expressions, optional chaining, and non-null assertions.
- Control flow including `if`/`else`, `for`, `while`, `return`, `break`, and `continue`.
- Match statements with identifier, literal, object, and struct/enum patterns.
- The built-in `print` function for console output.

`async` functions and `await` expressions are currently synchronous conveniences; no concurrent runtime is provided yet.

Refer to the [example scripts](examples.md) for runnable demonstrations of the supported features.
