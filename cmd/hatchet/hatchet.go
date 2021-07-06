package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/analogj/hatchet/pkg"
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

					if c.Bool("debug") {
						logrus.SetLevel(logrus.DebugLevel)
					} else {
						logrus.SetLevel(logrus.InfoLevel)
					}

					emailEngine, err := pkg.New(reportLogger, map[string]string{
						"imap-hostname": c.String("imap-hostname"),
						"imap-port":     c.String("imap-port"),
						"imap-username": c.String("imap-username"),
						"imap-password": c.String("imap-password"),
						"output-path":   c.String("output-path"),
					})
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
						Value: "993",
					},
					&cli.StringFlag{
						Name:  "imap-username",
						Usage: "The imap server username",
					},
					&cli.StringFlag{
						Name:  "imap-password",
						Usage: "The imap server password",
					},

					&cli.StringFlag{
						Name:  "output-path",
						Value: "sender_report.csv",
						Usage: "Path to output file",
					},

					&cli.BoolFlag{
						Name:  "debug",
						Usage: "Enable debug logging",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(color.HiRedString("ERROR: %v", err))
	}
}
