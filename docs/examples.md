---
layout: default
title: Example Scripts
---

# Example scripts

Selene ships with runnable programs that exercise the language surface. Execute any of them with `selene path/to/script.selene` or run the whole suite through `selene test`.

## Guided tour

- **`examples/language_tour.selene`** – a single script that touches modules, structs, enums, condition blocks, loops, and reusable helper functions.

## Fundamentals

- **`examples/hello.selene`** – immutable bindings, expression-bodied functions, and the `format` helper.
- **`examples/math.selene`** – modules exporting arithmetic helpers and calling into imported members.
- **`examples/collections.selene`** – objects, arrays, optional chaining, Elvis defaults, and indexing.
- **`examples/options.selene`** – navigating nested optionals, non-null assertions, and handling missing data.
- **`examples/control.selene`** – classic `for`/`while` loops, `if` branches, `continue`/`break`, and recursion.
- **`examples/strings.selene`** – string interpolation, raw literals, format specifiers, and string extension methods.

## Data types and patterns

- **`examples/types.selene`** – struct/class/enum declarations plus pattern matching on enum cases.
- **`examples/patterns.selene`** – destructuring records with `match` expressions.
- **`examples/contracts.selene`** – reusable contract declarations and postconditions that validate results.
- **`examples/interfaces.selene`** – structural interfaces implemented by structs and extension functions.
- **`examples/pointers.selene`** – address-of/dereference operations and swapping values by reference.

## Modularity and packages

- **`examples/modules.selene`** – declaring modules and importing their exported members.
- **`examples/packages.selene`** – package headers plus augmented assignment operators in action.
- **`examples/dependency.selene`** – importing a vendored dependency referenced in `selene.lock`.

## Execution backends and tooling

- **`examples/vm.selene`** – tight loops and accumulation logic that are perfect for VM disassembly experiments.
- **`examples/extensions.selene`** – extension methods for strings and integers.
- **`examples/recursion.selene`** – recursive factorial and Fibonacci implementations.

## Advanced flow and runtime features

- **`examples/concurrency.selene`** – spawning tasks, communicating over channels, and awaiting futures.
- **`examples/errors.selene`** – resource-safe `using` blocks with `try`/`catch`/`finally` and explicit `throw`.
- **`examples/conditions.selene`** – rule-based `condition` blocks for declarative branching.

Explore the scripts, tweak them, and run `selene test --verbose` to see the console output produced by each example.
