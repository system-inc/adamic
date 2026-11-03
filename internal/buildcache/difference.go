package buildcache

import (
	"bytes"
	"debug/elf"
	"debug/macho"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// firstDifference names where two copies of one product first differ: the first file, by product-relative path in
// sorted order, that one copy holds and the other doesn't, or that the two hold differently, and where in it. In an ELF
// or Mach-O file that is the first section, in the stored copy's order, whose bytes differ (a path baked into a binary
// lands in .rodata or .debug_str, and its length moves the header's offsets after it, so the first differing byte of
// the whole file is often the header's, which says nothing). In any other file it is the first differing byte's offset.
// Two copies alike read as "", so a caller only asks after their digests differed. It reads; it changes nothing.
func firstDifference(stored, rebuilt string) (string, error) {
	storedFiles, err := differenceEntries(stored)
	if err != nil {
		return "", err
	}
	rebuiltFiles, err := differenceEntries(rebuilt)
	if err != nil {
		return "", err
	}
	var paths []string
	for path := range storedFiles {
		paths = append(paths, path)
	}
	for path := range rebuiltFiles {
		if _, both := storedFiles[path]; !both {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		was, inStored := storedFiles[path]
		now, inRebuilt := rebuiltFiles[path]
		switch {
		case !inRebuilt:
			return path + " (only in the stored copy)", nil
		case !inStored:
			return path + " (only in the rebuilt copy)", nil
		case was.kind != now.kind:
			return fmt.Sprintf("%s (a %s, rebuilt as a %s)", path, was.kind, now.kind), nil
		case was.kind == "link" && was.target != now.target:
			return fmt.Sprintf("%s (links to %s, rebuilt to %s)", path, was.target, now.target), nil
		case was.kind != "file":
			continue
		case was.executable != now.executable:
			return path + " (its executable bit)", nil
		}
		storedContent, err := os.ReadFile(was.file)
		if err != nil {
			return "", err
		}
		rebuiltContent, err := os.ReadFile(now.file)
		if err != nil {
			return "", err
		}
		if !bytes.Equal(storedContent, rebuiltContent) {
			return path + " " + differenceWithin(storedContent, rebuiltContent), nil
		}
	}
	return "", nil
}

// differenceEntry is one path of a product as firstDifference compares it.
type differenceEntry struct {
	kind, file, target string
	executable         bool
}

// differenceEntries is every path under a product's root, by its slash-separated path relative to the root.
func differenceEntries(directory string) (map[string]differenceEntry, error) {
	root, err := ProductRoot(directory)
	if err != nil {
		return nil, err
	}
	entries := map[string]differenceEntry{}
	err = filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, file)
		if err != nil || relative == "." {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		found := differenceEntry{kind: "other", file: file}
		switch {
		case entry.IsDir():
			found.kind = "directory"
		case info.Mode()&fs.ModeSymlink != 0:
			found.kind = "link"
			if found.target, err = os.Readlink(file); err != nil {
				return err
			}
		case info.Mode().IsRegular():
			found.kind = "file"
			found.executable = info.Mode()&0o111 != 0
		}
		entries[filepath.ToSlash(relative)] = found
		return nil
	})
	return entries, err
}

// differenceWithin says where two different contents of one file differ: in an object file both copies parse as, the
// first section that differs, else the offset of the first differing byte.
func differenceWithin(stored, rebuilt []byte) string {
	if section := objectSectionDifference(stored, rebuilt); section != "" {
		return section
	}
	offset := 0
	for offset < len(stored) && offset < len(rebuilt) && stored[offset] == rebuilt[offset] {
		offset++
	}
	return fmt.Sprintf("at byte %d", offset)
}

// objectSection is one section of an object file: its name and its bytes as the file holds them.
type objectSection struct {
	name string
	data []byte
}

// objectSectionDifference names the first section, in the stored copy's order, that the rebuilt copy lacks or holds
// different bytes in, when both copies parse as ELF or both as Mach-O; else, or when every section is alike (the
// difference is in headers or tables alone), "".
func objectSectionDifference(stored, rebuilt []byte) string {
	format, storedSections := objectSections(stored)
	rebuiltFormat, rebuiltSections := objectSections(rebuilt)
	if format == "" || format != rebuiltFormat {
		return ""
	}
	byName := map[string][]byte{}
	for _, section := range rebuiltSections {
		byName[section.name] = section.data
	}
	for _, section := range storedSections {
		data, held := byName[section.name]
		if !held {
			return fmt.Sprintf("(%s section %s, not in the rebuilt copy)", format, section.name)
		}
		if bytes.Equal(section.data, data) {
			continue
		}
		offset := 0
		for offset < len(section.data) && offset < len(data) && section.data[offset] == data[offset] {
			offset++
		}
		return fmt.Sprintf("(%s section %s, at its byte %d)", format, section.name, offset)
	}
	return ""
}

// objectSections parses content as ELF or Mach-O and lists its sections that hold bytes in the file, in its order.
func objectSections(content []byte) (string, []objectSection) {
	if file, err := elf.NewFile(bytes.NewReader(content)); err == nil {
		var sections []objectSection
		for _, section := range file.Sections {
			if section.Type == elf.SHT_NOBITS || section.Type == elf.SHT_NULL {
				continue
			}
			data, err := section.Data()
			if err != nil {
				return "", nil
			}
			sections = append(sections, objectSection{name: section.Name, data: data})
		}
		return "ELF", sections
	}
	if file, err := macho.NewFile(bytes.NewReader(content)); err == nil {
		var sections []objectSection
		for _, section := range file.Sections {
			data, err := section.Data()
			if err != nil {
				continue
			}
			sections = append(sections, objectSection{name: section.Seg + "," + section.Name, data: data})
		}
		return "Mach-O", sections
	}
	return "", nil
}
