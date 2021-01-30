package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/jxsl13/goripr"
)

const (
	lastModifiedKey = "________________LAST_MODIFIED________________"
)

var (
	// 1: range
	// 3: reason
	splitRegex = regexp.MustCompile(`^\s*([0-9\.\-\/]+)\s*(#\s*(.*[^\s])\s*)?$`)
)

// Use this to add blacklist domains and remove whitelisted domains afterwards
func updateRedisDatabase(initRdb *redis.Client, rdb *goripr.Client, cfg *Config) error {
	blacklistPath := path.Join(cfg.DataPath, cfg.BlacklistFolder)
	whitelistPath := path.Join(cfg.DataPath, cfg.WhitelistFolder)

	err := filepath.Walk(blacklistPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		dbLastModifiedStr, err := initRdb.HGet(lastModifiedKey, path).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		// never touched before
		if dbLastModifiedStr == "" {
			fileLastModifiedStr := info.ModTime().Format(time.RFC3339)
			_, err := initRdb.HSet(lastModifiedKey, path, fileLastModifiedStr).Result()
			if err != nil {
				return fmt.Errorf("Failed to set last modified time in database for file: %s", path)
			}
			return addIPsToDatabase(rdb, path, cfg)
		}
		// we have already seen this file before
		databaseLastModified, err := time.Parse(time.RFC3339, dbLastModifiedStr)
		fileLastModified := info.ModTime()

		// file has not been modified after the last time we saw it
		// these timestamps are rather inaccurate, thus we add one second to the
		// value in order not to have some precision problems
		if !fileLastModified.Round(time.Second).After(databaseLastModified.Round(time.Second)) {
			log.Printf("File has not been modified, skipping: %s\n", path)
			return nil
		}

		// file has been modified so we need to update the database
		fileLastModifiedStr := fileLastModified.Format(time.RFC3339)
		_, err = initRdb.HSet(lastModifiedKey, fileLastModifiedStr, path).Result()
		if err != nil {
			return fmt.Errorf("failed to update last modified state of file in database: %s", path)
		}
		return addIPsToDatabase(rdb, path, cfg)
	})
	if err != nil {
		return err
	}

	err = filepath.Walk(whitelistPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		return removeIPsFromDatabase(rdb, path, cfg)
	})
	if err != nil {
		return err
	}
	return nil
}

func parseLine(line string) (ipRange, reason string, err error) {
	matches := splitRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return "", "", errors.New("empty")
	}
	return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[3]), nil
}

func addIPsToDatabase(rdb *goripr.Client, filename string, cfg *Config) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	cnt := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip, reason, err := parseLine(scanner.Text())
		if err != nil {
			continue
		}
		if reason == "" {
			reason = cfg.VPNDefaultBanReason
		}

		err = rdb.Insert(ip, reason)
		cnt++
		if err != nil {
			if !errors.Is(err, goripr.ErrInvalidRange) {
				log.Println(err, "Skipped invalid range:", ip)
			}
			continue
		}
	}
	log.Printf("Added %7d IP ranges from: %s", cnt, filename)
	return nil
}

func removeIPsFromDatabase(rdb *goripr.Client, filename string, cfg *Config) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	cnt := 0
	for scanner.Scan() {
		ip, reason, err := parseLine(scanner.Text())
		if err != nil {
			continue
		}
		if reason == "" {
			reason = cfg.VPNDefaultBanReason
		}

		err = rdb.Remove(ip)
		cnt++
		if err != nil {
			if !errors.Is(err, goripr.ErrInvalidRange) {
				log.Println(err, "Skipped invalid range:", ip)
			}
			continue
		}
	}
	log.Printf("Removed %5d potential IP ranges from: %s", cnt, filename)
	return nil
}
