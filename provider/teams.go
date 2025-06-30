package provider

import (
	"fmt"
	"image/color"
	"os"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// The teams provider sends out a color when your teams status changes,
// based on the logfiles that teams writes out

var (
	statusRegex = regexp.MustCompile("^.*TaskbarBadgeServicePackaged:.*status (.*)$")
)

func init() {
	Providers["teams"] = newTeams
}

func newTeams(c chan color.RGBA, done chan struct{}) error {
	statusC := make(chan TeamsStatus)
	if err := newTeamsLogfileReader(statusC, done); err != nil {
		return err
	}

	go func() {
		for {
			select {
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

func newTeamsLogfileReader(tsc chan TeamsStatus, done chan struct{}) error {
	const logFilePath = `\Packages\MSTeams_8wekyb3d8bbwe\LocalCache\Microsoft\MSTeams\Logs`
	userDirPrefix, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	var logFileFolder = userDirPrefix + logFilePath

	logWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-logWatcher.Events:
				fmt.Printf("event: %s\n", event.Name)
				if !ok {
					return
				}
				if !isRelevant(event) {
					return
				}
				status, err := getLatestStatus(event.Name)
				if err != nil {
					fmt.Printf("%s", err.Error())
					return
				}
				if status != TeamsStatusUnknown {
					tsc <- status
				}
			case err, ok := <-logWatcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("%s", err.Error())
			case <-done:
				return
			}
		}
	}()

	if err := logWatcher.Add(logFileFolder); err != nil {
		return err
	}

	return nil
}

func isRelevant(event fsnotify.Event) bool {
	if event.Op != fsnotify.Write {
		return false
	}
	return strings.Contains(event.Name, "MSTeams_") && strings.HasSuffix(event.Name, ".log")
}

// read log file line by line backwards and try to find the latest userPresence status
func getLatestStatus(logFilePath string) (TeamsStatus, error) {
	logFile, err := os.Open(logFilePath)
	if err != nil {
		return TeamsStatusUnknown, err
	}
	fileInfo, err := logFile.Stat()
	if err != nil {
		return TeamsStatusUnknown, err
	}
	buffer := make([]byte, fileInfo.Size())
	logFile.Read(buffer)
	lines := strings.Split(string(buffer), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.ReplaceAll(lines[i], "\r", "")
		if !strings.Contains(line, "TaskbarBadgeServicePackaged") {
			continue
		}
		matches := statusRegex.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		if status, ok := StatusString[matches[1]]; ok {
			return status, nil
		}
	}

	return TeamsStatusUnknown, nil
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
