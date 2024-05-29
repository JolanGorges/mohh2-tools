package big

import (
	"fmt"
	"mohh2-tools/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func extractFile(filename string, sem chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	_, err := exec.Command("cmd", "/C", filepath.Join("tools", "quickbms.exe"), "-d", "-K", filepath.Join("tools", "ea_big4.bms"), filename, "").Output()
	if err != nil {
		fmt.Printf("Error extracting %s: %s\n", filename, err)
		return
	}

	err = os.Remove(filename)
	if err != nil {
		fmt.Printf("Error removing %s: %s\n", filename, err)
		return
	}
	<-sem
}

func ExtractBigVivFiles() {
	sem := make(chan bool, utils.MaxGoroutines)
	var wg sync.WaitGroup

	found := false
	err := filepath.Walk(utils.GameDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".big") || strings.HasSuffix(path, ".viv")) {
			found = true
			sem <- true
			wg.Add(1)
			go extractFile(path, sem, &wg)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %s: %s\n", utils.GameDir(), err)
		return
	}
	wg.Wait()
	if found {
		ExtractBigVivFiles()
	}
}
