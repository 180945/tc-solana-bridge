package deposit

import (
	"encoding/binary"
	"fmt"
	"github.com/gagliardetto/solana-go"
)

const DEPOSIT_TAG = 0x00

type Deposit struct {
	tcAddress string
	amount    uint64
	programID solana.PublicKey
	accounts  []*solana.AccountMeta
}

func NewDeposit(tcAddress string, amount uint64, programID solana.PublicKey, accounts []*solana.AccountMeta) Deposit {
	return Deposit{
		tcAddress: tcAddress,
		amount:    amount,
		programID: programID,
		accounts:  accounts,
	}
}

func (sh *Deposit) Build() *solana.GenericInstruction {
	if len(sh.tcAddress) != 42 {
		fmt.Printf("invalid tc address %v \n", sh.tcAddress)
		return nil
	}
	if len(sh.accounts) == 0 {
		fmt.Printf("invalid account list %v \n", sh.accounts)
		return nil
	}

	// build instruction
	// convert string to byte array
	tcAddressBytes := []byte(sh.tcAddress)
	// convert amount to bid le endian representation 8 bytes
	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBytes, sh.amount)
	temp1 := append([]byte{DEPOSIT_TAG}, amountBytes...)
	depositData := append(temp1, tcAddressBytes...)
	accountSlice := solana.AccountMetaSlice{}
	err := accountSlice.SetAccounts(
		sh.accounts,
	)
	if err != nil {
		fmt.Printf("init account slice failed %v \n", err)
		return nil
	}

	return solana.NewInstruction(
		sh.programID,
		accountSlice,
		depositData,
	)
}
