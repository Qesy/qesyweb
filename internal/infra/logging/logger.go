package logging

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Dir, Timezone      string
	RetentionDays      int
	Console, AddSource bool
}

type dailyWriter struct {
	mu          sync.Mutex
	dir, prefix string
	location    *time.Location
	file        *os.File
	date        string
	retention   int
}

var errDailyWriterClosed = errors.New("daily log writer is closed")

func newDailyWriter(dir, prefix, timezone string, retention int) (*dailyWriter, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("load log timezone: %w", err)
	}
	w := &dailyWriter{dir: dir, prefix: prefix, location: location, retention: retention}
	if err := w.rotate(time.Now()); err != nil {
		return nil, err
	}
	return w, nil
}
func (w *dailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return 0, errDailyWriterClosed
	}
	if time.Now().In(w.location).Format("2006-01-02") != w.date {
		if err := w.rotate(time.Now()); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}
func (w *dailyWriter) rotate(now time.Time) error {
	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return err
	}
	date := now.In(w.location).Format("2006-01-02")
	oldFile := w.file
	w.file = nil
	if oldFile != nil {
		_ = oldFile.Close()
	}
	file, err := os.OpenFile(filepath.Join(w.dir, w.prefix+"-"+date+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if err := w.clean(); err != nil {
		_ = file.Close()
		return err
	}
	w.file, w.date = file, date
	return nil
}
func (w *dailyWriter) clean() error {
	if w.retention <= 0 {
		return nil
	}
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return err
	}
	cutoff := time.Now().In(w.location).AddDate(0, 0, -w.retention).Format("2006-01-02")
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), w.prefix+"-") || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		date := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), w.prefix+"-"), ".log")
		if date < cutoff {
			_ = os.Remove(filepath.Join(w.dir, entry.Name()))
		}
	}
	return nil
}
func (w *dailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

type Logger struct {
	App, Access *log.Logger
	appFile     *dailyWriter
	accessFile  *dailyWriter
	appWriter   io.Writer
	flags       int
}

func New(cfg Config) (*Logger, error) {
	app, err := newDailyWriter(cfg.Dir, "app", cfg.Timezone, cfg.RetentionDays)
	if err != nil {
		return nil, err
	}
	access, err := newDailyWriter(cfg.Dir, "access", cfg.Timezone, cfg.RetentionDays)
	if err != nil {
		_ = app.Close()
		return nil, err
	}
	appWriter := io.Writer(app)
	accessWriter := io.Writer(access)
	if cfg.Console {
		appWriter = io.MultiWriter(os.Stdout, app)
		accessWriter = io.MultiWriter(os.Stdout, access)
	}
	flags := log.LstdFlags
	if cfg.AddSource {
		flags |= log.Lshortfile
	}
	return &Logger{
		App:        log.New(appWriter, "", flags),
		Access:     log.New(accessWriter, "", flags),
		appFile:    app,
		accessFile: access,
		appWriter:  appWriter,
		flags:      flags,
	}, nil
}
func (l *Logger) Close() {
	if l == nil {
		return
	}
	_ = l.appFile.Close()
	_ = l.accessFile.Close()
}
func (l *Logger) With(component string) *log.Logger {
	return log.New(l.appWriter, "component="+component+" ", l.flags)
}

func (l *Logger) ConfigureStandardLog() {
	log.SetOutput(l.appWriter)
	log.SetFlags(l.flags)
}
