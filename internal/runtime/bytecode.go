package runtime

import (
	"encoding/binary"
	"fmt"
	"slices"
	"strings"

	"github.com/cybellereaper/selenelang/internal/ast"
)

// OpCode represents a single virtual machine instruction.
type OpCode byte

const (
	// OpEvalItem evaluates a compiled program item via the standard runtime.
	OpEvalItem OpCode = iota
	// OpReturn terminates execution and yields the last computed value.
	OpReturn
)

// Chunk contains bytecode generated from a Selene program.
type Chunk struct {
	code  []byte
	items []ast.ProgramItem
}

// Instructions returns the raw bytecode instructions.
func (c *Chunk) Instructions() []byte {
	return slices.Clone(c.code)
}

// ItemCount reports the number of program items referenced by the chunk.
func (c *Chunk) ItemCount() int {
	return len(c.items)
}

// ProgramItem retrieves the program item stored at the given index.
func (c *Chunk) ProgramItem(index int) (ast.ProgramItem, bool) {
	if index < 0 || index >= len(c.items) {
		return nil, false
	}
	return c.items[index], true
}

// Disassemble renders a human-readable listing of the chunk.
func (c *Chunk) Disassemble() string {
	var b strings.Builder
	ip := 0
	for ip < len(c.code) {
		op := OpCode(c.code[ip])
		switch op {
		case OpEvalItem:
			if ip+2 >= len(c.code) {
				fmt.Fprintf(&b, "%04d ERROR truncated OpEvalItem\n", ip)
				return b.String()
			}
			index := int(binary.BigEndian.Uint16(c.code[ip+1 : ip+3]))
			fmt.Fprintf(&b, "%04d OpEvalItem %d", ip, index)
			if item, ok := c.ProgramItem(index); ok {
				fmt.Fprintf(&b, " (%T)\n", item)
			} else {
				b.WriteString(" <missing>\n")
			}
			ip += 3
		case OpReturn:
			fmt.Fprintf(&b, "%04d OpReturn\n", ip)
			ip++
		default:
			fmt.Fprintf(&b, "%04d UNKNOWN %d\n", ip, op)
			ip++
		}
	}
	return b.String()
}

func (c *Chunk) addItem(item ast.ProgramItem) int {
	c.items = append(c.items, item)
	return len(c.items) - 1
}

func (c *Chunk) writeOp(op OpCode) {
	c.code = append(c.code, byte(op))
}

func (c *Chunk) writeUint16(value uint16) {
	c.code = append(c.code, byte(value>>8), byte(value))
}

type compiler struct {
	chunk *Chunk
}

func newCompiler() *compiler {
	return &compiler{chunk: &Chunk{}}
}

func (c *compiler) compile(program *ast.Program) (*Chunk, error) {
	for _, item := range program.Items {
		c.emitEvalItem(item)
	}
	c.chunk.writeOp(OpReturn)
	return c.chunk, nil
}

func (c *compiler) emitEvalItem(item ast.ProgramItem) {
	index := c.chunk.addItem(item)
	c.chunk.writeOp(OpEvalItem)
	c.chunk.writeUint16(uint16(index))
}

// Compile converts a parsed program into bytecode that can be executed by the Selene VM.
func (r *Runtime) Compile(program *ast.Program) (*Chunk, error) {
	comp := newCompiler()
	return comp.compile(program)
}

// RunChunk executes compiled bytecode within the runtime's environment.
func (r *Runtime) RunChunk(chunk *Chunk) (Value, error) {
	vm := &vm{chunk: chunk, env: r.env}
	result, err := vm.run()
	if err != nil {
		return nil, err
	}
	program := &ast.Program{Items: chunk.items}
	analysis := AnalyzeMain(program)
	return InvokeMainIfNeeded(r.env, analysis, result)
}

type vm struct {
	chunk *Chunk
	env   *Environment
	ip    int
}

func (v *vm) run() (Value, error) {
	var last Value = NullValue
	for v.ip < len(v.chunk.code) {
		op := OpCode(v.chunk.code[v.ip])
		v.ip++
		switch op {
		case OpEvalItem:
			if v.ip+1 >= len(v.chunk.code) {
				return nil, fmt.Errorf("truncated OpEvalItem at %d", v.ip-1)
			}
			index := int(binary.BigEndian.Uint16(v.chunk.code[v.ip : v.ip+2]))
			v.ip += 2
			item, ok := v.chunk.ProgramItem(index)
			if !ok {
				return nil, fmt.Errorf("bytecode references missing program item %d", index)
			}
			val, err := evalProgramItem(item, v.env)
			if err != nil {
				return nil, err
			}
			last = val
		case OpReturn:
			return last, nil
		default:
			return nil, fmt.Errorf("unknown opcode %d", op)
		}
	}
	return last, nil
}
