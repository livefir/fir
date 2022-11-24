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
	case "signup-form":
		req := new(ProfileRequest)
		if err := event.DecodeFormParams(req); err != nil {
			return fir.PatchError(err)
		}

		if req.Email == "" {
			return fir.PatchError(fmt.Errorf("%w", errors.New("email is required")))
		}

		if req.Password == "" {
			return fir.PatchError(fmt.Errorf("%w", errors.New("password is required")))
		}

		attributes := make(map[string]interface{})
		attributes["name"] = req.Name

		if err := s.Auth.Signup(event.RequestContext(), req.Email, req.Password, attributes); err != nil {
			return nil
		}
		return fir.Patchset{fir.Morph{
			Selector: "#signup_container",
			HTML: &fir.Render{
				Template: "signup_container",
				Data: map[string]any{
					"sent_confirmation": true,
				}},
		}}
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", event)
	}
	return nil
}

func (s *SignupView) OnGet(w http.ResponseWriter, r *http.Request) fir.Pagedata {
	if _, err := s.Auth.CurrentAccount(r); err != nil {
		return fir.Pagedata{}
	}

	return fir.Pagedata{Data: map[string]any{
		"is_logged_in": true,
	},
	}
}

type ProfileRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
