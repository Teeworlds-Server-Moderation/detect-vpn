package main

import (
	"os"
	"path"
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
