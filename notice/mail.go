package notice

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/smtp"
	"regexp"
	"strconv"
)

type MailSender struct {
	smtpAddr string
	port     int
	mailUser string
	password string
}

var (
	reg             = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	mailContentType = "text/html; charset=UTF-8"
)

func NewMailSender(smtpAddr string, port int, mailUser, password string) MailSender {
	return MailSender{smtpAddr: smtpAddr, port: port, mailUser: mailUser, password: password}
}

// SendEmail send email
func (ms MailSender) SendEmail(fromName, to, subject, body string) error {
	smtpAddr := ms.smtpAddr
	port := ms.port
	mailUer := ms.mailUser
	password := ms.password

	if !MailCheck(mailUer) || !MailCheck(to) {
		return errors.New("mail check error")
	}

	addr := fmt.Sprintf("%s:%s", smtpAddr, strconv.Itoa(port))
	auth := smtp.PlainAuth("", mailUer, password, smtpAddr)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpAddr,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, smtpAddr)
	if err != nil {
		return err
	}
	defer c.Quit()

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(mailUer); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s<%s>", fromName, mailUer)
	headers["To"] = to
	headers["Subject"] = "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	headers["Content-Type"] = mailContentType

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	_, err = w.Write([]byte(message))
	return err
}

func MailCheck(mail string) bool {
	return reg.MatchString(mail)
}
