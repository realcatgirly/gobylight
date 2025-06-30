package provider

import (
	"fmt"
	"image/color"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// The teams provider sends out a color when your teams status changes,
// based on the logfiles that teams writes out

var (
	statusRegex = regexp.MustCompile("^.*Badge.*status (.*)$")
)

func init() {
	Providers["teams"] = newTeams
}

func newTeams(c chan color.RGBA, done chan struct{}) error {
	statusC := make(chan TeamsStatus)
	pathC := make(chan string, 1)
	readerDone := make(chan struct{})
	if err := newTeamsLogfileTracker(pathC, done); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case path := <-pathC:
				close(readerDone)
				readerDone = make(chan struct{})
				if err := newTeamsLogfileReader(path, statusC, readerDone); err != nil {
					panic(err)
				}
			case status := <-statusC:
				if color, ok := StatusColor[status]; ok {
					c <- color
				}
			case <-done:
				return
			}
		}
	}()

	return nil
}

func newTeamsLogfileTracker(pathC chan string, done chan struct{}) error {
	const logFilePath = `\Packages\MSTeams_8wekyb3d8bbwe\LocalCache\Microsoft\MSTeams\Logs`
	userDirPrefix, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	var logFileFolder = userDirPrefix + logFilePath
	_ = logFileFolder

	logWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-logWatcher.Events:
				if !ok {
					continue
				}
				if !isRelevantEvent(event) {
					continue
				}
				fileName, err := getLatestFileName(logFileFolder)
				if err != nil {
					fmt.Printf("%s", err.Error())
					continue
				}
				path := logFileFolder + "\\" + fileName
				pathC <- path
			case err, ok := <-logWatcher.Errors:
				if !ok {
					continue
				}
				fmt.Printf("%s", err.Error())
			case <-done:
				logWatcher.Close()
				return
			}
		}
	}()

	fileName, err := getLatestFileName(logFileFolder)
	if err != nil {
		return err
	}
	path := logFileFolder + "\\" + fileName
	pathC <- path

	return nil
}

func newTeamsLogfileReader(filePath string, statusC chan TeamsStatus, done chan struct{}) error {
	fmt.Printf("Reading Teams logfile %s\n", filePath)
	offset := int64(0)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				info, err := os.Stat(filePath)
				if err != nil {
					fmt.Printf("Error stating file %s: %s\n", filePath, err.Error())
					continue
				}
				if info.Size() <= offset {
					continue // file has not changed
				}
				file, err := os.Open(filePath)
				if err != nil {
					fmt.Printf("Error opening file %s: %s\n", filePath, err.Error())
					continue
				}
				defer file.Close()
				file.Seek(offset, 0) // Seek to the last read position
				buffer := make([]byte, info.Size()-offset)
				_, err = file.Read(buffer)
				if err != nil {
					fmt.Printf("Error reading file %s: %s\n", filePath, err.Error())
					continue
				}
				lines := strings.Split(string(buffer), "\n")
				newestStatus := TeamsStatusUnknown
				for _, line := range lines {
					if status := readLine(line); status != TeamsStatusUnknown {
						newestStatus = status
					}
				}
				if newestStatus != TeamsStatusUnknown {
					statusC <- newestStatus // Send the latest status to the channel
				}
				offset = info.Size() // Update the offset to the new end of the file
				time.Sleep(time.Second)
			}
		}
	}()

	return nil
}

func getLatestFileName(folderPath string) (string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}
	var newestFile os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !isRelevantFile(file.Name()) {
			continue
		}
		fileInfo, err := file.Info()
		if err != nil {
			continue
		}
		if newestFile == nil || fileInfo.ModTime().After(newestFile.ModTime()) {
			newestFile = fileInfo
		}
	}
	if newestFile == nil {
		return "", fmt.Errorf("no relevant files found")
	}
	return newestFile.Name(), nil
}

func isRelevantEvent(event fsnotify.Event) bool {
	if event.Op != fsnotify.Create {
		return false
	}
	return isRelevantFile(event.Name)
}

func isRelevantFile(path string) bool {
	return strings.Contains(path, "MSTeams_") && strings.HasSuffix(path, ".log")
}

func readLine(line string) TeamsStatus {
	if !strings.Contains(line, "Badge") || !strings.Contains(line, "status") {
		return TeamsStatusUnknown
	}
	line = strings.ReplaceAll(line, "\r", "")
	matches := statusRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return TeamsStatusUnknown
	}
	if status, ok := StatusString[matches[1]]; ok {
		return status
	}
	return TeamsStatusUnknown
}

type TeamsStatus int

const (
	TeamsStatusUnknown TeamsStatus = iota
	TeamsStatusAvailable
	TeamsStatusBusy
	TeamsStatusDoNotDisturb
	TeamsStatusAway
	TeamsStatusOffline
)

var StatusString = map[string]TeamsStatus{
	"Available":      TeamsStatusAvailable,
	"Busy":           TeamsStatusBusy,
	"Do not disturb": TeamsStatusDoNotDisturb,
	"Away":           TeamsStatusAway,
	"Offline":        TeamsStatusOffline,
}

var StatusColor = map[TeamsStatus]color.RGBA{
	TeamsStatusAvailable:    {R: 0, G: 255, B: 0},   //rgb(0,255,0)
	TeamsStatusBusy:         {R: 255, G: 0, B: 0},   //rgb(255,0,0)
	TeamsStatusDoNotDisturb: {R: 255, G: 0, B: 0},   //rgb(255,0,0)
	TeamsStatusAway:         {R: 234, G: 163, B: 0}, //rgb(234,163,0)
	TeamsStatusOffline:      {R: 0, G: 0, B: 0},     //rgb(0,0,0)
}
