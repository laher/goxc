package exefileparse

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"errors"
	"log"
	"os"

	"github.com/laher/goxc/platforms"
)

//I think plan9 uses a plain old a.out file format
var (
	MAGIC_PLAN9_386 = []byte{0, 0, 1, 235}
)

func Test(filename, expectedArch, expectedOs string, isVerbose bool) error {
	switch expectedOs {
	case platforms.WINDOWS:
		return TestPE(filename, expectedArch, expectedOs, isVerbose)
	case platforms.DARWIN:
		return TestMachO(filename, expectedArch, expectedOs, isVerbose)
	case platforms.PLAN9:
		return TestPlan9Exe(filename, expectedArch, expectedOs, isVerbose)
	default:
		return TestElf(filename, expectedArch, expectedOs, isVerbose)
	}

}

func TestElf(filename, expectedArch, expectedOs string, isVerbose bool) error {
	file, err := elf.Open(filename)

	if err != nil {
		log.Printf("File '%s' is not an ELF file: %v\n", filename, err)
		return err
	}
	defer file.Close()
	if isVerbose {
		log.Printf("File '%s' is an ELF file (arch: %s, osabi: %s)\n", filename, file.FileHeader.Machine.String(), file.FileHeader.OSABI.String())
	}
	if expectedOs == platforms.LINUX {
		if file.FileHeader.OSABI != elf.ELFOSABI_NONE && file.FileHeader.OSABI != elf.ELFOSABI_LINUX {
			return errors.New("Not a Linux executable")
		}
	}
	if expectedOs == platforms.NETBSD {
		if file.FileHeader.OSABI != elf.ELFOSABI_NETBSD {
			return errors.New("Not a NetBSD executable")
		}
	}
	if expectedOs == platforms.FREEBSD {
		if file.FileHeader.OSABI != elf.ELFOSABI_FREEBSD {
			return errors.New("Not a FreeBSD executable")
		}
	}
	if expectedOs == platforms.OPENBSD {
		if file.FileHeader.OSABI != elf.ELFOSABI_OPENBSD {
			return errors.New("Not an OpenBSD executable")
		}
	}

	if expectedArch == platforms.ARM {
		if file.FileHeader.Machine != elf.EM_ARM {
			return errors.New("Not an ARM executable")
		}
	}
	if expectedArch == platforms.X86 {
		if file.FileHeader.Machine != elf.EM_386 {
			return errors.New("Not a 386 executable")
		}

	}
	if expectedArch == platforms.AMD64 {
		if file.FileHeader.Machine != elf.EM_X86_64 {
			return errors.New("Not an AMD64 executable")
		}

	}
	return nil
}

func TestMachO(filename, expectedArch, expectedOs string, isVerbose bool) error {
	file, err := macho.Open(filename)
	if err != nil {

		log.Printf("File '%s' is not a Mach-O file: %v\n", filename, err)
		return err
	}
	defer file.Close()
	if isVerbose {
		log.Printf("File '%s' is a Mach-O file (arch: %s)\n", filename, file.FileHeader.Cpu.String())
	}
	if expectedArch == platforms.X86 {
		if file.FileHeader.Cpu != macho.Cpu386 {
			return errors.New("Not a 386 executable")
		}

	}
	if expectedArch == platforms.AMD64 {
		if file.FileHeader.Cpu != macho.CpuAmd64 {
			return errors.New("Not an AMD64 executable")
		}

	}
	return nil
}

func TestPE(filename, expectedArch, expectedOs string, isVerbose bool) error {
	file, err := pe.Open(filename)
	if err != nil {
		return errors.New("NOT a PE file")
	}
	defer file.Close()
	if isVerbose {
		log.Printf("File '%s' is a PE file, arch: %d (%d='X86' and %d='AMD64')\n", filename, file.FileHeader.Machine, pe.IMAGE_FILE_MACHINE_I386, pe.IMAGE_FILE_MACHINE_AMD64)
	}
	if expectedArch == platforms.X86 {
		if file.FileHeader.Machine != pe.IMAGE_FILE_MACHINE_I386 {
			return errors.New("Not a 386 executable")
		}

	}
	if expectedArch == platforms.AMD64 {
		if file.FileHeader.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
			return errors.New("Not an AMD64 executable")
		}

	}
	return nil

}

func TestPlan9Exe(filename, expectedArch, expectedOs string, isVerbose bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return errors.New("Could not open file")
	}
	defer file.Close()
	b := []byte{0, 0, 0, 0}
	i, err := file.Read(b)
	if err != nil || i < 4 {
		return errors.New("Could not read first 2 bytes of file")
	}

	if expectedArch == platforms.X86 {
		for i := range b {
			if b[i] != MAGIC_PLAN9_386[i] {
				return errors.New("NOT a known Plan9 executable format")
			}
		}
		if isVerbose {
			log.Printf("File '%s' is a Plan9 executable", filename)
		}
	}
	return nil

}
