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

func (s *SignupView) OnLiveEvent(ctx pwc.Context) error {
	switch ctx.Event().ID {
	case "auth/signup":
		return s.Signup(ctx)
	default:
		log.Printf("warning:handler not found for event => \n %+v\n", ctx.Event())
	}
	return nil
}

func (s *SignupView) OnMount(w http.ResponseWriter, r *http.Request) (pwc.Status, pwc.M) {
	if _, err := s.Auth.CurrentAccount(r); err != nil {
		return pwc.Status{Code: 200}, nil
	}

	return pwc.Status{Code: 200}, pwc.M{
		"is_logged_in": true,
	}
}

type ProfileRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *SignupView) Signup(ctx pwc.Context) error {
	ctx.Store().UpdateProp("show_loading_modal", true)
	defer func() {
		ctx.Store().UpdateProp("show_loading_modal", false)
	}()
	req := new(ProfileRequest)
	if err := ctx.Event().DecodeParams(req); err != nil {
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

	if err := s.Auth.Signup(ctx.Request().Context(), req.Email, req.Password, attributes); err != nil {
		return err
	}
	ctx.Morph("#signup_container", "signup_container", pwc.M{
		"sent_confirmation": true,
	})
	return nil
}
