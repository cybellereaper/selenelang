---
layout: default
title: Embedding Selene
---

# Embedding Selene

Selene's lexer, parser, and interpreter are exposed as Go packages. This guide shows how to execute Selene code from your own
applications and extend the runtime with custom functionality.

## Running a script from Go

```go
package main

import (
    "fmt"
    "os"

    "selenelang/internal/lexer"
    "selenelang/internal/parser"
    "selenelang/internal/runtime"
)

func main() {
    script, err := os.ReadFile("examples/fundamentals/hello.selene")
    if err != nil {
        panic(err)
    }

    l := lexer.New(string(script))
    p := parser.New(l)
    program := p.ParseProgram()
    if errs := p.Errors(); len(errs) > 0 {
        panic(errs)
    }

    rt := runtime.New()
    if _, err := rt.Run(program); err != nil {
        panic(err)
    }

    fmt.Println("script executed successfully")
}
```

The runtime evaluates statements sequentially and returns the value of the last expression. Errors produced during parsing or
evaluation bubble back as Go `error` values so you can integrate Selene into larger systems safely.

## Sharing environments

When embedding Selene you can reuse a single runtime to keep global state alive across multiple scripts:

```go
rt := runtime.New()

for _, path := range []string{"config.selene", "startup.selene"} {
    if err := runScript(rt, path); err != nil {
        panic(err)
    }
}
```

Each script receives the same top-level environment, making it easy to build REPLs or plugin systems.

## Adding custom builtins

Expose host functionality by injecting new built-in functions into the runtime environment:

```go
rt := runtime.New()
rt.Environment().Set("now", runtime.NewBuiltin("now", func(args []runtime.Value) (runtime.Value, error) {
    return runtime.NewString(time.Now().Format(time.RFC3339)), nil
}))
```

Selene scripts can now call `now()` to retrieve timestamps from the host application.
Remember to import Go's standard-library `time` package in the host program.

The default runtime already includes a handful of helpers—`print`, `format`, `spawn`, and `channel`. You can freely mix these
with your own builtins to expose logging, metrics, or IO capabilities to scripts.

## Handling results

A Selene program returns the last evaluated value. Use this to send structured data back to Go:

```go
result, err := rt.Run(program)
if err != nil {
    return err
}

fmt.Println("program produced:", result.Inspect())
```

Objects and arrays map cleanly onto Selene's native composite types, making it straightforward to implement serialization or
configuration pipelines.

## Embedding tips

- Use `runtime.Compile` to produce bytecode chunks when you want to validate syntax or inspect instructions before executing via `Runtime.RunChunk`.
- Wrap runtime execution in context-aware cancellation if scripts may run for an extended period.
- Pair Selene with Go's templating or HTTP packages to build dynamic configuration and scripting environments.
