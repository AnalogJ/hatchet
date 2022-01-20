package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/analogj/hatchet/pkg"
	"github.com/analogj/hatchet/pkg/config"
	"github.com/analogj/hatchet/pkg/version"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

var goos string
var goarch string

func main() {

	configuration, err := config.Create()
	if err != nil {
		fmt.Printf("FATAL: %+v\n", err)
		os.Exit(1)
	}
	//we're going to load the config file manually, since we need to validate it.
	_ = configuration.ReadConfig("hatchet.yaml") // Find and read the config file if it exists

	app := &cli.App{
		Name:     "hatchet",
		Usage:    "Cut down spam in your Gmail Inbox ",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			cli.Author{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			capsuleUrl := "AnalogJ/hatchet"

			versionInfo := fmt.Sprintf("%s.%s-%s", goos, goarch, version.VERSION)

			subtitle := capsuleUrl + utils.LeftPad2Len(versionInfo, " ", 53-len(capsuleUrl))

			fmt.Fprintf(c.App.Writer, fmt.Sprintf(utils.StripIndent(
				`
			 _   _    __   ____  ___  _   _  ____  ____ 
			( )_( )  /__\ (_  _)/ __)( )_( )( ___)(_  _)
			 ) _ (  /(__)\  )( ( (__  ) _ (  )__)   )(  
			(_) (_)(__)(__)(__) \___)(_) (_)(____) (__) 
			%s
			`), subtitle))
			return nil
		},

		Commands: []cli.Command{
			{
				Name:  "report",
				Usage: "Generate a report, listing all senders, number of emails, and most recent unsubscribe link.",
				Action: func(c *cli.Context) error {

					reportLogger := logrus.WithFields(logrus.Fields{
						"type": "email",
					})

					// map flags on-top of the configuration object
					if c.IsSet("imap-hostname") {
						configuration.Set("imap-hostname", c.String("imap-hostname"))
					}
					if c.IsSet("imap-port") {
						configuration.Set("imap-port", c.String("imap-port"))
					}
					if c.IsSet("imap-username") {
						configuration.Set("imap-username", c.String("imap-username"))
					}
					if c.IsSet("imap-password") {
						configuration.Set("imap-password", c.String("imap-password"))
					}
					if c.IsSet("imap-mailbox-name") {
						configuration.Set("imap-mailbox-name", c.String("imap-mailbox-name"))
					}
					if c.IsSet("output-path") {
						configuration.Set("output-path", c.String("output-path"))
					}
					if c.IsSet("fetch") {
						configuration.Set("fetch", c.Bool("fetch"))
					}
					if c.IsSet("debug") {
						configuration.Set("debug", c.Bool("debug"))
					}

					if configuration.GetBool("debug") {
						logrus.SetLevel(logrus.DebugLevel)
					} else {
						logrus.SetLevel(logrus.InfoLevel)
					}

					emailEngine, err := pkg.New(reportLogger, configuration)
					if err != nil {
						return err
					}
					return emailEngine.Start()
				},

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "imap-hostname",
						Usage: "The imap server hostname",
					},
					&cli.StringFlag{
						Name:  "imap-port",
						Usage: "The imap server port",
					},
					&cli.StringFlag{
						Name:     "imap-username",
						Usage:    "The imap server username",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "imap-password",
						Usage:    "The imap server password",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "imap-mailbox-name",
						Usage: "The imap mailbox to use. Defaults to Gmail's '[Gmail]/All Mail'",
					},

					&cli.StringFlag{
						Name:  "output-path",
						Usage: "Path to output file",
					},

					&cli.BoolFlag{
						Name:  "fetch",
						Usage: "Instead of PEEKing emails, update the status to each processed email to 'read'.",
					},

					&cli.BoolFlag{
						Name:  "debug",
						Usage: "Enable debug logging",
					},
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(color.HiRedString("ERROR: %v", err))
	}
}
