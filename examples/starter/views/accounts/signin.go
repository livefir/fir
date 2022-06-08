package accounts

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
)

type SigninView struct {
	fir.DefaultView
	Auth *authn.API
}

func (s *SigninView) Content() string {
	return "./templates/views/accounts/login"
}

func (s *SigninView) Layout() string {
	return "./templates/layouts/index.html"
}

func (s *SigninView) OnEvent(st fir.Socket) error {
	switch st.Event().ID {
	case "auth/magic-login":
		return s.MagicLogin(st)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", st.Event())
	}
	return nil
}

func (s *SigninView) OnRequest(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	if r.Method == "POST" {
		return s.LoginSubmit(w, r)
	}

	if _, err := s.Auth.CurrentAccount(r); err != nil {
		return fir.Status{Code: 200}, nil
	}

	return fir.Status{Code: 200}, fir.Data{
		"is_logged_in": true,
	}
}

func (s *SigninView) LoginSubmit(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	var email, password string
	_ = r.ParseForm()
	for k, v := range r.Form {
		if k == "email" && len(v) == 0 {
			return fir.Status{Code: 200}, fir.Data{
				"error": "email is required",
			}
		}

		if k == "password" && len(v) == 0 {
			return fir.Status{Code: 200}, fir.Data{
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
	if err := s.Auth.Login(w, r, email, password); err != nil {
		return fir.Status{Code: 200}, fir.Data{
			"error": fir.UserError(err),
		}
	}
	redirectTo := "/app"
	from := r.URL.Query().Get("from")
	if from != "" {
		redirectTo = from
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)

	return fir.Status{Code: 200}, fir.Data{}
}

func (s *SigninView) MagicLogin(st fir.Socket) error {
	st.Store().UpdateProp("show_loading_modal", true)
	defer func() {
		time.Sleep(time.Second * 1)
		st.Store().UpdateProp("show_loading_modal", false)
	}()
	r := new(ProfileRequest)
	if err := st.Event().DecodeParams(r); err != nil {
		return err
	}
	if r.Email == "" {
		return fmt.Errorf("%w", errors.New("email is required"))
	}
	if err := s.Auth.SendPasswordlessToken(st.Request().Context(), r.Email); err != nil {
		return err
	}
	st.Morph("#signin_container", "signin_container", fir.Data{"sent_magic_link": true})
	return nil
}
