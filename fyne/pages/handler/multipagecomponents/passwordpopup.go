package multipagecomponents

import (
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

	"github.com/raedahgroup/dcrlibwallet"

	"github.com/raedahgroup/godcr/fyne/pages/handler/values"
	"github.com/raedahgroup/godcr/fyne/widgets"
)

type PasswordPopUpObjects struct {
	InitOnConfirmation       func(string) error
	InitOnError              func(error)
	InitOnCancel, ExtraCalls func() // ExtraCalls is called when InitOnConfirmation is called and doesnt throw an error

	Title string

	Window fyne.Window
}

func (objects *PasswordPopUpObjects) PasswordPopUp() {
	errorLabel := widgets.NewTextWithStyle(values.WrongPasswordErr, values.ErrorColor, fyne.TextStyle{}, fyne.TextAlignLeading, values.DefaultErrTextSize)
	errorLabel.Hide()

	var confirmButton *widgets.Button

	walletPassword := widget.NewPasswordEntry()
	walletPassword.SetPlaceHolder(values.SpendingPasswordText)
	walletPassword.OnChanged = func(value string) {
		if value == "" {
			confirmButton.Disable()
		} else if confirmButton.Disabled() {
			confirmButton.Enable()
		}
	}

	var sendingPasswordPopup *widget.PopUp
	var popupContent *widget.Box

	cancelLabel := canvas.NewText(values.Cancel, values.Blue)
	cancelLabel.TextStyle.Bold = true

	cancelButton := widgets.NewClickableWidget(widget.NewHBox(cancelLabel), func() {
		sendingPasswordPopup.Hide()
		objects.InitOnCancel()
	})

	confirmButton = widgets.NewButton(values.Blue, values.Confirm, func() {
		confirmButton.Disable()
		cancelButton.Disable()

		var err error
		if objects.InitOnConfirmation != nil {
			err = objects.InitOnConfirmation(walletPassword.Text)
		}

		if err != nil {
			// do not exit password popup on invalid passphrase
			if err.Error() == dcrlibwallet.ErrInvalidPassphrase {
				errorLabel.Show()
				popupContent.Refresh()
				confirmButton.Enable()
				cancelButton.Enable()
			} else {
				log.Println(err)
				sendingPasswordPopup.Hide()
				if objects.InitOnError != nil {
					objects.InitOnError(err)
				}
			}

			return
		}

		objects.ExtraCalls()
		sendingPasswordPopup.Hide()
	})
	confirmButton.SetMinSize(confirmButton.MinSize().Add(fyne.NewSize(32, 24)))
	confirmButton.SetTextStyle(fyne.TextStyle{Bold: true})
	confirmButton.Disable()

	popupContent = widget.NewHBox(
		widgets.NewHSpacer(values.SpacerSize20),
		widget.NewVBox(
			widgets.NewVSpacer(values.SpacerSize20),
			widget.NewLabelWithStyle(objects.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widgets.NewVSpacer(values.SpacerSize20),
			walletPassword,
			errorLabel,
			widgets.NewVSpacer(values.SpacerSize20),
			widget.NewHBox(layout.NewSpacer(), widgets.NewHSpacer(values.SpacerSize140), cancelButton, widgets.NewHSpacer(values.SpacerSize24), confirmButton.Container),
			widgets.NewVSpacer(values.SpacerSize20),
		),
		widgets.NewHSpacer(20),
	)

	sendingPasswordPopup = widget.NewModalPopUp(popupContent, objects.Window.Canvas())
}
