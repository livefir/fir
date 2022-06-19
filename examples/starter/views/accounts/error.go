package accounts

import "github.com/adnaan/fir"

func errorPatch(err error) fir.Patchset {
	return fir.Patchset{
		fir.Morph{
			Template: "fir-event-error",
			Selector: "#fir-event-error",
			Data:     fir.Data{"eventError": err.Error()},
		},
	}
}
