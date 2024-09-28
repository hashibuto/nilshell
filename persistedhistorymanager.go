package ns

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

type PersistedHistoryManager struct {
	*BasicHistoryManager
	filename  string
	fileLock  *FileLock
	wg        sync.WaitGroup
	killChan  chan struct{}
	unwritten []string
	flushLock sync.Mutex
}

func NewPersistedHistoryManager(maxKeep int, filename string) *PersistedHistoryManager {

	pm := &PersistedHistoryManager{
		BasicHistoryManager: NewBasicHistoryManager(maxKeep),
		filename:            filename,
		fileLock:            NewFileLock(fmt.Sprintf("%s.lock", filename)),
		killChan:            make(chan struct{}, 1),
	}
	pm.load()

	pm.wg.Add(1)
	go pm.flushThread()

	return pm
}

func (pm *PersistedHistoryManager) Push(value string) {
	if value == pm.BasicHistoryManager.prev {
		return
	}

	pm.flushLock.Lock()
	defer pm.flushLock.Unlock()
	pm.BasicHistoryManager.Push(value)
	pm.unwritten = append(pm.unwritten, base64.StdEncoding.EncodeToString([]byte(value))+"\n")
}

func (pm *PersistedHistoryManager) Exit() {
	pm.BasicHistoryManager.Exit()
	close(pm.killChan)
	pm.wg.Wait()
}

func (pm *PersistedHistoryManager) flushThread() {
	defer pm.wg.Done()

	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			err := pm.flushChanges()
			if err != nil {
				slog.Error(err.Error())
			}
		case <-pm.killChan:
			err := pm.flushChanges()
			if err != nil {
				slog.Error(err.Error())
			}
			return
		}
	}
}

func (pm *PersistedHistoryManager) flushChanges() error {
	var unwrittenCopy []string
	pm.flushLock.Lock()
	if len(pm.unwritten) > 0 {
		unwrittenCopy = make([]string, len(pm.unwritten))
		copy(unwrittenCopy, pm.unwritten)
		pm.unwritten = []string{}
	}
	pm.flushLock.Unlock()

	if len(unwrittenCopy) > 0 {
		pm.fileLock.Lock()
		defer pm.fileLock.Unlock()

		f, err := os.OpenFile(pm.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		_, err = f.WriteString(strings.Join(unwrittenCopy, ""))
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *PersistedHistoryManager) load() error {
	pm.fileLock.Lock()
	defer pm.fileLock.Unlock()

	fBytes, err := os.ReadFile(pm.filename)
	if err != nil {
		return nil
	}

	fStrs := string(fBytes)
	fHist := strings.Split(fStrs, "\n")

	for _, enc := range fHist {
		b, err := base64.StdEncoding.DecodeString(enc)
		if err != nil {
			continue
		}

		cmd := string(b)
		cmd = strings.Trim(cmd, "\r\n ")
		if len(cmd) > 0 {
			pm.BasicHistoryManager.Push(cmd)
		}
	}

	iter := pm.BasicHistoryManager.index.GetIter()
	commands := []string{}
	for iter.Next() {
		commands = append(commands, base64.StdEncoding.EncodeToString([]byte(iter.Value()))+"\n")
	}
	err = os.WriteFile(pm.filename, []byte(strings.Join(commands, "")), 0644)
	if err != nil {
		return err
	}

	return nil
}
