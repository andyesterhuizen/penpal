package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/andrewesterhuizen/penpal/penpal"

	"github.com/andrewesterhuizen/penpal/assembler"
	"github.com/andrewesterhuizen/penpal/midi"
	"github.com/andrewesterhuizen/penpal/vm"
)

func printMidiDevices() {
	midiHandler := midi.NewPortMidiMidiHandler()
	inputs, outputs := midiHandler.GetDevices()

	fmt.Println("inputs:")
	for _, d := range inputs {
		fmt.Printf("[%v] %s\n", d.Id, d.Name)
	}

	fmt.Println("outputs:")
	for _, d := range outputs {
		fmt.Printf("[%v] %s\n", d.Id, d.Name)
	}
}

func compileFromFile(filename string) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	systemIncludes, err := penpal.GetSystemIncludes()
	if err != nil {
		log.Fatal(err)
	}

	a := assembler.New(assembler.Config{SystemIncludes: systemIncludes})

	program, err := a.GetProgram(filename, string(f))
	if err != nil {
		log.Fatal(err)
	}

	binary.Write(os.Stdout, binary.LittleEndian, program)
}

func loadProgramFromFile(filename string) []byte {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	// determine if file is compiled binary by checking header
	header := []byte("PENPAL")
	binary := true

	for i, c := range header {
		if f[i] != c {
			binary = false
			break
		}
	}

	if binary {
		return f
	}

	systemIncludes, err := penpal.GetSystemIncludes()
	if err != nil {
		log.Fatal(err)
	}

	a := assembler.New(assembler.Config{SystemIncludes: systemIncludes})

	program, err := a.GetProgram(filename, string(f))
	if err != nil {
		log.Fatal(err)
	}

	return program
}

func executeProgramFromFile(filename string) {
	program := loadProgramFromFile(filename)

	midiHandler := midi.NewPortMidiMidiHandler()
	defer midiHandler.Close()

	vm := vm.New()

	msPerMinute := 60 * 1000

	// TODO: clock should be enabled according to a flag
	go func() {
		for {
			bpm := vm.GetMemory(penpal.AddressBPM)
			ppqn := vm.GetMemory(penpal.AddressPPQN)

			if bpm == 0 || ppqn == 0 {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			interval := (msPerMinute / int(bpm)) / int(ppqn)

			// TODO: call interupt

			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()

	vm.Load(program)

	for !vm.Halted {
		vm.Tick()

		midiBytes := vm.GetMemorySection(penpal.AddressMidiMessageStart, 4)

		if midiBytes[3] > 0 {
			midiHandler.Send(midiBytes[0], midiBytes[1], midiBytes[2])
			vm.SetMemory(penpal.AddressSendMessage, 0x0)
		}
	}

	// vm.PrintReg()
	// vm.PrintMem(0, 0xf)
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "devices":
			printMidiDevices()
			return

		case "compile":
			if len(args) < 2 {
				log.Fatal("no input file")
			}

			compileFromFile(args[1])
			return

		default:
			executeProgramFromFile(args[0])
		}

		return
	}

	// TODO: print help info if no args supplied
}
