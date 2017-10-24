// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"net/textproto"
	"strings"
)

type (
	// Emailer is the interface for email sending.
	Emailer interface {
		From(email string)
		FromName(name string)
		ReplyTo(email string)
		To(emails ...string)
		Bcc(emails ...string)
		Subject(subject string)
		Send() error
		HTML() *BodyPart
		Plain() *BodyPart
	}

	// Email struct contains the email parameters.
	Email struct {
		engine *Engine

		fromEmail   string
		fromName    string
		replyTo     string
		to          []string
		bcc         []string
		subject     string
		html        BodyPart
		plain       BodyPart
		attachments []attachment
	}

	// BodyPart is the structure used for html and plain.
	BodyPart struct{ bytes.Buffer }

	attachment struct {
		filename string
		path     string
	}
)

// NewEmail creates a new Email instance.
func NewEmail() *Email {
	return &Email{engine: App}
}

// From sets the sender email address.
func (m *Email) From(addr string) {
	m.fromEmail = addr
}

// FromName sets the sender name.
func (m *Email) FromName(name string) {
	m.fromName = name
}

// ReplyTo sends the Reply-To email address.
func (m *Email) ReplyTo(addr string) {
	m.replyTo = addr
}

// To sets a list of recipient addresses.
//
// You can pass single or multiple addresses to this method.
//
//	mail.To("first@email.com", "second@email.com")
func (m *Email) To(emails ...string) {
	m.to = []string{}

	// for _, email := range emails {
	m.to = append(m.to, emails...)
	// }
}

// Bcc sets a list of blind carbon copy (BCC) addresses.
//
// You can pass single or multiple addresses to this method.
//
//	mail.Bcc("first@email.com", "second@email.com")
func (m *Email) Bcc(emails ...string) {
	m.bcc = []string{}

	// for _, email := range emails {
	m.bcc = append(m.bcc, emails...)
	// }
}

// Subject sets the email subject.
func (m *Email) Subject(subject string) {
	m.subject = subject
}

// ClearAttachments removes all attachments.
func (m *Email) ClearAttachments() {
	m.attachments = []attachment{}
}

// Attach adds an attachment to the email.
func (m *Email) Attach(filename, path string) {
	m.attachments = append(m.attachments, attachment{
		filename: filename,
		path:     path,
	})
}

// Send attempts to send the built email.
func (m *Email) Send() error {
	var buf bytes.Buffer
	var err error
	var alternative *multipart.Writer
	var contentType string

	// Write Headers
	if m.fromName == "" {
		fmt.Fprintf(&buf, "From: %s\r\n", m.fromEmail)
	} else {
		fmt.Fprintf(&buf, "From: %s <%s>\r\n", m.fromName, m.fromEmail)
	}

	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")

	if m.replyTo != "" {
		fmt.Fprintf(&buf, "Reply-To: %s\r\n", m.replyTo)
	}

	fmt.Fprintf(&buf, "Subject: %s\r\n", m.subject)

	for _, to := range m.to {
		fmt.Fprintf(&buf, "To: %s\r\n", to)
	}

	// Prepare multipart/mixed part message
	mb, _ := randomBoundary() // Mixed boundary
	ab, _ := randomBoundary() // Alternative boundary

	mixed := multipart.NewWriter(&buf)
	_ = mixed.SetBoundary(mb)
	defer func() {
		_ = mixed.Close()
	}()

	fmt.Fprintf(&buf, "Content-Type: multipart/mixed;\r\n\tboundary=\"%s\"; charset=UTF-8\r\n\r\n", mixed.Boundary())

	// Create alternative boundary only if we have HTML content, otherwise, we use multipart/mixed for text
	if m.html.Len() > 0 {
		alternative = multipart.NewWriter(&buf)
		_ = alternative.SetBoundary(ab)
		defer func() {
			_ = alternative.Close()
		}()

		contentType = fmt.Sprintf("multipart/alternative;\r\n\tboundary=\"%s\"", ab)
		if _, err = mixed.CreatePart(textproto.MIMEHeader{"Content-Type": {contentType}}); err != nil {
			return err
		}
	}

	// Write Body to alternative part
	writePart := func(ctype string, data []byte) {
		if len(data) == 0 || err != nil {
			return
		}

		c := fmt.Sprintf("%s; charset=UTF-8", ctype)

		var w *multipart.Writer

		if ctype == "text/plain" && m.html.Len() == 0 {
			w = mixed
		} else {
			w = alternative
		}

		part, err2 := w.CreatePart(textproto.MIMEHeader{"Content-Type": {c}})
		if err2 != nil {
			err = err2
			return
		}

		_, _ = part.Write(data)
	}

	if m.plain.Len() > 0 {
		writePart("text/plain", m.plain.Bytes())
	}

	if m.html.Len() > 0 {
		writePart("text/html", m.html.Bytes())
	}

	// Write attachments to mixed content
	for _, item := range m.attachments {
		var data []byte
		var part io.Writer

		data, err = ioutil.ReadFile(item.path)
		if err != nil {
			return err
		}

		contentType = fmt.Sprintf("%s;\n\tfilename=%s", http.DetectContentType(data[:512]), item.filename)
		contentDisposition := fmt.Sprintf("attachment;\n\tfilename=%s", item.filename)
		part, err = mixed.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {contentType},
			"Content-Disposition":       {contentDisposition},
			"Content-Transfer-Encoding": {"base64"},
		})
		if err != nil {
			return err
		}

		encoder := base64.NewEncoder(base64.StdEncoding, part)
		_, _ = encoder.Write(data)
		_ = encoder.Close()
	}

	var auth smtp.Auth

	// Choose authentication
	if Config.String("smtp_auth") == "md5" {
		auth = smtp.CRAMMD5Auth(Config.String("smtp_username"), Config.String("smtp_password"))
	} else {
		host := Config.String("smtp_server")
		if strings.Contains(host, ":") {
			host = host[:strings.Index(host, ":")]
		}
		auth = smtp.PlainAuth("", Config.String("smtp_username"), Config.String("smtp_password"), host)
	}

	// Send email
	err = smtp.SendMail(
		Config.String("smtp_server"),
		auth,
		m.fromEmail,
		append(m.to, m.bcc...),
		buf.Bytes(),
	)

	return err
}

// randomBoundary creates a random boundary string for message.
func randomBoundary() (string, error) {
	buf := make([]byte, 30)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", buf), nil
}

// HTML returns a BodyPart for the HTML email body.
func (m *Email) HTML() *BodyPart {
	return &m.html
}

// Plain returns a BodyPart for the plain-text email body.
func (m *Email) Plain() *BodyPart {
	return &m.plain
}

// Set accepts a string as the contents of a BodyPart and replaces any existing data.
func (w *BodyPart) Set(s string) {
	w.Reset()
	_, _ = w.WriteString(s)
}
