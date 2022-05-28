package config

import (
	"fmt"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/adnaan/authn"

	"github.com/jordan-wright/email"

	"github.com/matcornic/hermes/v2"
)

func SendEmailFunc(cfg Config) authn.SendMailFunc {
	appName := strings.Title(strings.ToLower(cfg.Name))
	h := hermes.Hermes{
		Product: hermes.Product{
			Name: appName,
			Link: cfg.Domain,
			//Logo: "https://github.com/matcornic/hermes/blob/master/examples/gopher.png?raw=true",
		},
	}

	pool := newEmailPool(cfg)
	return func(mailType authn.MailType, token, sendTo string, metadata map[string]interface{}) error {
		var name string
		var ok bool
		if metadata["name"] != nil {
			name, ok = metadata["name"].(string)
			if !ok {
				name = ""
			}
		}

		var emailTmpl hermes.Email
		var subject string

		switch mailType {
		case authn.Confirmation:
			subject = fmt.Sprintf("Welcome to %s!", appName)
			emailTmpl = confirmation(appName, name, fmt.Sprintf("%s/confirm/%s", cfg.Domain, token))
		case authn.Recovery:
			subject = fmt.Sprintf("Reset password on %s", appName)
			emailTmpl = recovery(appName, name, fmt.Sprintf("%s/reset/%s", cfg.Domain, token))
		case authn.ChangeEmail:
			subject = fmt.Sprintf("Change email on %s", appName)
			emailTmpl = changeEmail(appName, name, fmt.Sprintf("%s/account/email/change/%s", cfg.Domain, token))
		case authn.Passwordless:
			subject = fmt.Sprintf("Magic link to log into %s", appName)
			emailTmpl = magic(appName, name, fmt.Sprintf("%s/magic-login/%s", cfg.Domain, token))
		}

		res, err := h.GenerateHTML(emailTmpl)
		if err != nil {
			return err
		}

		e := &email.Email{
			To:      []string{sendTo},
			Subject: subject,
			HTML:    []byte(res),
			Headers: textproto.MIMEHeader{},
			From:    cfg.SMTPAdminEmail,
		}

		return pool.Send(e, 20*time.Second)
	}
}

func confirmation(appName, name, link string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				fmt.Sprintf("Welcome to %s! We're very excited to have you on board.", appName),
			},
			Actions: []hermes.Action{
				{
					Instructions: fmt.Sprintf("To get started with %s, please click here:", appName),
					Button: hermes.Button{
						Text: "Confirm your account",
						Link: link,
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}
}

func changeEmail(appName, name, link string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				fmt.Sprintf("You have received this email because you have requested to change the email linked to your %s account", appName),
			},
			Actions: []hermes.Action{
				{
					Instructions: fmt.Sprintf("Click the button below to change the email linked to your %s account", appName),
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Confirm email change",
						Link:  link,
					},
				},
			},
			Outros: []string{
				"If you did not request a email change, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

func recovery(appName, name, link string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				fmt.Sprintf("You have received this email because a password reset request for %s account was received.", appName),
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to reset your password:",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Reset your password",
						Link:  link,
					},
				},
			},
			Outros: []string{
				"If you did not request a password reset, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

func magic(appName, name, link string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				fmt.Sprintf("You have received this email because a request for a magic login link for your %s account was received.", appName),
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to login:",
					Button: hermes.Button{
						Text: "Login with magic link",
						Link: link,
					},
				},
			},
			Outros: []string{
				"If you did not request a magic login link, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

func newEmailPool(cfg Config) *email.Pool {

	var pool *email.Pool
	var err error

	if cfg.SMTPDebug {
		pool, err = email.NewPool(
			fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
			10, &unencryptedAuth{
				smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)},
		)

		if err != nil {
			panic(err)
		}

		return pool
	}

	pool, err = email.NewPool(
		fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		10,
		smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost),
	)

	if err != nil {
		panic(err)
	}

	return pool
}

type unencryptedAuth struct {
	smtp.Auth
}

// Start starts the auth process for the specified SMTP server.
func (u *unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	server.TLS = true
	return u.Auth.Start(server)
}
