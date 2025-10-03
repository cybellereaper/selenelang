---
layout: default
title: Example Showcase
---

# Example showcase

Jump into any of the curated scripts below to see Selene's features in motion. Each section links directly to the reorganized folders under `examples/` so you can copy, tweak, and rerun them instantly.

<div align="center">
  <img src="https://media.giphy.com/media/v1.Y2lkPTc5MGI3NjExZ2d2enAydmJya3l2Y3g5OGpwM3hoYXg4Ymh6NGJlNmRwNTZwNHlwMSZlcD12MV9naWZzX3NlYXJjaCZjdD1n/11XgUdoQtx2rTy/giphy.gif" width="280" alt="Rocket ship launching animation" />
</div>

## Guided tour

- **`examples/showcase/language_tour.selene`** – one epic script that touches modules, structs, enums, condition blocks, loops, and reusable helper functions.

## Fundamentals

- **`examples/fundamentals/hello.selene`** – immutable bindings, expression-bodied functions, and the `format` helper.
- **`examples/fundamentals/math.selene`** – modules exporting arithmetic helpers and calling into imported members.
- **`examples/fundamentals/collections.selene`** – objects, arrays, optional chaining, Elvis defaults, and indexing.
- **`examples/fundamentals/options.selene`** – navigating nested optionals, non-null assertions, and handling missing data.
- **`examples/fundamentals/control.selene`** – classic `for`/`while` loops, `if` branches, `continue`/`break`, and recursion.
- **`examples/fundamentals/strings.selene`** – string interpolation, raw literals, format specifiers, and string extension methods.

## Data types and patterns

- **`examples/types-patterns/types.selene`** – struct/class/enum declarations plus pattern matching on enum cases.
- **`examples/types-patterns/patterns.selene`** – destructuring records with `match` expressions.
- **`examples/types-patterns/contracts.selene`** – reusable contract declarations and postconditions that validate results.
- **`examples/types-patterns/interfaces.selene`** – structural interfaces implemented by structs and extension functions.
- **`examples/types-patterns/pointers.selene`** – address-of/dereference operations and swapping values by reference.

## Modularity and packages

- **`examples/modularity/modules.selene`** – declaring modules and importing their exported members.
- **`examples/modularity/packages.selene`** – package headers plus augmented assignment operators in action.
- **`examples/modularity/dependency.selene`** – importing a vendored dependency referenced in `selene.lock`.

## Execution backends and tooling

- **`examples/tooling/vm.selene`** – tight loops and accumulation logic that are perfect for VM disassembly experiments.
- **`examples/tooling/extensions.selene`** – extension methods for strings and integers.
- **`examples/tooling/recursion.selene`** – recursive factorial and Fibonacci implementations.

## Advanced flow and runtime features

- **`examples/runtime/concurrency.selene`** – spawning tasks, communicating over channels, and awaiting futures.
- **`examples/runtime/errors.selene`** – resource-safe `using` blocks with `try`/`catch`/`finally` and explicit `throw`.
- **`examples/runtime/conditions.selene`** – rule-based `condition` blocks for declarative branching.

Explore the scripts, tweak them, and run `selene test --mode all --verbose` to see the console output produced by each example across the interpreter, VM, and JIT backends.
