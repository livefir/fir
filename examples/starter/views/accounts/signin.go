package accounts

import (
	"errors"
	"fmt"
	"log"
	"net/http"

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

func (s *SigninView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "magic-login":
		r := new(ProfileRequest)
		if err := event.DecodeFormParams(r); err != nil {
			return fir.PatchError(err)
		}
		if r.Email == "" {
			return fir.PatchError(fmt.Errorf("%w", errors.New("email is required")))
		}
		if err := s.Auth.SendPasswordlessToken(event.RequestContext(), r.Email); err != nil {
			return fir.PatchError(err)
		}

		return fir.Patchset{fir.Morph{
			Selector: "#signin_container",
			HTML: &fir.Render{
				Template: "signin_container",
				Data:     map[string]any{"sent_magic_link": true},
			},
		}}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}

func (s *SigninView) OnGet(w http.ResponseWriter, r *http.Request) fir.Page {
	if _, err := s.Auth.CurrentAccount(r); err != nil {
		return fir.Page{}
	}

	return fir.Page{
		Data: map[string]any{
			"is_logged_in": true,
		},
	}
}

func (s *SigninView) OnPost(w http.ResponseWriter, r *http.Request) fir.Page {
	return s.LoginSubmit(w, r)
}

func (s *SigninView) LoginSubmit(w http.ResponseWriter, r *http.Request) fir.Page {
	var email, password string
	_ = r.ParseForm()
	for k, v := range r.Form {
		if k == "email" && len(v) == 0 {
			return fir.PageError(fmt.Errorf("email is required"))
		}

		if k == "password" && len(v) == 0 {
			return fir.PageError(fmt.Errorf("password is required"))
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
		return fir.PageError(err, "unauthorized")
	}
	redirectTo := "/app"
	from := r.URL.Query().Get("from")
	if from != "" {
		redirectTo = from
	}

	http.Redirect(w, r, redirectTo, http.StatusSeeOther)

	return fir.Page{}
}
