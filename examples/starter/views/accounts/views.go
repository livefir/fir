package accounts

import (
	"github.com/adnaan/authn"
	"github.com/adnaan/pineview/controller"
)

type Views struct {
	Auth *authn.API
}

func (v Views) Confirm() controller.View {
	return &ConfirmView{Auth: v.Auth}
}

func (v Views) ConfirmEmailChange() controller.View {
	return &ConfirmEmailChangeView{Auth: v.Auth}
}

func (v Views) ConfirmMagic() controller.View {
	return &ConfirmMagicView{Auth: v.Auth}
}

func (v Views) Forgot() controller.View {
	return &ForgotView{Auth: v.Auth}
}

func (v Views) Login() controller.View {
	return &LoginView{Auth: v.Auth}
}

func (v Views) Reset() controller.View {
	return &ResetView{Auth: v.Auth}
}

func (v Views) Settings() controller.View {
	return &SettingsView{Auth: v.Auth}
}

func (v Views) Signup() controller.View {
	return &SignupView{Auth: v.Auth}
}
