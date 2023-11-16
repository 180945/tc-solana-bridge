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
	fmt.Println(string(signDataJson))
	fmt.Println(string([]byte{123, 34, 97, 109, 111, 117, 110, 116, 115, 34, 58, 91, 49, 48, 48, 48, 48, 48, 48, 48, 44, 49, 48, 48, 48, 48, 48, 48, 48, 93, 44, 34, 97, 99, 99, 111, 117, 110, 116, 115, 34, 58, 91, 91, 54, 48, 44, 49, 56, 52, 44, 49, 57, 50, 44, 49, 56, 44, 49, 51, 44, 49, 53, 44, 50, 51, 56, 44, 49, 50, 48, 44, 50, 51, 50, 44, 51, 53, 44, 50, 49, 53, 44, 49, 48, 54, 44, 49, 44, 49, 54, 51, 44, 49, 49, 48, 44, 50, 48, 56, 44, 50, 53, 51, 44, 49, 48, 51, 44, 50, 48, 56, 44, 50, 51, 49, 44, 49, 50, 57, 44, 56, 53, 44, 49, 57, 50, 44, 50, 50, 49, 44, 50, 54, 44, 49, 50, 50, 44, 50, 49, 52, 44, 50, 52, 57, 44, 54, 49, 44, 49, 53, 49, 44, 49, 48, 52, 44, 50, 49, 53, 93, 44, 91, 49, 52, 52, 44, 49, 56, 44, 48, 44, 49, 50, 56, 44, 49, 55, 50, 44, 52, 54, 44, 49, 49, 50, 44, 50, 52, 50, 44, 49, 54, 51, 44, 57, 52, 44, 49, 49, 57, 44, 49, 50, 50, 44, 50, 48, 53, 44, 52, 50, 44, 50, 50, 57, 44, 49, 48, 49, 44, 49, 48, 50, 44, 49, 56, 53, 44, 49, 50, 53, 44, 49, 55, 44, 49, 56, 55, 44, 55, 49, 44, 56, 51, 44, 50, 49, 56, 44, 49, 56, 44, 49, 50, 49, 44, 49, 56, 57, 44, 50, 51, 44, 50, 53, 52, 44, 51, 48, 44, 50, 49, 54, 44, 50, 49, 54, 93, 93, 44, 34, 110, 111, 110, 99, 101, 34, 58, 48, 125}))
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
