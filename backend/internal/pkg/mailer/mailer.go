package mailer

import (
	"fmt"
	"net/smtp"
)

// Mailer mengirim email. Interface sederhana supaya job (Phase 08/10) tidak perlu
// tahu detail SMTP — cukup panggil Send.
type Mailer interface {
	Send(to []string, subject, body string) error
}

type smtpMailer struct {
	host     string
	port     string
	user     string
	password string
	from     string
}

func NewSMTPMailer(host, port, user, password, from string) Mailer {
	return &smtpMailer{host: host, port: port, user: user, password: password, from: from}
}

func (m *smtpMailer) Send(to []string, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	// Auth opsional — sandbox/relay SMTP lokal (Mailhog, aiosmtpd debug server, dst)
	// sering tidak advertise AUTH sama sekali; smtp.SendMail akan gagal kalau kita
	// paksa PlainAuth padahal server tidak menawarkannya.
	var auth smtp.Auth
	if m.user != "" && m.password != "" {
		auth = smtp.PlainAuth("", m.user, m.password, m.host)
	}

	msg := fmt.Appendf(nil, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		m.from, joinAddrs(to), subject, body)

	return smtp.SendMail(addr, auth, m.from, to, msg)
}

func joinAddrs(addrs []string) string {
	out := ""
	for i, a := range addrs {
		if i > 0 {
			out += ", "
		}
		out += a
	}
	return out
}
