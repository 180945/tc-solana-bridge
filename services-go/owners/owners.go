package owners

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
)

const INIT_OWNER_TAG = 0x02

type OwnerInit struct {
	owners    [][]byte // list owner public keys
	programID solana.PublicKey
	accounts  []*solana.AccountMeta
}

func NewOwnerInit(owners [][]byte, programID solana.PublicKey, accounts []*solana.AccountMeta) OwnerInit {
	return OwnerInit{
		owners:    owners,
		programID: programID,
		accounts:  accounts,
	}
}

func (sh *OwnerInit) Build() *solana.GenericInstruction {
	if len(sh.owners) == 0 {
		fmt.Printf("invalid ownerlist %v \n", sh.owners)
		return nil
	}

	if len(sh.accounts) == 0 {
		fmt.Printf("invalid account list %v \n", sh.accounts)
		return nil
	}

	// build instruction
	ownerInitData := append([]byte{byte(INIT_OWNER_TAG)}, byte(len(sh.owners)))

	// instruction owner list
	for _, v := range sh.owners {
		ownerInitData = append(ownerInitData, v[:]...)
	}

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
		ownerInitData,
	)
}
