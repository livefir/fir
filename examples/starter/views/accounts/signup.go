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

func (s *SignupView) OnEvent(st fir.Socket) error {
	switch st.Event().ID {
	case "auth/signup":
		return s.Signup(st)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", st.Event())
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

func (s *SignupView) Signup(st fir.Socket) error {
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
	st.Morph("#signup_container", "signup_container", fir.Data{
		"sent_confirmation": true,
	})
	return nil
}
