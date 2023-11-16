package withdraw

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
)

const WITHDRAW_TAG = 0x1

type Withdraw struct {
	ownerKeys       []string
	withdrawAmounts []uint64
	addresses       [][32]byte
	programID       solana.PublicKey
	accounts        []*solana.AccountMeta
	nonce           uint64
}

func NewWithdraw(ownerPrivKey []string, withdrawAmounts []uint64, addresses [][32]byte, nonce uint64, programID solana.PublicKey, accounts []*solana.AccountMeta) Withdraw {
	return Withdraw{
		ownerKeys:       ownerPrivKey,
		withdrawAmounts: withdrawAmounts,
		programID:       programID,
		accounts:        accounts,
		addresses:       addresses,
		nonce:           nonce,
	}
}

type SignData struct {
	Amounts  []uint64   `json:"amounts"`
	Accounts [][32]byte `json:"accounts"`
	Nonce    uint64     `json:"nonce"`
}

func (us *Withdraw) Build() *solana.GenericInstruction {
	if len(us.addresses) != len(us.withdrawAmounts) {
		fmt.Printf("length of addresses and amounts not match")
		return nil
	}

	// build Withdraw instruction
	tag := WITHDRAW_TAG
	withdrawData := append([]byte{byte(tag)}, byte(len(us.withdrawAmounts)))
	for _, v := range us.withdrawAmounts {
		amountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(amountBytes, v)
		withdrawData = append(withdrawData, amountBytes...)
	}

	// build sign data
	// example
	// "{\"amounts\":[1,2],\"accounts\":[[130,130,145,153,93,114,117,199,108,190,233,244,53,240,247,48,207,19,94,245,14,171,207,124,157,177,173,139,253,237,36,168],[123,130,145,153,93,114,117,199,108,190,233,244,53,240,247,48,207,19,94,245,14,171,207,124,157,177,173,139,253,237,36,168]],\"nonce\":1}"
	signDataJson, _ := json.Marshal(
		SignData{
			Amounts:  us.withdrawAmounts,
			Accounts: us.addresses,
			Nonce:    us.nonce,
		})
	fmt.Println(signDataJson)
	signHash := keccak256(signDataJson)

	// append signature
	withdrawData = append(withdrawData, byte(len(us.ownerKeys)))
	for _, v := range us.ownerKeys {
		pr, _ := hex.DecodeString(v)
		sk, err := crypto.ToECDSA(pr)
		if err != nil {
			return nil
		}

		signature, err := crypto.Sign(signHash[:], sk)
		if err != nil {
			return nil
		}
		withdrawData = append(withdrawData, signature...)
	}

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
