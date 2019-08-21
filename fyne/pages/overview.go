package pages

import (
	"image/color"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/decred/dcrd/dcrutil"
	godcrApp "github.com/raedahgroup/godcr/app"
	"github.com/raedahgroup/godcr/app/walletcore"
	"github.com/raedahgroup/godcr/fyne/widgets"
)

// overviewPageData contains widgets that needs to be updated realtime
type overviewPageData struct {
	balance         *widget.Label
	noActivityLabel *widget.Label
	txTable         widgets.TableStruct
	icon            *canvas.Image
	goDcrLabel      *canvas.Text
}

var overview overviewPageData

func overviewPageUpdates(wallet godcrApp.WalletMiddleware) {
	overview.balance.SetText(fetchBalance(wallet))
	var txTable widgets.TableStruct
	fetchTxTable(false, &txTable, 0, 5, wallet, nil)
	overview.txTable.Result.Children = txTable.Result.Children
	widget.Refresh(overview.txTable.Result)
}

// statusUpdates updates peerconn and blkheight
func statusUpdates(wallet godcrApp.WalletMiddleware) {
	info, _ := wallet.WalletConnectionInfo()

	if info.PeersConnected <= 1 {
		menu.peerConn.SetText(strconv.Itoa(int(info.PeersConnected)) + " Peer Connected")
	} else {
		menu.peerConn.SetText(strconv.Itoa(int(info.PeersConnected)) + " Peers Connected")
	}

	menu.blkHeight.SetText(strconv.Itoa(int(info.LatestBlock)) + " Blocks Connected")
}

func overviewPage(wallet godcrApp.WalletMiddleware, fyneApp fyne.App) fyne.CanvasObject {
	fyneTheme := fyneApp.Settings().Theme()
	if fyneTheme.BackgroundColor() == theme.LightTheme().BackgroundColor() {
		decredDark, err := ioutil.ReadFile("./fyne/pages/png/decredDark.png")
		if err != nil {
			log.Fatalln("decred dark png file missing", err)
		}
		overview.goDcrLabel = canvas.NewText(godcrApp.DisplayName, color.RGBA{0, 0, 255, 0})
		overview.icon = canvas.NewImageFromResource(fyne.NewStaticResource("Decred", decredDark))

	} else if fyneTheme.BackgroundColor() == theme.DarkTheme().BackgroundColor() {
		decredLight, err := ioutil.ReadFile("./fyne/pages/png/decredLight.png")
		if err != nil {
			log.Fatalln("decred light file missing", err)
		}
		overview.goDcrLabel = canvas.NewText(godcrApp.DisplayName, color.RGBA{255, 255, 255, 0})
		overview.icon = canvas.NewImageFromResource(fyne.NewStaticResource("Decred", decredLight))
	}
	overview.icon.FillMode = canvas.ImageFillOriginal
	iconEnlarge := fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(66, 55)), overview.icon)

	overview.goDcrLabel.TextSize = 20
	overview.goDcrLabel.TextStyle = fyne.TextStyle{Bold: true}
	canvas.Refresh(overview.goDcrLabel)

	iconLabel := fyne.NewContainerWithLayout(layout.NewBorderLayout(iconEnlarge, overview.goDcrLabel, nil, nil), iconEnlarge, overview.goDcrLabel)
	overview.goDcrLabel.Move(fyne.NewPos(15, 40))

	balanceLabel := widget.NewLabelWithStyle("Current Total Balance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	activityLabel := widget.NewLabelWithStyle("Recent Activity", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	overview.balance = widget.NewLabel(fetchBalance(wallet))
	overview.noActivityLabel = widget.NewLabelWithStyle("No activities yet", fyne.TextAlignCenter, fyne.TextStyle{})

	fetchTxTable(false, &overview.txTable, 0, 5, wallet, nil)
	output := widget.NewVBox(
		widgets.NewVSpacer(20),
		iconLabel,
		widgets.NewVSpacer(20),
		balanceLabel,
		overview.balance,
		widgets.NewVSpacer(5),
		activityLabel,
		overview.noActivityLabel,
		fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(overview.txTable.Result.MinSize().Width, overview.txTable.Result.MinSize().Height)), overview.txTable.Container))

	return widget.NewHBox(widgets.NewHSpacer(20), output)
}

func fetchBalance(wallet godcrApp.WalletMiddleware) string {
	accounts, err := wallet.AccountsOverview(walletcore.DefaultRequiredConfirmations)
	if err != nil {
		return err.Error()
	}

	return walletcore.WalletBalance(accounts)
}

func fetchTxTable(isHistoryPage bool, txTable *widgets.TableStruct, offset, counter int32, wallet godcrApp.WalletMiddleware, window fyne.Window) {
	var txs []*walletcore.Transaction
	if !isHistoryPage {
		txs, _ = wallet.TransactionHistory(offset, counter, nil)
	} else {
		splittedWord := strings.Split(history.txFilters.Selected, " ")
		txs, _ = wallet.TransactionHistory(offset, counter, walletcore.BuildTransactionFilter(splittedWord[0]))
	}
	if len(txs) > 0 {
		overview.noActivityLabel.Hide()
	}

	heading := widget.NewHBox(
		widget.NewLabelWithStyle("Date (UTC)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Type", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Direction", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Amount", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Fee", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Status", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Hash", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	var hBox []*widget.Box
	for _, tx := range txs {
		trimmedHash := tx.Hash[:15] + "..." + tx.Hash[len(tx.Hash)-15:]
		var hash fyne.CanvasObject
		if isHistoryPage {
			hash = widget.NewButton(trimmedHash, func() {
				getTxDetails(tx.Hash, wallet, window)
			})
		} else {
			hash = widget.NewLabelWithStyle(trimmedHash, fyne.TextAlignCenter, fyne.TextStyle{})
		}

		hBox = append(hBox, widget.NewHBox(
			widget.NewLabelWithStyle(tx.LongTime, fyne.TextAlignCenter, fyne.TextStyle{}),
			widget.NewLabelWithStyle(tx.Type, fyne.TextAlignCenter, fyne.TextStyle{}),
			widget.NewLabelWithStyle(tx.Direction.String(), fyne.TextAlignLeading, fyne.TextStyle{}),
			widget.NewLabelWithStyle(dcrutil.Amount(tx.Amount).String(), fyne.TextAlignTrailing, fyne.TextStyle{}),
			widget.NewLabelWithStyle(dcrutil.Amount(tx.Fee).String(), fyne.TextAlignCenter, fyne.TextStyle{}),
			widget.NewLabelWithStyle(tx.Status, fyne.TextAlignCenter, fyne.TextStyle{}),
			hash,
		))
	}
	txTable.NewTable(heading, hBox...)
}
