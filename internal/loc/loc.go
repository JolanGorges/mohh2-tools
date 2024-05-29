package loc

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"mohh2-tools/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf16"
)

type loc struct {
	Label   string
	DirToId map[string]uint32
}

func extractHashes(path string, sem chan bool, wg *sync.WaitGroup, m *sync.Mutex, dict map[uint32]loc) {
	defer wg.Done()
	file, err := os.Open(path)
	if err != nil {
		<-sem
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var header idxHeader
	err = binary.Read(reader, binary.BigEndian, &header)
	if err != nil || header.Signature != 0x20040519 {
		fmt.Println("Invalid header")
		<-sem
		return
	}

	dir := filepath.Dir(path)
	for {
		var entry idxEntry
		err = binary.Read(reader, binary.BigEndian, &entry)
		if err != nil {
			break
		}
		m.Lock()
		if value, ok := dict[entry.Hash]; ok {
			if value.DirToId == nil {
				value.DirToId = make(map[string]uint32)
			} else if _, ok := value.DirToId[dir]; ok {
				fmt.Printf("Hash collision: %s\n", dir)
			}
			value.DirToId[dir] = entry.Id
			dict[entry.Hash] = value
		} else {
			dict[entry.Hash] = loc{DirToId: map[string]uint32{dir: entry.Id}}
		}
		m.Unlock()
	}

	<-sem
}

func extractAllHashes(path string, dict map[uint32]loc) {
	sem := make(chan bool, utils.MaxGoroutines)
	var wg sync.WaitGroup
	var m sync.Mutex

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "string.idx" {
			wg.Add(1)
			sem <- true
			go extractHashes(path, sem, &wg, &m, dict)
		}
		return nil
	})
	wg.Wait()
}

func ExtractAllStrings(path string) {
	var dict = make(map[uint32]loc)
	extractAllHashes(path, dict)
	findAllLabels(path, dict)
	extractAllLocs(path, dict)
}

func findLabels(path string, sem chan bool, wg *sync.WaitGroup, m *sync.Mutex, dict map[uint32]loc) {
	defer wg.Done()
	file, err := os.Open(path)
	if err != nil {
		<-sem
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	var buffer []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		if b >= 0x30 && b <= 0x39 || b >= 0x41 && b <= 0x5A || b >= 0x61 && b <= 0x7A || b == 0x5F {
			buffer = append(buffer, b)
		} else if b == 0x24 && len(buffer) == 0 {
			continue
		} else if b == 0 && len(buffer) > 1 {
			label := string(buffer)
			hash := GetHash(label)
			m.Lock()
			if value, ok := dict[hash]; ok {
				if value.Label != "" && value.Label != label {
					fmt.Printf("Hash collision: \"%s\" -> \"%s\" \n", value.Label, label)
				}
				value.Label = label
				dict[hash] = value
			}

			m.Unlock()
			buffer = nil
		} else {
			buffer = nil
		}
	}

	<-sem
}

type idxHeader struct {
	Signature uint32
	Size      uint32
}

type idxEntry struct {
	Hash uint32
	Id   uint32
}

func findAllLabels(path string, dict map[uint32]loc) {
	sem := make(chan bool, utils.MaxGoroutines)
	var wg sync.WaitGroup
	var m sync.Mutex

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			wg.Add(1)
			sem <- true
			go findLabels(path, sem, &wg, &m, dict)
		}
		return nil
	})
	wg.Wait()
}

func FindDumps() []string {
	var paths []string
	filepath.Walk(utils.GameDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "game.dol" {
			paths = append(paths, filepath.Dir(path))
		}
		return nil
	})
	return paths
}

type locFile struct {
	Header  lochHeader
	Offsets []uint32
	Entries []loclEntry
}

type loclEntry struct {
	Header  loclHeader
	Offsets []uint32
	Entries []string
}

type lochHeader struct {
	Signature  uint32
	Unknown1   uint32
	Unknown2   uint32
	NumEntries uint32
}

type loclHeader struct {
	Signature  uint32
	Size       uint32
	Unknown1   uint32
	NumEntries uint32
}

func extractLocs(path string, dict map[uint32]loc) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file", err)
		return
	}
	defer file.Close()

	locFile := locFile{}

	err = binary.Read(file, binary.LittleEndian, &locFile.Header)
	if err != nil || locFile.Header.Signature != 0x48434F4C {
		fmt.Println("Invalid header")
		return
	}
	if locFile.Header.NumEntries != 1 {
		fmt.Println("Invalid number of entries")
		return
	}

	locFile.Offsets = make([]uint32, locFile.Header.NumEntries)
	binary.Read(file, binary.LittleEndian, &locFile.Offsets)

	path2 := filepath.Join("extracted", strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))))
	dir := filepath.Dir(path)
	for {
		if _, err := os.Stat(path2 + ".txt"); os.IsNotExist(err) {
			break
		}
		path2 = path2 + "_" + filepath.Base(dir)
		dir = filepath.Dir(dir)
	}
	file2, err := os.Create(path2 + ".txt")
	if err != nil {
		fmt.Println("Error creating file", err)
		return
	}
	defer file2.Close()
	locFile.Entries = make([]loclEntry, locFile.Header.NumEntries)
	dir = filepath.Dir(path)
	for i, offset := range locFile.Offsets {
		file.Seek(int64(offset), 0)
		loclEntry := &locFile.Entries[i]
		binary.Read(file, binary.LittleEndian, &loclEntry.Header)
		if loclEntry.Header.Signature != 0x4C434F4C {
			fmt.Println("Invalid header")
			return
		}

		loclEntry.Offsets = make([]uint32, loclEntry.Header.NumEntries)
		binary.Read(file, binary.LittleEndian, &loclEntry.Offsets)
		loclEntry.Entries = make([]string, loclEntry.Header.NumEntries)

		for i := 0; i < int(loclEntry.Header.NumEntries); i++ {
			file.Seek(int64(loclEntry.Offsets[i])+0x14, 0)
			var buffer []uint16
			var err error
			for {
				var c uint16
				err = binary.Read(file, binary.LittleEndian, &c)
				if err != nil {
					break
				}
				if c == 0 && len(buffer) > 0 {
					for key, value := range dict {
						if _, ok := value.DirToId[dir]; ok {
							if value.DirToId[dir] == uint32(i) {
								file2.WriteString(fmt.Sprintf("0x%08X=%s=%s\n", key, value.Label, string(utf16.Decode(buffer))))
							}
						}
					}
					break
				}
				buffer = append(buffer, c)
			}
			if err == io.EOF {
				break
			}
		}

	}
}

func extractAllLocs(path string, dict map[uint32]loc) {
	if _, err := os.Stat("extracted"); os.IsNotExist(err) {
		os.Mkdir("extracted", 0755)
	} else {
		os.RemoveAll("extracted")
		os.Mkdir("extracted", 0755)
	}
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".loc" {
			fmt.Println(path)
			extractLocs(path, dict)
		}
		return nil
	})
}

func GetHash(inputString string) uint32 {
	if inputString[0] == '$' {
		inputString = inputString[1:]
	}
	var hash int32 = -1
	for _, c := range inputString {
		hash = 33*hash + int32(c)
	}
	return uint32(hash)
}
