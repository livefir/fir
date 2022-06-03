package accounts

import (
	"github.com/adnaan/authn"
	"github.com/adnaan/fir"
)

type Views struct {
	Auth *authn.API
}

func (v Views) Confirm() fir.View {
	return &ConfirmView{Auth: v.Auth}
}

func (v Views) ConfirmEmailChange() fir.View {
	return &ConfirmEmailChangeView{Auth: v.Auth}
}

func (v Views) ConfirmMagic() fir.View {
	return &ConfirmMagicView{Auth: v.Auth}
}

func (v Views) Forgot() fir.View {
	return &ForgotView{Auth: v.Auth}
}

func (v Views) Login() fir.View {
	return &LoginView{Auth: v.Auth}
}

func (v Views) Reset() fir.View {
	return &ResetView{Auth: v.Auth}
}

func (v Views) Settings() fir.View {
	return &SettingsView{Auth: v.Auth}
}

func (v Views) Signup() fir.View {
	return &SignupView{Auth: v.Auth}
}
