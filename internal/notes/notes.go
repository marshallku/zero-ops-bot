package notes

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Store struct {
	baseDir string
	mu      sync.RWMutex
}

func NewStore(baseDir string) (*Store, error) {
	for _, dir := range []string{
		filepath.Join(baseDir, "daily"),
		filepath.Join(baseDir, "categories"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return &Store{baseDir: baseDir}, nil
}

func (s *Store) BaseDir() string {
	return s.baseDir
}

func (s *Store) Add(text, category string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	timestamp := now.Format("15:04")
	entry := fmt.Sprintf("- %s | %s\n", timestamp, text)

	if category == "" || category == "daily" {
		return s.appendDaily(now, entry)
	}
	return s.appendCategory(category, now, entry)
}

func (s *Store) GetToday() (string, error) {
	return s.GetByDate(time.Now().Format("2006-01-02"))
}

func (s *Store) GetByDate(date string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.baseDir, "daily", date+".md")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *Store) GetRecent(days int) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sb strings.Builder
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		path := filepath.Join(s.baseDir, "daily", date+".md")
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		sb.Write(data)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (s *Store) GetByCategory(category string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.baseDir, "categories", category+".md")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *Store) ListCategories() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(filepath.Join(s.baseDir, "categories"))
	if err != nil {
		return nil, err
	}

	var categories []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		categories = append(categories, strings.TrimSuffix(e.Name(), ".md"))
	}
	return categories, nil
}

func (s *Store) Remove(date string, index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.baseDir, "daily", date+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	noteIndex := 0
	removed := false

	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, "- ") {
			noteIndex++
			if noteIndex == index {
				removed = true
				continue
			}
		}
		result = append(result, line)
	}

	if !removed {
		return fmt.Errorf("note #%d not found", index)
	}

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

func (s *Store) Search(query string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	var results []string

	for _, dir := range []string{"daily", "categories"} {
		dirPath := filepath.Join(s.baseDir, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}

			path := filepath.Join(dirPath, e.Name())
			matches, err := searchFile(path, query)
			if err != nil {
				continue
			}

			source := strings.TrimSuffix(e.Name(), ".md")
			for _, m := range matches {
				results = append(results, fmt.Sprintf("[%s] %s", source, m))
			}
		}
	}

	return results, nil
}

func searchFile(path, query string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var matches []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "- ") && strings.Contains(strings.ToLower(line), query) {
			matches = append(matches, line)
		}
	}
	return matches, scanner.Err()
}

func (s *Store) appendDaily(date time.Time, entry string) error {
	dateStr := date.Format("2006-01-02")
	path := filepath.Join(s.baseDir, "daily", dateStr+".md")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		content := fmt.Sprintf("# %s\n\n%s", dateStr, entry)
		return os.WriteFile(path, []byte(content), 0644)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

func (s *Store) appendCategory(category string, date time.Time, entry string) error {
	path := filepath.Join(s.baseDir, "categories", category+".md")
	dateStr := date.Format("2006-01-02")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		content := fmt.Sprintf("# %s\n\n## %s\n%s", category, dateStr, entry)
		return os.WriteFile(path, []byte(content), 0644)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	dateHeader := fmt.Sprintf("## %s", dateStr)

	if strings.Contains(content, dateHeader) {
		idx := strings.Index(content, dateHeader)
		insertAt := idx + len(dateHeader) + 1
		content = content[:insertAt] + entry + content[insertAt:]
	} else {
		content += fmt.Sprintf("\n%s\n%s", dateHeader, entry)
	}

	return os.WriteFile(path, []byte(content), 0644)
}
