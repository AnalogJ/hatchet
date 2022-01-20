package pkg

import (
	"crypto/tls"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/analogj/hatchet/pkg/config"
	"github.com/analogj/hatchet/pkg/model"
	"github.com/anaskhan96/soup"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const BATCH_SIZE = 5000

type EmailEngine struct {
	logger        *logrus.Entry
	client        *client.Client
	configuration config.Interface

	progressBar *progressbar.ProgressBar
	report      map[string]*model.SenderReport
}

func New(logger *logrus.Entry, configuration config.Interface) (EmailEngine, error) {
	imap.CharsetReader = charset.Reader

	emailEngine := EmailEngine{}
	emailEngine.logger = logger
	emailEngine.report = map[string]*model.SenderReport{}
	emailEngine.configuration = configuration

	emailEngine.logger.Infoln("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(fmt.Sprintf("%s:%s", configuration.GetString("imap-hostname"), configuration.GetString("imap-port")), &tls.Config{ServerName: configuration.GetString("imap-hostname")})
	if err != nil {
		emailEngine.logger.Fatal(err)
	}
	emailEngine.client = c
	emailEngine.logger.Infoln("Connected")

	// Login
	if err := emailEngine.client.Login(configuration.GetString("imap-username"), configuration.GetString("imap-password")); err != nil {
		emailEngine.logger.Fatal(err)
	}
	emailEngine.logger.Println("Logged in")

	return emailEngine, nil
}

func (ee *EmailEngine) Start() error {
	// Don't forget to logout
	defer ee.client.Logout()

	var mailboxName string
	if ee.configuration.IsSet("imap-mailbox-name") {
		mailboxName = ee.configuration.GetString("imap-mailbox-name")
		ee.logger.Infof("Mailbox name from configuration: %s", mailboxName)
	} else {
		mailboxName = ee.getMailboxName()
		ee.logger.Infof("Mailbox name: %s", mailboxName)

	}

	//retrieve messages from mailbox
	//note: the number of messages may be absurdly large, so lets do this in batches for safety (sets of 100 messages)

	// message ranges are 1 base indexed.
	// ie, batches include messages from 1-100
	var totalMessages uint32
	totalMessages = 0
	page := 0

	// get latest mailbox information
	//https://bitmapcake.blogspot.com/2018/07/gmail-mailbox-names-for-imap-connections.html
	mbox, err := ee.client.Select(mailboxName, false)
	if err != nil {
		ee.logger.Fatal(err)
	}
	// Get count of all messages
	totalMessages = mbox.Messages

	//set a progress bar
	ee.progressBar = progressbar.Default(int64(totalMessages))

	if totalMessages == 0 {
		//if theres no messages to process, break out of the loop
		ee.logger.Printf("No messages to process")
		return errors.New("No messages to process")
	}

	for {
		// 1-500, 501-1000
		from := uint32((page * BATCH_SIZE) + 1)
		to := uint32((page + 1) * BATCH_SIZE)
		if totalMessages < to {
			to = totalMessages
		}
		page += 1

		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		ee.logger.Debugf("Retrieving messages (%d-%d, page: %d, total: %d)", from, to, page, totalMessages)
		ee.retrieveMessages(seqset)

		if totalMessages <= to {
			break
		}
		//todo publish/generate events for stored documents
		//ee.generateEvent()
	}

	return ee.Export()
}

func (ee *EmailEngine) Export() error {
	file, err := os.Create(ee.configuration.GetString("output-path"))
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)

	err = w.Write([]string{"company", "email", "count", "unsubscribe_link_oneclick", "unsubscribe_link", "unsubscribe_email", "last_msg_date", "last_msg_subject"})
	if err != nil {
		return err
	}

	for email, record := range ee.report {
		if err := w.Write([]string{
			record.CompanyName,
			email,
			strconv.FormatInt(record.MessageCount, 10),
			strconv.FormatBool(record.UnsubscribeLinkOneClick),
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

//depending on the user's language settings, the Gmail/All Mail string may differ
//instead we'll look for the \All attribute, then findind the associated Gmail inbox name
func (ee *EmailEngine) getMailboxName() string {
	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- ee.client.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		for _, attr := range m.Attributes {
			if attr == `\All` {
				return m.Name
			}
		}
	}

	// we didn't find the the `\All` attribute on any mailbox, fallback to Gmail's default
	return "[Gmail]/All Mail"
}

func (ee *EmailEngine) retrieveMessages(seqset *imap.SeqSet) {
	section := &imap.BodySectionName{Peek: true}

	// Get the whole message body
	var headerFetchInstruction imap.FetchItem = "BODY.PEEK[HEADER]"
	if ee.configuration.GetBool("fetch") {
		headerFetchInstruction = "BODY[HEADER]"
	}

	items := []imap.FetchItem{imap.FetchEnvelope, section.FetchItem(), imap.FetchUid, headerFetchInstruction}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- ee.client.Fetch(seqset, items, messages)
	}()

	for msg := range messages {
		/* read and process the email */
		err := ee.processMessage(msg)
		if err != nil {
			ee.logger.Errorf("error processing message: %v", err)
		}
		ee.progressBar.Add(1)
	}

	if err := <-done; err != nil {
		ee.logger.Fatal(err)
		return
	}
}

func (ee *EmailEngine) processMessage(msg *imap.Message) error {
	//ee.logger.Infof("%v", .Fields)

	ee.logger.Debugf("ID: %d", msg.Uid)
	ee.logger.Debugf("Env Date: %s", msg.Envelope.Date)
	ee.logger.Debugf("Env Subject: %s", msg.Envelope.Subject)
	if len(msg.Envelope.From) > 0 {
		ee.logger.Debugf("Env From: %s", msg.Envelope.From[0].Address())
	} else {
		return errors.New("No From Address provided")
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

		//get unsubscribe link
		unsubscribeUris, unsubscribeOneClick, err := ee.extractHeaderUnsubscribe(msg)
		if err != nil || len(unsubscribeUris) == 0 {
			//no unsubscribe link found in header, lets check the body
			unsubscribeUris, err = ee.extractBodyUnsubscribe(msg)
			if err != nil {
				return err
			}
		}
		senderReport.UnsubscribeLinkOneClick = unsubscribeOneClick
		for _, unsubscribeUri := range unsubscribeUris {
			if strings.HasPrefix(unsubscribeUri, "mailto:") {
				senderReport.UnsubscribeEmail = unsubscribeUri
			} else {
				senderReport.UnsubscribeLink = unsubscribeUri
			}
		}

		senderReport.LatestMessage = model.SenderMessage{
			Date:     msg.Envelope.Date,
			Subject:  cleanSubject,
			Category: "UNKNOWN", //TODO: find a waay to get Gmail categories.
		}
	}

	return nil
}

func (ee *EmailEngine) extractHeaderUnsubscribe(msg *imap.Message) ([]string, bool, error) {
	unsubscribeUris := []string{}
	unsubscribeOneClick := false

	headerSection, _ := imap.ParseBodySectionName("RFC822.HEADER")

	msgHeader := msg.GetBody(headerSection)
	if msgHeader == nil {
		ee.logger.Warnf("Failed to parse message headers")
		return nil, false, errors.New("could not parse headers")
	} else {

		messageHeaders, err := message.Read(msgHeader)
		if message.IsUnknownCharset(err) {
			ee.logger.Warnf("Unknown encoding for message headers: %v", err)
		} else if err != nil {
			return nil, false, err
		}

		if messageHeaders.Header.Has("List-Unsubscribe-Post") && strings.TrimSpace(messageHeaders.Header.Get("List-Unsubscribe-Post")) == "List-Unsubscribe=One-Click" {
			unsubscribeOneClick = true
		}
		if messageHeaders.Header.Has("List-Unsubscribe") {
			unsubscribeWrappedUris := strings.Split(messageHeaders.Header.Get("List-Unsubscribe"), ",")
			for _, unsubscribeUri := range unsubscribeWrappedUris {
				unsubscribeUris = append(unsubscribeUris, strings.Trim(unsubscribeUri, " ><"))
			}
		}
	}
	return unsubscribeUris, unsubscribeOneClick, nil

}

func (ee *EmailEngine) extractBodyUnsubscribe(msg *imap.Message) ([]string, error) {
	unsubscribeUris := []string{}

	//bodySection, _ := imap.ParseBodySectionName("BODY[TEXT]")
	//msgBody := msg.GetBody(bodySection)

	var bodySection imap.BodySectionName
	msgBody := msg.GetBody(&bodySection)

	if msgBody == nil {
		ee.logger.Warnf("Failed to parse message body")
		return nil, errors.New("could not parse body")
	}

	// Create a new mail reader
	msgBodyReader, err := mail.CreateReader(msgBody)
	if err != nil {
		return nil, errors.New("could not create body reader")
	}

	// Process each message's part
	for {
		p, err := msgBodyReader.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.New("error processing next body part")
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			//content type:
			contentTypeHeader, _, err := h.ContentType()
			if err != nil || !strings.Contains(contentTypeHeader, "html") {
				continue
			}

			// This is the message's text (should be HTML)
			msgBodyBytes, _ := ioutil.ReadAll(p.Body)
			msgBodyDoc := soup.HTMLParse(string(msgBodyBytes))
			links := msgBodyDoc.FindAll("a")
			for _, link := range links {
				//ee.logger.Debugf("body link -> [%s](%s)\n", link.Text(), link.Attrs()["href"])
				linkTextCompare := strings.ToLower(link.Text())

				if strings.Contains(linkTextCompare, "subscribe") {
					ee.logger.Debugf("body unsubscribe link -> [%s](%s)\n", link.Text(), link.Attrs()["href"])
					unsubscribeUris = append(unsubscribeUris, link.Attrs()["href"])
				}
			}
		}
	}

	return unsubscribeUris, nil
}
