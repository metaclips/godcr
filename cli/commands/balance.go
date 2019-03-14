package commands

import (
	"context"
	"fmt"

	"github.com/raedahgroup/godcr/app/walletcore"
	"github.com/raedahgroup/godcr/cli/termio"
)

// Balance displays the user's account balance.
type BalanceCommand struct {
	commanderStub
	Detailed bool `short:"d" long:"detailed" description:"Display detailed account balance report"`
}

// Run runs the `balance` command, displaying the user's account balance.
func (balanceCommand BalanceCommand) Run(ctx context.Context, wallet walletcore.Wallet) error {
	accounts, err := wallet.AccountsOverview(walletcore.DefaultRequiredConfirmations)
	if err != nil {
		return err
	}

	if balanceCommand.Detailed {
		showDetailedBalance(accounts)
	} else {
		showBalanceSummary(accounts)
	}

	return nil
}

func showDetailedBalance(accountBalances []*walletcore.Account) {
	columns := []string{
		"Account",
		"Total",
		"Spendable",
		"Locked By Tickets",
		"Voting Authority",
		"Unconfirmed",
	}
	rows := make([][]interface{}, len(accountBalances))
	for i, account := range accountBalances {
		rows[i] = []interface{}{
			account.Name,
			account.Balance.Total,
			account.Balance.Spendable,
			account.Balance.LockedByTickets,
			account.Balance.VotingAuthority,
			account.Balance.Unconfirmed,
		}
	}
	termio.PrintTabularResult(termio.StdoutWriter, columns, rows)
}

func showBalanceSummary(accounts []*walletcore.Account) {
	if len(accounts) == 1 {
		commandOutput := accounts[0].String()
		termio.PrintStringResult(commandOutput)
	} else {
		commandOutput := make([]string, len(accounts))
		for i, account := range accounts {
			commandOutput[i] = fmt.Sprintf("%s \t %s", account.Name, account.String())
		}
		termio.PrintStringResult(commandOutput...)
	}
}
