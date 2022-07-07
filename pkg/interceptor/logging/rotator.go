package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultTimeFormat  = time.RFC3339
	defaultMaxDuration = 1 * time.Hour
	defaultMaxSize     = uint64(100 << 20)
)

// FileRotator implements file rotator as io.WriteCloser for logging.
type FileRotator struct {
	fd    *os.File
	id    uint64
	size  uint64
	ctime time.Time

	FileName    string
	MaxSize     uint64
	MaxDuration time.Duration
	TimeFormat  string
}

func (r *FileRotator) fileName() string {
	if r.FileName != "" {
		return r.FileName
	}
	// Default: create logs in the working directory
	return filepath.Join(".", filepath.Base(os.Args[0])+".log")
}

func (r *FileRotator) maxSize() uint64 {
	if r.MaxSize > 0 {
		return r.MaxSize
	}
	return defaultMaxSize
}

func (r *FileRotator) maxDuration() time.Duration {
	if r.MaxDuration > 0 {
		return r.MaxDuration
	}
	return defaultMaxDuration
}

func (r *FileRotator) timeFormat() string {
	if r.TimeFormat != "" {
		return r.TimeFormat
	}
	return defaultTimeFormat
}

func (r *FileRotator) init() error {
	r.id = 0
	r.size = 0
	r.ctime = time.Now()
	err := os.MkdirAll(filepath.Dir(r.fileName()), os.FileMode(0755))
	if err != nil {
		return err
	}
	return r.open()
}

func (r *FileRotator) rotate(ctime time.Time) error {
	r.id++
	r.size = 0
	r.ctime = ctime
	return r.open()
}

func (r *FileRotator) open() error {
	if err := r.close(); err != nil {
		return err
	}
	// Open as a new file anyway
	name := fmt.Sprintf("%s_%s_%v", r.fileName(), r.ctime.Format(r.timeFormat()), r.id)
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	mode := os.FileMode(0644)
	if fd, err := os.OpenFile(filepath.Clean(name), flag, mode); err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	} else {
		r.fd = fd
	}

	return nil
}

func (r *FileRotator) close() error {
	if r.fd == nil {
		return nil
	}
	_ = r.fd.Sync()
	err := r.fd.Close()
	r.fd = nil
	return err
}

// Write implements io.Writer.
func (r *FileRotator) Write(p []byte) (n int, err error) {
	// Check write length
	writeLen := uint64(len(p))
	if writeLen > r.maxSize() {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, r.maxSize(),
		)
	}
	// Trigger the initial open
	if r.fd == nil {
		if err = r.init(); err != nil {
			return 0, err
		}
	}
	// Trigger log rotate
	if ctime := time.Now(); r.size+writeLen > r.maxSize() || ctime.Sub(r.ctime) > r.maxDuration() {
		if err = r.rotate(ctime); err != nil {
			return 0, err
		}
	}
	// Write
	n, err = r.fd.Write(p)
	r.size += uint64(n)

	return n, err
}

// Close implements io.Closer.
func (r *FileRotator) Close() error {
	return r.close()
}
