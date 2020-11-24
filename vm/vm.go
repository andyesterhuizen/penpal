package vm

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/andrewesterhuizen/penpal/instructions"
	"github.com/andrewesterhuizen/penpal/midi"
)

const memorySize = 0xffff

const midiMessageStartLocation = 0x0

type VM struct {
	ip           uint16
	sp           uint16
	fp           uint16
	a            uint8
	b            uint8
	memory       [memorySize]uint8
	instructions []uint8
	midi         midi.MidiHandler
}

func New() VM {
	rand.Seed(time.Now().UnixNano())

	vm := VM{midi: midi.NewPortMidiMidiHandler()}
	vm.ip = 0
	vm.sp = memorySize - 1
	vm.fp = memorySize - 1

	return vm
}

func (vm *VM) push(value uint8) {
	vm.memory[vm.sp] = value
	vm.sp--
}

func (vm *VM) pop() uint8 {
	vm.sp++
	return vm.memory[vm.sp]
}

func (vm *VM) push16(n uint16) {
	h := uint8((n & 0xff00) >> 8)
	l := uint8(n & 0xff)

	vm.push(l)
	vm.push(h)
}

func (vm *VM) pop16() uint16 {
	var h, l uint16
	h = uint16(vm.pop())
	l = uint16(vm.pop())
	return (h << 8) | l
}

func (vm *VM) fetch() uint8 {
	vm.ip++
	return vm.instructions[vm.ip]
}

func (vm *VM) fetch16() uint16 {
	vm.ip++
	h := uint16(vm.instructions[vm.ip])
	vm.ip++
	l := uint16(vm.instructions[vm.ip])
	return (h << 8) | l
}

func (vm *VM) execute(instruction uint8) {
	name, exists := instructions.Names[instruction]
	if !exists {
		log.Fatalf("encountered unknown instruction 0x%02x", instruction)
	}

	fmt.Printf("Executing %s\n", name)

	switch instruction {
	case instructions.SWAP:
		vm.a, vm.b = vm.b, vm.a
		vm.ip++

	case instructions.MOV:
		flags := vm.fetch()

		mode, dest := instructions.DecodeFlags(flags)

		var destAddress *uint8

		switch dest {
		case instructions.DestRegisterA:
			destAddress = &vm.a
		case instructions.DestRegisterB:
			destAddress = &vm.b
		default:
			log.Fatalf("encountered unknown destination for MOV, 0x%02x", dest)
		}

		switch mode {
		case instructions.AddressingModeImmediate:
			*destAddress = vm.fetch()
		case instructions.AddressingModeFPRelative:
			offset := int8(vm.fetch())
			addr := vm.fp

			// is there an easier way to do this?
			if offset >= 0 {
				addr += uint16(offset)
			} else {
				offset *= -1
				addr -= uint16(offset)
			}

			v := vm.memory[addr]
			*destAddress = v
		default:
			log.Fatalf("encountered unknown addressing mode for MOV, 0x%02x", mode)
		}

		vm.ip++

	case instructions.STORE:
		addr := vm.fetch16()
		vm.memory[addr] = vm.a
		vm.ip++

	case instructions.LOAD:
		addr := vm.fetch16()
		vm.a = vm.memory[addr]
		vm.ip++

	case instructions.ADD:
		vm.a += vm.b
		vm.ip++

	case instructions.SUB:
		vm.a -= vm.b
		vm.ip++

	case instructions.MUL:
		vm.a *= vm.b
		vm.ip++

	case instructions.DIV:
		vm.a /= vm.b
		vm.ip++

	case instructions.SHL:
		vm.a = vm.a << vm.b
		vm.ip++

	case instructions.SHR:
		vm.a = vm.a >> vm.b
		vm.ip++

	case instructions.AND:
		vm.a = vm.a & vm.b
		vm.ip++

	case instructions.OR:
		vm.a = vm.a | vm.b
		vm.ip++

	case instructions.JUMP:
		addr := vm.fetch16()
		vm.ip = addr

	case instructions.JUMPZ:
		addr := vm.fetch16()
		if vm.a == 0 {
			vm.ip = addr
		} else {
			vm.ip++
		}

	case instructions.JUMPNZ:
		addr := vm.fetch16()
		if vm.a != 0 {
			vm.ip = addr
		} else {
			vm.ip++
		}

	case instructions.PUSH:
		vm.push(vm.a)
		vm.ip++

	case instructions.POP:
		vm.a = vm.pop()
		vm.ip++

	case instructions.CALL:
		addr := vm.fetch16()

		// save state
		// vm.push(vm.a)
		vm.push(vm.b)
		vm.push16(vm.ip + 1) // return address
		// vm.push(4 + 1)       // frame size, including this byte

		vm.fp = vm.sp // save frame pointer
		vm.ip = addr  // set ip to called address

	case instructions.RET:
		vm.sp = vm.fp

		// restore state
		// frameSize := vm.pop()
		addr := vm.pop16()
		vm.b = vm.pop()
		// vm.a = vm.pop()

		// remove args from stack
		nArgs := vm.pop()
		for i := 0; i < int(nArgs); i++ {
			vm.pop()
		}

		vm.fp += uint16(4) // frame size is always 4 but this could change
		vm.ip = addr

	case instructions.SEND:
		status := vm.memory[midiMessageStartLocation]
		data1 := vm.memory[midiMessageStartLocation+1]
		data2 := vm.memory[midiMessageStartLocation+2]

		vm.midi.Send(status, data1, data2)
		vm.ip++

	case instructions.RAND:
		vm.a = uint8(rand.Intn(255))
		vm.ip++

	default:
		log.Fatalf("encountered unknown instruction 0x%02x, name=%s", instruction, instructions.Names[instruction])
	}

}

func (vm *VM) Load(instructions []uint8) {
	vm.instructions = instructions
}

func (vm *VM) PrintReg() {
	fmt.Printf("a: 0x%02x | b: 0x%02x\n", vm.a, vm.b)
}

func (vm *VM) PrintMem(start uint16, n uint16) {
	for i := start; i < start+n; i++ {
		fmt.Printf("%04x: 0x%02x\n", i, vm.memory[i])
	}
}

func (vm *VM) Run() {
	ins := vm.instructions

	for {
		if ins[vm.ip] == instructions.HALT {
			break
		}

		vm.execute(ins[vm.ip])
	}
}

func (vm *VM) Close() {
	vm.midi.Close()
}
