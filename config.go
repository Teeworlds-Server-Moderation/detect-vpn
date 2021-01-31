package main

import (
	"time"

	configo "github.com/jxsl13/simple-configo"
)

const (
	brokerAddressRegex = `^tcp://[a-z0-9-\.:]+:([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	redisAddressRegex  = `^[a-z0-9-\.:]+:([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`

	folderRegex  = `^[a-zA-Z0-9-]+$`
	errFolderMsg = "The folder name must only contain alphanumeric characters, no special caracters nor whitespaces and must not be empty."
)

// Config is the configuration for this microservice
type Config struct {
	BrokerAddress       string
	RedisAddress        string
	RedisPassword       string
	RedisDatabase       int
	DataPath            string
	BlacklistFolder     string
	WhitelistFolder     string
	BroadcastBans       bool
	DefaultBanCommand   string
	VPNBanDuration      time.Duration
	VPNDefaultBanReason string
}

// Name is the name of the configuration Cache
func (c *Config) Name() (name string) {
	return "Detect VPN"
}

// Options returns a list of available options that can be configured for this
// config object
func (c *Config) Options() (options configo.Options) {

	// this default value allows for local development
	// while the defalt environment value in the Dockerfile allows for overriding this
	// default value when running in a container.
	optionsList := configo.Options{
		{
			Key:           "BROKER_ADDRESS",
			Description:   "The address of the Mosquitto broker. In the container environemt it's tcp://mosquitto:1883",
			DefaultValue:  "tcp://localhost:1883",
			ParseFunction: configo.DefaultParserRegex(&c.BrokerAddress, brokerAddressRegex, "BROKER_ADDRESS must have the format: tcp://<hostname/ip>:<port>"),
		},
		{
			Key:           "REDIS_ADDRESS",
			Description:   "This is the root path that is used to store and read configuration data.",
			DefaultValue:  "localhost:6379",
			ParseFunction: configo.DefaultParserRegex(&c.RedisAddress, redisAddressRegex, "The REDIS_ADDRESS must have the following format: <hostname/ip>:<port>"),
		},
		{
			Key:           "REDIS_PASSWORD",
			Description:   "Pasword used for the redis database, can be left empty.",
			DefaultValue:  "",
			ParseFunction: configo.DefaultParserString(&c.RedisPassword),
		},
		{
			Key:           "REDIS_DB",
			Description:   "Is one of the 16 [0:15] ditinct databases that redis offers.",
			DefaultValue:  "1",
			ParseFunction: configo.DefaultParserRangesInt(&c.RedisDatabase, 0, 15),
		},
		{
			Key:           "DATA_PATH",
			Description:   "Is the root folder that contains all of the data of this service.",
			DefaultValue:  "data",
			ParseFunction: configo.DefaultParserString(&c.DataPath),
		},
		{
			Key:           "BLACKLIST_FOLDER",
			Description:   "This is a folder WITHIN the DATA_PATH that is created and used to store and retrieve blacklists",
			DefaultValue:  "blacklists",
			ParseFunction: configo.DefaultParserRegex(&c.BlacklistFolder, folderRegex, errFolderMsg),
		},
		{
			Key:           "WHITELIST_FOLDER",
			Description:   "This is a folder WITHIN the DATA_PATH that is created and used to whitelists",
			DefaultValue:  "whitelists",
			ParseFunction: configo.DefaultParserRegex(&c.WhitelistFolder, folderRegex, errFolderMsg),
		},
		{
			Key:           "VPN_BAN_REASON",
			Description:   "The default reason that is used when a specific ban range does not specify a reason with # comments",
			DefaultValue:  "VPN - https://zcat.ch/bans",
			ParseFunction: configo.DefaultParserString(&c.VPNDefaultBanReason),
		},
		{
			Key:           "VPN_BAN_DURATION",
			Description:   "The duration a VPN IP is banned by default.(e.g. 10s, 5m, 1h, 1h5m10s, 24h)",
			DefaultValue:  "24h",
			ParseFunction: configo.DefaultParserDuration(&c.VPNBanDuration),
		},
		{
			Key:           "BROADCAST_BANS",
			Description:   "If a VPN user is detected on one server, you may want to execute the ban command on all servers that are connected to this system.",
			DefaultValue:  "false",
			ParseFunction: configo.DefaultParserBool(&c.BroadcastBans),
		},
		{
			Key:           "DEFAULT_BAN_COMMAND",
			Description:   "You may use the variables {IP}, {ID}, {DURATION:MINUTES}, {DURATION:SECONDS}, {REASON}",
			DefaultValue:  "ban {IP} {DURATION:MINUTES} {REASON}",
			ParseFunction: configo.DefaultParserString(&c.DefaultBanCommand),
		},
	}

	return optionsList
}
