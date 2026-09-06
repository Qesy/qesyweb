package logging

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Dir, Timezone      string
	RetentionDays      int
	Console, AddSource bool
	Level              string
}

type dailyWriter struct {
	mu          sync.Mutex
	dir, prefix string
	location    *time.Location
	file        *os.File
	date        string
	retention   int
}

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
	if w.file != nil {
		_ = w.file.Close()
	}
	file, err := os.OpenFile(filepath.Join(w.dir, w.prefix+"-"+date+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file, w.date = file, date
	return w.clean()
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
	return w.file.Close()
}

type Logger struct {
	App, Access *slog.Logger
	closers     []io.Closer

	appWriter io.Writer // 标准log添加
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
	level := slog.LevelInfo
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level, AddSource: cfg.AddSource}
	appWriter := io.Writer(app)
	accessWriter := io.Writer(access)
	if cfg.Console {
		appWriter = io.MultiWriter(os.Stdout, app)
		accessWriter = io.MultiWriter(os.Stdout, access)
	}
	return &Logger{App: slog.New(slog.NewTextHandler(appWriter, opts)), Access: slog.New(slog.NewTextHandler(accessWriter, opts)), appWriter: appWriter, closers: []io.Closer{app, access}}, nil
}
func (l *Logger) Close() {
	for _, closer := range l.closers {
		_ = closer.Close()
	}
}
func (l *Logger) With(component string) *slog.Logger       { return l.App.With("component", component) }
func (l *Logger) SortFilesForTest(names []string) []string { sort.Strings(names); return names }

func (l *Logger) ConfigureStandardLog() {
	log.SetOutput(l.appWriter)
	log.SetFlags(0)
}
