package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

const MaxGoroutines = 100

var TempDir string

func GameDir() string {
	gameFilesDir := "game_files"
	if _, err := os.Stat(gameFilesDir); os.IsNotExist(err) {
		os.Mkdir(gameFilesDir, 0777)
	}
	return gameFilesDir
}

func MakeTempDir() {
	var err error
	TempDir, err = os.MkdirTemp(".", "tmp")
	if err != nil {
		fmt.Println("Error creating temp directory")
		os.Exit(1)
	}
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err

}

func unzip(source string, dest string, files []string) error {
	read, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer read.Close()
	for _, file := range read.File {
		for _, f := range files {
			if filepath.Base(file.Name) == f {
				open, err := file.Open()
				if err != nil {
					return err
				}
				name := path.Join(dest, f)
				os.MkdirAll(path.Dir(name), os.ModeDir)
				create, err := os.Create(name)
				if err != nil {
					return err
				}
				defer create.Close()
				create.ReadFrom(open)
				break
			}
		}
	}
	return nil

}

func downloadAndUnzip(fileUrl string, tools []string) error {
	f, err := os.CreateTemp("", "temp")
	if err != nil {
		fmt.Println("Error creating temp file", err)
		os.Exit(1)
	}
	defer os.Remove(f.Name())
	err = downloadFile(f.Name(), fileUrl)
	if err != nil {
		return err
	}
	unzip(f.Name(), "tools", tools)
	return nil
}

func DonwloadWitTools() {
	fmt.Println("Downloading WIT tools...")
	tools := []string{"wit.exe", "cygwin1.dll", "cygcrypto-1.1.dll", "cygncursesw-10.dll", "cygz.dll"}
	err := downloadAndUnzip("https://wit.wiimm.de/download/wit-v3.05a-r8638-cygwin64.zip", tools)
	if err != nil {
		fmt.Println("Error downloading WIT tools", err)
		os.Exit(1)
	}
}

func DownloadQuickBMS() {
	fmt.Println("Downloading QuickBMS...")
	err := downloadAndUnzip("https://aluigi.altervista.org/papers/quickbms.zip", []string{"quickbms.exe"})
	if err != nil {
		fmt.Println("Error downloading QuickBMS", err)
		os.Exit(1)
	}

	err = downloadFile("tools/ea_big4.bms", "https://aluigi.altervista.org/bms/ea_big4.bms")
	if err != nil {
		fmt.Println("Error downloading ea_big4.bms", err)
		os.Exit(1)
	}
}
