package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/raedahgroup/godcr/app/config"
	w "github.com/raedahgroup/godcr/app/wallet"
	"github.com/raedahgroup/godcr/app/wallet/libwallet"
	"github.com/raedahgroup/godcr/fyne"
	"github.com/raedahgroup/godcr/app"
)

func main() {
	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	initLogRotator(config.LogFile)
	defer func() {
		if logRotator != nil {
			logRotator.Close()
		}
	}()

	var appUI app.UserInterface

	// initialize appropriate ui here, fyne for now, could be any other interface in the future
	appUI = fyne.InitializeUserInterface()

	// nb: cli support will require loading from a config file
	cfg, err := config.LoadConfigFromDb()
	if err != nil {
		errorMessage := fmt.Sprintf("Error loading config from db: %v", err)
		log.Errorf(errorMessage)
		appUI.DisplayPreLaunchError(errorMessage)
		return
	}

	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(cfg.DebugLevel); err != nil {
		errorMessage := fmt.Sprintf("error setting log levels: %v", err)
		log.Errorf(errorMessage)
		appUI.DisplayPreLaunchError(errorMessage)
		return
	}

	// use wait group to keep main alive until shutdown completes
	shutdownWaitGroup := &sync.WaitGroup{}
	go handleShutdownRequests(shutdownWaitGroup)
	go listenForShutdownRequests()

	// open connection to wallet and add wallet shutdown function to shutdownOps
	wallet, err := connectToWallet(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to wallet.", err.Error())
		fmt.Println("Exiting.")
		os.Exit(1)
	}
	shutdownOps = append(shutdownOps, wallet.Shutdown)

	// use ctx to monitor potentially long running operations
	// such operations should listen for ctx.Done and stop further processing
	ctx, cancel := context.WithCancel(context.Background())
	shutdownOps = append(shutdownOps, cancel)

	appUI.LaunchApp(ctx, cfg, wallet)

	// wait for handleShutdownRequests goroutine, to finish before exiting main
	shutdownWaitGroup.Wait()
}

// connectToWallet opens connection to a wallet via dcrlibwallet (LibWallet)
// or dcrwalletrpc (RpcWallet, currently unimplemented.unsuported)
func connectToWallet(cfg *config.Config) (w.Wallet, error) {
	netType := "mainnet"
	if cfg.UseTestnet {
		netType = "testnet3"
	}
	return libwallet.Init(cfg.AppDataDir, netType)
}
