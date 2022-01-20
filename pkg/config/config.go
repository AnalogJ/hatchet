package config

import (
	"errors"
	"github.com/analogj/go-util/utils"
	"github.com/spf13/viper"
	"log"
	"os"
)

// When initializing this class the following methods must be called:
// Config.New
// Config.Init
// This is done automatically when created via the Factory.
type configuration struct {
	*viper.Viper
}

//Viper uses the following precedence order. Each item takes precedence over the item below it:
// explicit call to Set
// flag
// env
// config
// key/value store
// default

func (c *configuration) Init() error {
	c.Viper = viper.New()
	//set defaults
	c.SetDefault("imap-hostname", "imap.gmail.com")
	c.SetDefault("imap-port", "993")
	c.SetDefault("imap-username", "")
	c.SetDefault("imap-password", "")
	c.SetDefault("imap-mailbox-name", "[Gmail]/All Mail")

	c.SetDefault("output-path", "sender_report.csv")

	c.SetDefault("fetch", false)
	c.SetDefault("debug", false)

	//if you want to load a non-standard location system config file (~/drawbridge.yml), use ReadConfig
	c.SetConfigType("yaml")
	c.SetConfigName("hatchet")

	c.SetEnvPrefix("HATCHET")
	c.AutomaticEnv()

	//CLI options will be added via the `Set()` function
	return nil
}

func (c *configuration) ReadConfig(configFilePath string) error {
	configFilePath, err := utils.ExpandPath(configFilePath)
	if err != nil {
		return err
	}

	if !utils.FileExists(configFilePath) {
		log.Printf("No configuration file found at %v. Using Defaults.", configFilePath)
		return errors.New("The configuration file could not be found.")
	}

	log.Printf("Loading configuration file: %s", configFilePath)

	config_data, err := os.Open(configFilePath)
	if err != nil {
		log.Printf("Error reading configuration file: %s", err)
		return err
	}

	err = c.MergeConfig(config_data)
	if err != nil {
		return err
	}

	return c.ValidateConfig()
}

// This function ensures that the merged config works correctly.
func (c *configuration) ValidateConfig() error {
	return nil
}
