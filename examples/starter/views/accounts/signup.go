package accounts

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
)

type SignupView struct {
	fir.DefaultView
	Auth *authn.API
}

func (s *SignupView) Content() string {
	return "./templates/views/accounts/signup"
}

func (s *SignupView) Layout() string {
	return "./templates/layouts/index.html"
}

func (s *SignupView) OnEvent(event fir.Event) fir.Patchset {
	switch event.ID {
	case "auth/signup":
		req := new(ProfileRequest)
		if err := event.DecodeParams(req); err != nil {
			return fir.Error(err)
		}

		if req.Email == "" {
			return fir.Error(fmt.Errorf("%w", errors.New("email is required")))
		}

		if req.Password == "" {
			return fir.Error(fmt.Errorf("%w", errors.New("password is required")))
		}

		attributes := make(map[string]interface{})
		attributes["name"] = req.Name

		if err := s.Auth.Signup(event.RequestContext(), req.Email, req.Password, attributes); err != nil {
			return nil
		}
		return fir.Patchset{fir.Morph{
			Selector: "#signup_container",
			Template: &fir.Template{
				Name: "signup_container",
				Data: fir.Data{
					"sent_confirmation": true,
				}},
		}}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}

func (s *SignupView) OnRequest(w http.ResponseWriter, r *http.Request) (fir.Status, fir.Data) {
	if _, err := s.Auth.CurrentAccount(r); err != nil {
		return fir.Status{Code: 200}, nil
	}

	return fir.Status{Code: 200}, fir.Data{
		"is_logged_in": true,
	}
}

type ProfileRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
