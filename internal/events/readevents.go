package events

import (
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
)

// calendar event cache
var (
	calCache      []models.Event
	calCacheAt    time.Time
	calCacheMu    sync.Mutex
	calRefreshing bool
	calCacheTTL   = time.Hour
)

// dialFTP opens and authenticates an FTP connection.
func dialFTP() (*ftp.ServerConn, error) {
	c, err := ftp.Dial(os.Getenv("FTP_HOST"), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	if err = c.Login(os.Getenv("FTP_USER"), os.Getenv("FTP_PASS")); err != nil {
		return nil, err
	}
	return c, nil
}

// listMarkdownPaths walks the FTP root and returns all .md file paths.
func listMarkdownPaths(c *ftp.ServerConn) ([]string, error) {
	var paths []string
	walker := c.Walk("/")
	for walker.Next() {
		if strings.HasSuffix(walker.Path(), ".md") {
			paths = append(paths, walker.Path())
		}
	}
	if err := walker.Err(); err != nil && err != io.EOF {
		return nil, err
	}
	return paths, nil
}

// fetchEvent downloads and parses a single .md file from an open FTP connection.
func fetchEvent(c *ftp.ServerConn, path string) (models.Event, error) {
	var e models.Event
	slug := strings.Split(path, "/")[1]
	e.URL = "https://prozessor-aarau.ch/kalender/" + slug

	r, err := c.Retr(path)
	if err != nil {
		return e, err
	}
	file, err := io.ReadAll(r)
	if err != nil {
		return e, err
	}
	if err = r.Close(); err != nil {
		return e, err
	}
	_, err = frontmatter.MustParse(strings.NewReader(string(file)), &e)
	return e, err
}

// fetchConcurrently downloads all paths using a pool of FTP connections.
func fetchConcurrently(paths []string, workers int) ([]models.Event, error) {
	type result struct {
		event models.Event
		err   error
	}

	pathCh := make(chan string, len(paths))
	for _, p := range paths {
		pathCh <- p
	}
	close(pathCh)

	resultCh := make(chan result, len(paths))

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := dialFTP()
			if err != nil {
				resultCh <- result{err: err}
				return
			}
			defer c.Quit()
			for path := range pathCh {
				e, err := fetchEvent(c, path)
				resultCh <- result{event: e, err: err}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var events []models.Event
	for r := range resultCh {
		if r.err != nil {
			return nil, r.err
		}
		events = append(events, r.event)
	}
	return events, nil
}

// GetEvents returns published upcoming events (next 2-3 months).
// Uses a single FTP connection for minimal connection overhead.
func GetEvents() ([]models.Event, error) {
	if err := godotenv.Load(); err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	c, err := dialFTP()
	if err != nil {
		return nil, err
	}
	defer c.Quit()

	now := time.Now()
	y, m, d := now.Date()
	if d < 10 {
		m += 2
	} else {
		m += 3
	}
	maxDate := time.Date(y, m, 0, 0, 0, 0, 0, time.Local)

	walker := c.Walk("/")
	var events []models.Event
	for walker.Next() {
		if !strings.HasSuffix(walker.Path(), ".md") {
			continue
		}
		e, err := fetchEvent(c, walker.Path())
		if err != nil {
			return nil, err
		}
		if e.Published && e.StartTime.After(now) && e.StartTime.Before(maxDate) {
			events = append(events, e)
		}
	}
	if err := walker.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime.Time.Before(events[j].StartTime.Time)
	})
	return events, nil
}

// GetCalendarEvents returns all published events with a 1-hour TTL cache.
// Stale data is served immediately while a background refresh runs.
func GetCalendarEvents() ([]models.Event, error) {
	calCacheMu.Lock()
	if len(calCache) > 0 && time.Since(calCacheAt) < calCacheTTL {
		result := make([]models.Event, len(calCache))
		copy(result, calCache)
		calCacheMu.Unlock()
		return result, nil
	}
	if len(calCache) > 0 {
		result := make([]models.Event, len(calCache))
		copy(result, calCache)
		if !calRefreshing {
			calRefreshing = true
			go refreshCalendarCache()
		}
		calCacheMu.Unlock()
		return result, nil
	}
	calCacheMu.Unlock()
	return refreshCalendarCache()
}

func refreshCalendarCache() ([]models.Event, error) {
	if err := godotenv.Load(); err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	c, err := dialFTP()
	if err != nil {
		calCacheMu.Lock()
		calRefreshing = false
		calCacheMu.Unlock()
		return nil, err
	}

	paths, err := listMarkdownPaths(c)
	if err != nil {
		calCacheMu.Lock()
		calRefreshing = false
		calCacheMu.Unlock()
		return nil, err
	}
	c.Quit()

	all, err := fetchConcurrently(paths, 5)
	if err != nil {
		calCacheMu.Lock()
		calRefreshing = false
		calCacheMu.Unlock()
		return nil, err
	}

	var events []models.Event
	for _, e := range all {
		if e.Published {
			events = append(events, e)
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime.Time.Before(events[j].StartTime.Time)
	})

	calCacheMu.Lock()
	calCache = events
	calCacheAt = time.Now()
	calRefreshing = false
	calCacheMu.Unlock()

	return events, nil
}
