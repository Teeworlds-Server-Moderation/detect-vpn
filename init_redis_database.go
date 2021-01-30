package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jxsl13/goripr"
)

// Use this to add blacklist domains and remove whitelisted domains afterwards
func updateRedisDatabase(rdb *goripr.Client, cfg *Config) error {
	blacklistPath := path.Join(cfg.DataPath, cfg.BlacklistFolder)
	whitelistPath := path.Join(cfg.DataPath, cfg.WhitelistFolder)

	err := filepath.Walk(blacklistPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
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
		if err != nil {
			if !errors.Is(err, goripr.ErrInvalidRange) {
				log.Println(err, "Skipped invalid range:", ip)
			}
			continue
		}
	}
	return nil
}

func removeIPsFromDatabase(rdb *goripr.Client, filename string, cfg *Config) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip, reason, err := parseLine(scanner.Text())
		if err != nil {
			continue
		}
		if reason == "" {
			reason = cfg.VPNDefaultBanReason
		}

		err = rdb.Remove(ip)
		if err != nil {
			if !errors.Is(err, goripr.ErrInvalidRange) {
				log.Println(err, "Skipped invalid range:", ip)
			}
			continue
		}
	}
	return nil
}
