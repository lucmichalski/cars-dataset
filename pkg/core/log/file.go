package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type fileLogWriter struct {
	sync.Mutex
	fd       *os.File
	opendate time.Time
	openday  int
	FileName string
	MaxDays  int
}

func (fw *fileLogWriter) createLogFile() (*os.File, error) {
	// Open the log file
	var err error
	fw.fd, err = os.OpenFile(fw.FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err == nil {
		fw.opendate = time.Now()
		fw.openday = fw.opendate.Day()
	}
	return fw.fd, err
}

func (fw *fileLogWriter) Write(b []byte) (int, error) {
	fw.Lock()
	defer fw.Unlock()
	fw.check()
	return fw.fd.Write(b)
}

func (fw *fileLogWriter) check() {
	if time.Now().Day() != fw.openday {
		if err := fw.rotate(); err != nil {
			fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", fw.FileName, err)
			return
		}
	}
}

// write log in new file.
// new file name like xx.log.2013-01-01.2
func (fw *fileLogWriter) rotate() error {
	_, err := os.Lstat(fw.FileName)
	if err == nil { // file exists
		fname := fw.FileName + fmt.Sprintf(".%s", fw.opendate.Format("2006-01-02"))

		// close file before rename
		fw.fd.Close()

		// Rename the file to its newfound home
		err = os.Rename(fw.FileName, fname)
		if err != nil {
			return fmt.Errorf("Rotate: %s\n", err)
		}
		// re-start logger
		_, err := fw.createLogFile()
		if err != nil {
			return fmt.Errorf("Rotate StartLogger: %s\n", err)
		}

		go fw.deleteOldLog()
	}

	return nil
}
func (fw *fileLogWriter) deleteOldLog() (err error) {
	if fw.MaxDays <= 0 {
		return
	}
	dir := filepath.Dir(fw.FileName)
	filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				returnErr = fmt.Errorf("Unable to delete old log '%s', error: %+v", path, r)
				fmt.Println(returnErr)
			}
		}()

		if !info.IsDir() && info.ModTime().Unix() < (time.Now().Unix()-60*60*24*int64(fw.MaxDays)) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(fw.FileName)) {
				os.Remove(path)
			}
		}
		return
	})
	return
}
