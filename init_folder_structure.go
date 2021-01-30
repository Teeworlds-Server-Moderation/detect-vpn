package main

import (
	"os"
	"path"
	"regexp"
)

var (
	// 1: range
	// 3: reason
	splitRegex = regexp.MustCompile(`^\s*([0-9\.\-\/]+)\s*(#\s*(.*[^\s])\s*)?$`)
)

func initFolderStructure(cfg *Config) error {
	blacklistPath := path.Join(cfg.DataPath, cfg.BlacklistFolder)
	whitelistPath := path.Join(cfg.DataPath, cfg.WhitelistFolder)

	err := os.MkdirAll(blacklistPath, 0777)
	if err != nil {
		return err
	}
	err = os.MkdirAll(whitelistPath, 0777)
	if err != nil {
		return err
	}
	return nil
}
