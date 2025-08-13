package laminate

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/pathologize"
)

// Cache manages the cache for laminate
type Cache struct {
	dir      string
	duration time.Duration
}

// NewCache creates a new cache instance
func NewCache(duration time.Duration) *Cache {
	return &Cache{
		dir:      getCachePath(),
		duration: duration,
	}
}

// Get retrieves cached data if it exists and is not expired
func (c *Cache) Get(lang, input, ext string) ([]byte, bool) {
	if c.duration == 0 {
		return nil, false
	}
	cachePath := c.getCacheFilePath(lang, input, ext)

	// Check if cache file exists
	info, err := os.Stat(cachePath)
	if err != nil {
		return nil, false
	}

	// Check if cache is expired
	if time.Since(info.ModTime()) > c.duration {
		return nil, false
	}

	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}
	return data, true
}

// Set stores data in cache
func (c *Cache) Set(lang, input, ext string, data []byte) error {
	if c.duration == 0 {
		return nil
	}
	cachePath := c.getCacheFilePath(lang, input, ext)

	// Create cache directory if it doesn't exist
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	// Write cache file
	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	return nil
}

// getCacheFilePath returns the cache file path for given parameters
func (c *Cache) getCacheFilePath(lang, input, ext string) string {
	// Calculate MD5 hash of input
	hash := md5.Sum([]byte(input))
	hashStr := fmt.Sprintf("%x", hash)

	// Sanitize lang for filesystem safety
	safeLang := pathologize.Clean(lang)

	// Build cache file path: {{lang}}/{{hash(input)}}.{{ext}}
	return filepath.Join(c.dir, safeLang, hashStr+"."+ext)
}

// Clean removes expired cache files
func (c *Cache) Clean() error {
	if c.duration == 0 {
		return nil
	}
	now := time.Now()
	err := filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && now.Sub(info.ModTime()) > c.duration {
			os.Remove(path) // Ignore errors
		}
		return nil
	})
	return err
}
