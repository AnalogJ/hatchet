package pkg

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"github.com/analogj/hatchet/pkg/model"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"regexp"
	"strconv"
)

const BATCH_SIZE = 500

type EmailEngine struct {
	logger *logrus.Entry
	client *client.Client

	reportPath string
	report     map[string]*model.SenderReport
}

func New(logger *logrus.Entry, config map[string]string) (EmailEngine, error) {
	emailEngine := EmailEngine{}
	emailEngine.logger = logger
	emailEngine.report = map[string]*model.SenderReport{}
	emailEngine.reportPath = config["output-path"]

	emailEngine.logger.Infoln("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(fmt.Sprintf("%s:%s", config["imap-hostname"], config["imap-port"]), &tls.Config{ServerName: config["imap-hostname"]})
	if err != nil {
		emailEngine.logger.Fatal(err)
	}
	emailEngine.client = c
	emailEngine.logger.Infoln("Connected")

	// Login
	if err := emailEngine.client.Login(config["imap-username"], config["imap-password"]); err != nil {
		emailEngine.logger.Fatal(err)
	}
	emailEngine.logger.Println("Logged in")

	return emailEngine, nil
}

func (ee *EmailEngine) Start() error {
	// Don't forget to logout
	defer ee.client.Logout()

	//retrieve messages from mailbox
	//note: the number of messages may be absurdly large, so lets do this in batches for safety (sets of 100 messages)

	// message ranges are 1 base indexed.
	// ie, batches include messages from 1-100

	page := 0
	for {
		// get lastest mailbox information
		//https://bitmapcake.blogspot.com/2018/07/gmail-mailbox-names-for-imap-connections.html
		mbox, err := ee.client.Select("[Gmail]/All Mail", false)
		if err != nil {
			ee.logger.Fatal(err)
		}
		// Get all messages
		if mbox.Messages == 0 {
			//if theres no messages to process, break out of the loop
			ee.logger.Printf("No messages to process")
			break
		}
		// 1-500, 501-1000
		from := uint32((page * BATCH_SIZE) + 1)
		to := uint32((page + 1) * BATCH_SIZE)
		if mbox.Messages < to {
			to = mbox.Messages
		}
		page += 1

		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		ee.logger.Printf("Retrieving messages (%d-%d, page: %d)", from, to, page)
		ee.retrieveMessages(seqset)

		if mbox.Messages <= to {
			break
		}
		//todo publish/generate events for stored documents
		//ee.generateEvent()
	}

	//reportJson, _ := json.MarshalIndent(ee.report, "", "    ")
	//
	//ee.logger.Infof("Report: %v", string(reportJson))

	return ee.Export()
}

func (ee *EmailEngine) Export() error {
	file, err := os.Create(ee.reportPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)

	err = w.Write([]string{"company", "email", "count", "unsubscribe_link", "unsubscribe_email", "last_msg_date", "last_msg_subject"})
	if err != nil {
		return err
	}

	for email, record := range ee.report {
		if err := w.Write([]string{
			record.CompanyName,
			email,
			strconv.FormatInt(record.MessageCount, 10),
			record.UnsubscribeLink,
			record.UnsubscribeEmail,
			record.LatestMessage.Date.String(),
			record.LatestMessage.Subject,
		}); err != nil {
			return err
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func (ee *EmailEngine) retrieveMessages(seqset *imap.SeqSet) {
	section := &imap.BodySectionName{Peek: true}

	// Get the whole message body
	items := []imap.FetchItem{imap.FetchEnvelope, section.FetchItem(), imap.FetchUid, "BODY.PEEK[HEADER]"}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- ee.client.Fetch(seqset, items, messages)
	}()

	for msg := range messages {
		/* read and process the email */
		ee.processMessage(msg)

	}

	if err := <-done; err != nil {
		ee.logger.Fatal(err)
	}
}

func (ee *EmailEngine) processMessage(msg *imap.Message) error {
	//ee.logger.Infof("%v", .Fields)

	ee.logger.Debugf("ID: %d", msg.Uid)
	ee.logger.Debugf("Env Date: %s", msg.Envelope.Date)
	ee.logger.Debugf("Env Subject: %s", msg.Envelope.Subject)
	ee.logger.Debugf("Env From: %s", msg.Envelope.From[0].Address())
	headerSection, _ := imap.ParseBodySectionName("RFC822.HEADER")

	msgHeader := msg.GetBody(headerSection)
	if msgHeader == nil {
		ee.logger.Warnf("Failed to parse message headers")
	} else {

		//headerBody, _ := ioutil.ReadAll(msgHeader)
		//ee.logger.Debugf("Header Body: %s", string(headerBody))

		//var r io.Reader

		messageHeader, err := message.Read(msgHeader)
		if message.IsUnknownCharset(err) {
			// This error is not fatal
			log.Println("Unknown encoding:", err)
		} else if err != nil {
			log.Fatal(err)
		}

		ee.logger.Debugf("Unsubscribe Header: %s", messageHeader.Header.Get("List-Unsubscribe"))
		//imap.ParseAddressList()

	}

	// Aggregate/store some info about the message
	fromList := msg.Envelope.From
	if len(fromList) > 1 {
		ee.logger.Warnf("More than 1 from address detected, only using first: %v", fromList)
	}

	from := msg.Envelope.From[0]
	if _, ok := ee.report[from.Address()]; !ok {
		ee.report[from.Address()] = &model.SenderReport{
			CompanyName:   from.PersonalName,
			Email:         from.Address(),
			MessageCount:  0,
			LatestMessage: model.SenderMessage{},
		}
	}
	senderReport := ee.report[from.Address()]
	ee.report[from.Address()].MessageCount += 1

	//check if the current message is newer than the stored message for this sender
	if (senderReport.LatestMessage == model.SenderMessage{}) || senderReport.LatestMessage.Date.Before(msg.Envelope.Date) {
		// latest message is unset, or the current message is newer than the latest message for this sender.

		//get clean subject
		// Make a Regex to say we only want letters and numbers
		reg, err := regexp.Compile("[^a-zA-Z0-9 _-]+")
		if err != nil {
			log.Fatal(err)
		}
		cleanSubject := reg.ReplaceAllString(msg.Envelope.Subject, "")

		// get category
		//TODO:

		////get unsubscribe link
		//body := bufio.NewReader(bytes.NewReader(msg.Body))
		//hdr, err := textproto.ReadHeader(body)
		//unsubscribeLink := msg.Envelope.Get("List-UnsubscribeLink")
		//if len(unsubscribeLink) > 0 {
		//	senderReport.UnsubscribeLink = unsubscribeLink
		//} else {
		//	senderReport.UnsubscribeLink = "UNKNOWN LINK"
		//}

		senderReport.LatestMessage = model.SenderMessage{
			Date:     msg.Envelope.Date,
			Subject:  cleanSubject,
			Category: "UNKNOWN", //TODO: find a waay to get Gmail categories.
		}
	}

	return nil
}
