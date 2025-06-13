package notify

import (
	"bytes"
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
	"github.com/ptokihery/pymon/config"
)

func SendMail(cfg config.Config,subject, body string, logContent string) {
	if !cfg.Email.Enabled {
		return
	}

	e := email.NewEmail()
	e.From = cfg.Email.Username
	e.To = cfg.Email.To
	e.Subject = fmt.Sprintf("%s : %s", cfg.Server.Name, subject)
	e.Text = []byte(body)
	e.Attach(bytes.NewReader([]byte(logContent)), "log.txt", "text/plain")

	addr := fmt.Sprintf("%s:%d", cfg.Email.SMTPServer, cfg.Email.SMTPPort)
	auth := smtp.PlainAuth("", cfg.Email.Username, cfg.Email.Password, cfg.Email.SMTPServer)

	if err := e.Send(addr, auth); err != nil {
		fmt.Println("Erreur envoi email:", err)
	}
}
