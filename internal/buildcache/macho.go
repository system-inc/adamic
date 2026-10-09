package buildcache

import (
	"bytes"
	"debug/macho"
	"encoding/binary"
	"fmt"
)

// auditBytes removes only the linker-variable UUID and the signature over it, in memory. Neither fetched nor
// rebuilt executables are changed on disk: a Mach-O without its UUID cannot run. Store blob hashes remain raw.
func auditBytes(content []byte) ([]byte, error) {
	if len(content) < 4 {
		return content, nil
	}
	magic := binary.BigEndian.Uint32(content)
	if magic != macho.Magic32 && magic != macho.Magic64 && magic != 0xcefaedfe && magic != 0xcffaedfe {
		return content, nil
	}
	file, err := macho.NewFile(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	normalized := bytes.Clone(content)
	offset := uint64(28)
	if file.Magic == macho.Magic64 {
		offset = 32
	}
	zero := func(start, size uint64) error {
		if start > uint64(len(normalized)) || size > uint64(len(normalized))-start {
			return fmt.Errorf("Mach-O normalization range %d+%d exceeds file size %d", start, size, len(normalized))
		}
		clear(normalized[start : start+size])
		return nil
	}
	for _, load := range file.Loads {
		command := load.Raw()
		if len(command) < 8 {
			return nil, fmt.Errorf("short Mach-O load command")
		}
		switch file.ByteOrder.Uint32(command) {
		case 0x1b: // LC_UUID: command header, then exactly sixteen UUID bytes.
			if len(command) != 24 {
				return nil, fmt.Errorf("invalid LC_UUID size %d", len(command))
			}
			if err := zero(offset+8, 16); err != nil {
				return nil, err
			}
		case 0x1d: // LC_CODE_SIGNATURE: linkedit_data_command's dataoff and datasize.
			if len(command) != 16 {
				return nil, fmt.Errorf("invalid LC_CODE_SIGNATURE size %d", len(command))
			}
			start, size := file.ByteOrder.Uint32(command[8:12]), file.ByteOrder.Uint32(command[12:16])
			if err := zero(uint64(start), uint64(size)); err != nil {
				return nil, err
			}
		}
		offset += uint64(len(command))
	}
	return normalized, nil
}
