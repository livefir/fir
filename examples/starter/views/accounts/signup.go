package accounts

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/adnaan/authn"
	pwc "github.com/adnaan/fir/controller"
)

type SignupView struct {
	pwc.DefaultView
	Auth *authn.API
}

func (s *SignupView) Content() string {
	return "./templates/views/accounts/signup"
}

func (s *SignupView) Layout() string {
	return "./templates/layouts/index.html"
}

func (s *SignupView) OnEvent(st pwc.Socket) error {
	switch st.Event().ID {
	case "auth/signup":
		return s.Signup(st)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", st.Event())
	}
	return nil
}

func (s *SignupView) OnRequest(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.Data) {
	if _, err := s.Auth.CurrentAccount(r); err != nil {
		return pwc.Status{Code: 200}, nil
	}

	return pwc.Status{Code: 200}, pwc.Data{
		"is_logged_in": true,
	}
}

type ProfileRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *SignupView) Signup(st pwc.Socket) error {
	st.Store().UpdateProp("show_loading_modal", true)
	defer func() {
		st.Store().UpdateProp("show_loading_modal", false)
	}()
	req := new(ProfileRequest)
	if err := st.Event().DecodeParams(req); err != nil {
		return err
	}

	if req.Email == "" {
		return fmt.Errorf("%w", errors.New("email is required"))
	}
	if req.Password == "" {
		return fmt.Errorf("%w", errors.New("password is required"))
	}

	attributes := make(map[string]interface{})
	attributes["name"] = req.Name

	if err := s.Auth.Signup(st.Request().Context(), req.Email, req.Password, attributes); err != nil {
		return err
	}
	st.Morph("#signup_container", "signup_container", pwc.Data{
		"sent_confirmation": true,
	})
	return nil
}
