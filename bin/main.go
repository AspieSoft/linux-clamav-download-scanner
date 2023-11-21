package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AspieSoft/go-regex-re2/v2"
	"github.com/AspieSoft/goutil/bash"
	"github.com/AspieSoft/goutil/fs/v2"
	"github.com/AspieSoft/goutil/v7"
)

func main(){
	rootDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}

	user := os.Getenv("USER")
	var userDBUS string
	if os.Geteuid() == 0 {
		user = os.Getenv("SUDO_USER")
		if user == "" || user == "root" {
			user = ""
			if out, err := bash.Run([]string{`w`}, "", nil); err == nil {
				regex.Comp(`(?m)^([\w_\-]+)\s+seat[0-9]+`).RepFunc(out, func(data func(int) []byte) []byte {
					if user == "" {
						user = string(data(1))
					}
					return nil
				}, true)
			}

			if user == "" {
				user = "root"
			}
		}

		user = string(regex.Comp(`[^\w_\-]+`).RepStrLit([]byte(user), []byte{}))

		if out, err := bash.Run([]string{`runuser`, `-l`, user, `-c`, `echo $UID`}, "", nil); err == nil && len(out) != 0 {
			out = bytes.Trim(out, "\r\n ")
			if len(out) != 0 {
				userDBUS = `unix:path=/run/user/`+string(out)+`/bus`
			}
		}
	}

	user = string(regex.Comp(`[^\w_\-]+`).RepStrLit([]byte(user), []byte{}))

	newFiles := map[string]uint{}
	hasFiles := map[string]uint{}
	var mu sync.Mutex
	lastNotify := uint(0)
	notifyDelay := uint(3000)

	scanDirList := []string{
		"Downloads",
		"Desktop",
		"Documents",
		"Pictures",
		"Videos",
		"Music",
		"Public",
		"Templates",
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || !strings.HasPrefix(homeDir, "/home") {
		if os.Geteuid() != 0 {
			log.Fatal(errors.New("error: failed to get user home directory!"))
		}

		if out, err := bash.RunRaw(`getent passwd `+user+` | cut -d: -f6`, "", nil); err == nil {
			homeDir = string(bytes.Trim(out, "\r\n "))
		}else{
			log.Fatal(errors.New("error: failed to get user home directory!"))
		}
	}

	for _, arg := range os.Args[1:] {
		if dir := string(regex.Comp(`[^\w_-]+`).RepStr([]byte(arg), []byte{})); dir != "" {
			scanDirList = append(scanDirList, dir)
		}
	}


	// create quarantine directory if it does not exist
	if _, err := os.Stat("/VirusScan/quarantine"); err == nil || !strings.HasSuffix(err.Error(), "permission denied") {
		os.MkdirAll("/VirusScan", 0644)
		os.MkdirAll("/VirusScan/quarantine", 2660)
	}


	// add user dirs to scanDirList
	if buf, err := os.ReadFile(homeDir+"/.config/user-dirs.dirs"); err == nil {
		regex.Comp(`(?m)^[\w_\-]+\s*=\s*(.*)$`).RepFunc(buf, func(data func(int) []byte) []byte {
			dirPath := string(data(1)[len([]byte(homeDir))+1:])
			if !goutil.Contains(scanDirList, dirPath) {
				scanDirList = append(scanDirList, dirPath)
			}
			return []byte{}
		}, true)
	}


	// add browser (and other) extensions directories to scanDirList
	if out, err := bash.Run([]string{`find`, homeDir, `-type`, `d`, `-name`, `*xtensions`}, "", nil); err == nil {
		for _, dir := range bytes.Split(out, []byte{'\n'}) {
			dir = bytes.Trim(dir, "\r\n ")
			if len(dir) != 0 {
				dirPath := string(dir[len([]byte(homeDir))+1:])
				if !goutil.Contains(scanDirList, dirPath) {
					scanDirList = append(scanDirList, dirPath)
				}
			}
		}
	}

	if out, err := bash.Run([]string{`find`, homeDir, `-type`, `d`, `-name`, `*xtension`}, "", nil); err == nil {
		for _, dir := range bytes.Split(out, []byte{'\n'}) {
			dir = bytes.Trim(dir, "\r\n ")
			if len(dir) != 0 {
				dirPath := string(dir[len([]byte(homeDir))+1:])
				if !goutil.Contains(scanDirList, dirPath) {
					scanDirList = append(scanDirList, dirPath)
				}
			}
		}
	}


	// add custom dirs to scanDirList
	if buf, err := os.ReadFile(homeDir+"/.aspiesoft-clamav-auto-scan"); err == nil {
		regex.Comp(`(?m)^[\w_\-]+\s*=\s*(.*)$`).RepFunc(buf, func(data func(int) []byte) []byte {
			dirPath := string(data(1)[len([]byte(homeDir))+1:])
			if !goutil.Contains(scanDirList, dirPath) {
				scanDirList = append(scanDirList, dirPath)
			}
			return []byte{}
		}, true)
	}

	if buf, err := os.ReadFile(homeDir+"/.clamav-auto-scan"); err == nil {
		regex.Comp(`(?m)^[\w_\-]+\s*=\s*(.*)$`).RepFunc(buf, func(data func(int) []byte) []byte {
			dirPath := string(data(1)[len([]byte(homeDir))+1:])
			if !goutil.Contains(scanDirList, dirPath) {
				scanDirList = append(scanDirList, dirPath)
			}
			return []byte{}
		}, true)
	}

	// add custom dirs to scanDirList from root
	if buf, err := os.ReadFile(homeDir+"/usr/share/config/aspiesoft-clamav-auto-scan"); err == nil {
		regex.Comp(`(?m)^[\w_\-]+\s*=\s*(.*)$`).RepFunc(buf, func(data func(int) []byte) []byte {
			dirPath := string(data(1)[len([]byte(homeDir))+1:])
			if !goutil.Contains(scanDirList, dirPath) {
				scanDirList = append(scanDirList, dirPath)
			}
			return []byte{}
		}, true)
	}

	if buf, err := os.ReadFile(homeDir+"/usr/share/config/clamav-auto-scan"); err == nil {
		regex.Comp(`(?m)^[\w_\-]+\s*=\s*(.*)$`).RepFunc(buf, func(data func(int) []byte) []byte {
			dirPath := string(data(1)[len([]byte(homeDir))+1:])
			if !goutil.Contains(scanDirList, dirPath) {
				scanDirList = append(scanDirList, dirPath)
			}
			return []byte{}
		}, true)
	}


	watcher := fs.Watcher()
	defer watcher.CloseWatcher("*")

	var downloadDir string

	for _, dir := range scanDirList {
		if path, err := fs.JoinPath(homeDir, dir); err == nil {
			watcher.WatchDir(path)
			if downloadDir == "" && dir == "Downloads" {
				downloadDir = path
			}
		}
	}

	watcher.OnFileChange = func(path, op string) {
		mu.Lock()
		newFiles[path] = uint(time.Now().UnixMilli())
		hasFiles[path] = uint(time.Now().UnixMilli())
		mu.Unlock()
	}

	watcher.OnRemove = func(path, op string) (removeWatcher bool) {
		mu.Lock()
		delete(newFiles, path)
		delete(hasFiles, path)
		mu.Unlock()
		return true
	}

	scanFile := make(chan string)

	running := true

	go func(){
		for {
			if !running {
				break
			}

			mu.Lock()
			now := uint(time.Now().UnixMilli())
			for path, modified := range newFiles {
				if now - modified > 1000 {
					scanFile <- path
					delete(newFiles, path)
				}
			}
			mu.Unlock()
		}
	}()

	go func(){
		for {
			file := <- scanFile

			if file == "" {
				break
			}

			// prevent removed or recently changed files from staying at the beginning of the queue
			mu.Lock()
			now := uint(time.Now().UnixMilli())
			if modified, ok := hasFiles[file]; !ok || now - modified < 1000 {
				mu.Unlock()
				continue
			}
			delete(hasFiles, file)
			mu.Unlock()

			cmd := exec.Command(`nice`, `-n`, `15`, `clamscan`, `-r`, `--bell`, `--move=/VirusScan/quarantine`, `--exclude-dir=/VirusScan/quarantine`, file)
			cmd.Dir = homeDir
			cmd.Env = os.Environ()

			success := false

			if stdout, err := cmd.StdoutPipe(); err == nil {
				go func(){
					onSummary := false
					for {
						b := make([]byte, 1024)
						_, err := stdout.Read(b)
						if err != nil {
							break
						}

						if !onSummary && regex.Comp(`(?i)-+\s*scan\s+summ?[ae]ry\s*-+`).Match(b) {
							onSummary = true
							success = true
						}

						if onSummary && regex.Comp(`(?i)infected\s+files:?\s*([0-9]+)`).Match(b) {
							inf := 0
							regex.Comp(`(?i)infected\s+files:?\s*([0-9]+)`).RepFunc(b, func(data func(int) []byte) []byte {
								if i, err := strconv.Atoi(string(data(1))); err == nil && i > inf {
									inf = i
								}
								return nil
							}, true)

							fmt.Println("\nFile/Dir:", file, "\n  Infected files:", inf)

							if inf == 0 && downloadDir != "" && strings.HasPrefix(file, downloadDir) {
								now := uint(time.Now().UnixMilli())
								if now - lastNotify > notifyDelay {
									lastNotify = now
									if os.Geteuid() == 0 {
										bash.Run([]string{`pkexec`, `--user`, user, `./notify.sh`, userDBUS, rootDir+`/assets/green.png`, `File Is Safe`, file}, rootDir, nil)
									}else{
										bash.Run([]string{`notify-send`, `-i`, rootDir+`/assets/green.png`, `-t`, `3`, `File Is Safe`, file}, rootDir, nil)
									}
								}
							}else if inf != 0 {
								now := uint(time.Now().UnixMilli())
								if now - lastNotify > notifyDelay {
									lastNotify = now
									if os.Geteuid() == 0 {
										bash.Run([]string{`pkexec`, `--user`, user, `./notify.sh`, userDBUS, rootDir+`/assets/red.png`, `Warning: File Has Been Moved To Quarantine`, file}, rootDir, nil)
									}else{
										bash.Run([]string{`notify-send`, `-i`, rootDir+`/assets/red.png`, `-t`, `3`, `Warning: File Has Been Moved To Quarantine`, file}, rootDir, nil)
									}
								}
							}

							break
						}
					}
				}()
			}

			if downloadDir != "" && strings.HasPrefix(file, downloadDir) && user != "" && user != "root" {
				now := uint(time.Now().UnixMilli())
				if now - lastNotify > notifyDelay {
					lastNotify = now
					if os.Geteuid() == 0 {
						bash.Run([]string{`pkexec`, `--user`, user, `./notify.sh`, userDBUS, rootDir+`/assets/blue.png`, `Started Scanning File`, file}, rootDir, nil)
					}else{
						bash.Run([]string{`notify-send`, `-i`, rootDir+`/assets/blue.png`, `-t`, `3`, `Started Scanning File`, file}, rootDir, nil)
					}
				}
			}

			err := cmd.Run()
			if err != nil && !success {
				fmt.Println(err)

				if downloadDir != "" && strings.HasPrefix(file, downloadDir) && user != "" && user != "root" {
					now := uint(time.Now().UnixMilli())
					if now - lastNotify > notifyDelay {
						lastNotify = now
						if os.Geteuid() == 0 {
							bash.Run([]string{`pkexec`, `--user`, user, `./notify.sh`, userDBUS, rootDir+`/assets/blue.png`, `Error: Failed To Scan File`, file}, rootDir, nil)
						}else{
							bash.Run([]string{`notify-send`, `-i`, rootDir+`/assets/blue.png`, `-t`, `3`, `Error: Failed To Scan File`, file}, rootDir, nil)
						}
					}
				}
			}

			time.Sleep(250 * time.Millisecond)
		}
	}()

	watcher.Wait()
	running = false
	scanFile <- ""
}
