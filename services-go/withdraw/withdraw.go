package withdraw

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
)

const WITHDRAW_TAG = 0x1

type Withdraw struct {
	ownerKeys       []string
	withdrawAmounts []uint64
	programID       solana.PublicKey
	accounts        []*solana.AccountMeta
}

func NewWithdraw(ownerPrivKey []string, withdrawAmounts []uint64, programID solana.PublicKey, accounts []*solana.AccountMeta) Withdraw {
	return Withdraw{
		ownerKeys:       ownerPrivKey,
		withdrawAmounts: withdrawAmounts,
		programID:       programID,
		accounts:        accounts,
	}
}

func (us *Withdraw) Build() *solana.GenericInstruction {

	// build Withdraw instruction
	tag := WITHDRAW_TAG

	// todo update
	withdrawData := append([]byte{byte(tag)}, byte(1))

	accountSlice := solana.AccountMetaSlice{}
	err := accountSlice.SetAccounts(
		us.accounts,
	)
	if err != nil {
		fmt.Printf("init account slice failed %v \n", err)
		return nil
	}

	return solana.NewInstruction(
		us.programID,
		accountSlice,
		withdrawData,
	)
}
