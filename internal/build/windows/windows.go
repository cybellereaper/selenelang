// Package windows contains helpers for building Selene executables on Windows.
package windows

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var stubTemplate = template.Must(template.New("seleneWindowsStub").Parse(`package main

import (
    "encoding/base64"
    "log"
    "strings"

    "github.com/cybellereaper/selenelang/internal/jit"
    "github.com/cybellereaper/selenelang/internal/lexer"
    "github.com/cybellereaper/selenelang/internal/parser"
    "github.com/cybellereaper/selenelang/internal/runtime"
)

const embeddedSourceName = "{{.SourceName}}"
const embeddedProgram = "{{.EncodedSource}}"

func main() {
    data, err := base64.StdEncoding.DecodeString(embeddedProgram)
    if err != nil {
        log.Fatalf("selene jit loader: failed to decode embedded program: %v", err)
    }
    source := string(data)
    lexer := lexer.New(source)
    parser := parser.New(lexer)
    program := parser.ParseProgram()
    if errs := parser.Errors(); len(errs) > 0 {
        log.Fatalf("selene jit loader: parse error in %s:\n%s", embeddedSourceName, strings.Join(errs, "\n"))
    }
    rt := runtime.New()
    compiled, err := jit.Compile(program)
    if err != nil {
        log.Fatalf("selene jit compile error: %v", err)
    }
    if _, err := compiled.Run(rt); err != nil {
        log.Fatalf("selene jit runtime error: %v", err)
    }
}
`))

type stubData struct {
	SourceName    string
	EncodedSource string
}

// BuildExecutable assembles a Windows executable that embeds the Selene program.
func BuildExecutable(startDir, sourceName, sourceCode, output string) error {
	moduleRoot, err := findModuleRoot(startDir)
	if err != nil {
		return fmt.Errorf("windows build: %w", err)
	}
	absOut, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absOut), 0o755); err != nil {
		return err
	}
	workdir, err := os.MkdirTemp(moduleRoot, "selene-win-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workdir)

	data := stubData{
		SourceName:    sourceName,
		EncodedSource: base64.StdEncoding.EncodeToString([]byte(sourceCode)),
	}
	var buf bytes.Buffer
	if err := stubTemplate.Execute(&buf, data); err != nil {
		return err
	}
	mainFile := filepath.Join(workdir, "main.go")
	if err := os.WriteFile(mainFile, buf.Bytes(), 0o644); err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", absOut, mainFile)
	cmd.Dir = moduleRoot
	cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
	buildOutput, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("windows build: go build failed: %v\n%s", err, string(buildOutput))
	}
	return nil
}

func findModuleRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return dir, nil
		} else if os.IsNotExist(statErr) {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
			continue
		} else {
			return "", statErr
		}
	}
	return "", fmt.Errorf("could not locate go.mod from %s", start)
}
