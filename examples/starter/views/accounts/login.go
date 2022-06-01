package accounts

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/fir/controller"
)

type LoginView struct {
	pwc.DefaultView
	Auth *authn.API
}

func (l *LoginView) Content() string {
	return "./templates/views/accounts/login"
}

func (l *LoginView) Layout() string {
	return "./templates/layouts/index.html"
}

func (l *LoginView) OnEvent(s pwc.Socket) error {
	s.Store().UpdateProp("show_loading_modal", true)
	defer func() {
		s.Store().UpdateProp("show_loading_modal", false)
	}()
	switch s.Event().ID {
	case "auth/magic-login":
		return l.MagicLogin(s)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", s.Event())
	}
	return nil
}

func (l *LoginView) OnRequest(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.Data) {
	if r.Method == "POST" {
		return l.LoginSubmit(w, r)
	}

	if _, err := l.Auth.CurrentAccount(r); err != nil {
		return pwc.Status{Code: 200}, nil
	}

	return pwc.Status{Code: 200}, pwc.Data{
		"is_logged_in": true,
	}
}

func (l *LoginView) LoginSubmit(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.Data) {
	var email, password string
	_ = r.ParseForm()
	for k, v := range r.Form {
		if k == "email" && len(v) == 0 {
			return pwc.Status{Code: 200}, pwc.Data{
				"error": "email is required",
			}
		}

		if k == "password" && len(v) == 0 {
			return pwc.Status{Code: 200}, pwc.Data{
				"error": "password is required",
			}
		}

		if len(v) == 0 {
			continue
		}

		if k == "email" && len(v) > 0 {
			email = v[0]
			continue
		}

		if k == "password" && len(v) > 0 {
			password = v[0]
			continue
		}
	}
	if err := l.Auth.Login(w, r, email, password); err != nil {
		return pwc.Status{Code: 200}, pwc.Data{
			"error": pwc.UserError(err),
		}
	}
	redirectTo := "/app"
	from := r.URL.Query().Get("from")
	if from != "" {
		redirectTo = from
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)

	return pwc.Status{Code: 200}, pwc.Data{}
}

func (l *LoginView) MagicLogin(s pwc.Socket) error {
	r := new(ProfileRequest)
	if err := s.Event().DecodeParams(r); err != nil {
		return err
	}
	if r.Email == "" {
		return fmt.Errorf("%w", errors.New("email is required"))
	}
	if err := l.Auth.SendPasswordlessToken(s.Request().Context(), r.Email); err != nil {
		return err
	}
	s.Morph("#signin_container", "signin_container", pwc.Data{"sent_magic_link": true})
	return nil
}
