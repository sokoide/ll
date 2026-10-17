package minilang

import (
	"fmt"
	"io"
	"strconv"
)

// Execute runs compiler-produced instructions with a fresh PC, operand stack,
// and scope stack. The VM never visits AST nodes.
func (b *Bytecode) Execute(output io.Writer) error {
	if output == nil {
		return fmt.Errorf("output writer must not be nil")
	}
	v := &vm{code: b.code, scopes: NewScopeStack(), output: output}
	return v.run()
}

type vm struct {
	code   []instruction
	pc     int
	stack  []int
	scopes *ScopeStack
	output io.Writer
}

func (v *vm) pop() (int, error) {
	if len(v.stack) == 0 {
		return 0, fmt.Errorf("operand stack underflow")
	}
	last := len(v.stack) - 1
	value := v.stack[last]
	v.stack = v.stack[:last]
	return value, nil
}

// run is fetch -> advance PC -> execute. A jump replaces the advanced PC.
func (v *vm) run() error {
	for v.pc < len(v.code) {
		address := v.pc
		ins := v.code[v.pc]
		v.pc++
		if err := v.execute(ins); err != nil {
			return fmt.Errorf("pc %d: %w", address, err)
		}
	}
	if len(v.stack) != 0 {
		return fmt.Errorf("unbalanced operand stack: %d values", len(v.stack))
	}
	return nil
}

func (v *vm) execute(ins instruction) error {
	switch ins.op {
	case opPush:
		value, err := strconv.Atoi(ins.text)
		if err != nil {
			return err
		}
		v.stack = append(v.stack, value)
	case opLoad:
		v.stack = append(v.stack, v.scopes.load(ins.text))
	case opStore:
		value, err := v.pop()
		if err != nil {
			return err
		}
		v.scopes.assign(ins.text, value)
	case opBinary:
		rhs, err := v.pop()
		if err != nil {
			return err
		}
		lhs, err := v.pop()
		if err != nil {
			return err
		}
		value, err := applyBinary(ins.arg, lhs, rhs)
		if err != nil {
			return err
		}
		v.stack = append(v.stack, value)
	case opPrintInt:
		value, err := v.pop()
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(v.output, value)
		return err
	case opPrintText:
		_, err := io.WriteString(v.output, ins.text)
		return err
	case opNewline:
		_, err := fmt.Fprintln(v.output)
		return err
	case opEnterScope:
		v.scopes.PushScope()
	case opLeaveScope:
		v.scopes.PopScope()
	case opJump:
		return v.jump(ins.arg)
	case opJumpIfFalse:
		value, err := v.pop()
		if err != nil {
			return err
		}
		if value <= 0 {
			return v.jump(ins.arg)
		}
	default:
		return fmt.Errorf("unknown opcode %d", ins.op)
	}
	return nil
}

func (v *vm) jump(address int) error {
	if address < 0 || address > len(v.code) {
		return fmt.Errorf("invalid jump target %d", address)
	}
	v.pc = address
	return nil
}

// Disassemble writes stable instruction addresses and operands for learning.
func (b *Bytecode) Disassemble(output io.Writer) error {
	for address, ins := range b.code {
		var text string
		switch ins.op {
		case opPush:
			text = "PUSH " + ins.text
		case opLoad:
			text = "LOAD " + ins.text
		case opStore:
			text = "STORE " + ins.text
		case opBinary:
			text = "BINARY " + tokenName(ins.arg)
		case opPrintInt:
			text = "PRINT_INT"
		case opPrintText:
			text = "PRINT_TEXT " + strconv.Quote(ins.text)
		case opNewline:
			text = "NEWLINE"
		case opEnterScope:
			text = "ENTER_SCOPE"
		case opLeaveScope:
			text = "LEAVE_SCOPE"
		case opJump:
			text = fmt.Sprintf("JUMP %d", ins.arg)
		case opJumpIfFalse:
			text = fmt.Sprintf("JUMP_IF_FALSE %d", ins.arg)
		default:
			return fmt.Errorf("unknown opcode %d", ins.op)
		}
		if _, err := fmt.Fprintf(output, "%04d %s\n", address, text); err != nil {
			return err
		}
	}
	return nil
}
