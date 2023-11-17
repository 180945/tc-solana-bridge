package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/180945/tc-contracts/services-go/deposit"
	"github.com/180945/tc-contracts/services-go/owners"
	"github.com/180945/tc-contracts/services-go/withdraw"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	_ "github.com/gagliardetto/solana-go/rpc/ws"
)

const DEPOSIT = "Deposit"
const WITHDRAW = "Withdraw"
const PROGRAM_ID = "8paNz9QEpU2PAPaZgQkm9cS6DLLkAvh3Eq6SHUq8JcEC"
const VAULT_ACC = "G65gJS4feG1KXpfDXiySUGT7c6QosCJcGa4nUZsF55Du"
const SYNC_NATIVE_TAG = 0x11
const NEW_TOKEN_ACC = 0x1
const ACCCOUN_SIZE = 165

func main() {
	// init vars
	//// Create a new WS client (used for confirming transactions)
	//wsClient, err := ws.Connect(context.Background(), rpc.DevNet_WS)
	//if err != nil {
	//	panic(err)
	//}

	program := solana.MustPublicKeyFromBase58(PROGRAM_ID)

	// gen tc owner program
	tcOwners, _, err := solana.FindProgramAddress(
		[][]byte{{0}},
		program,
	)

	// gen nonce program
	nonceProgram, _, err := solana.FindProgramAddress(
		[][]byte{{1}},
		program,
	)

	//vaultTokenAuthority, _, err := solana.FindProgramAddress(
	//	[][]byte{tcOwners.Bytes()},
	//	program,
	//)
	feePayer, err := solana.PrivateKeyFromBase58("36iz7qASXP1TECiAWJ3xESwYeRmTAF7ti3SAHaYuFJqHkmKenRACcJ9q48df1Bpn4QQSv3J89BGjuznaGAh9zPXg")
	if err != nil {
		panic(err)
	}
	//depositor, err := solana.PrivateKeyFromBase58("28BD5MCpihGHD3zUfPv4VcBizis9zxFdaj9fGJmiyyLmezT94FRd9XiiLjz5gasgyX3TmH1BU4scdVE6gzDFzrc7")
	//if err != nil {
	//	panic(err)
	//}

	// Create a new RPC client:
	rpcClient := rpc.New(rpc.DevNet_RPC)

	// test deposit tx
	recent, err := rpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}
	//tokenSell := solana.MustPublicKeyFromBase58("BEcGFQK1T1tSu3kvHC17cyCkQ5dvXqAJ7ExB2bb5Do7a")
	//tokenBuy := solana.MustPublicKeyFromBase58("FSRvxBNrQWX2Fy2qvKMLL3ryEdRtE3PUTZBcdKwASZTU")

	if false {
		fmt.Println("============ TRANSFER SOL =============")
		testPubKey, err := solana.WalletFromPrivateKeyBase58("28BD5MCpihGHD3zUfPv4VcBizis9zxFdaj9fGJmiyyLmezT94FRd9XiiLjz5gasgyX3TmH1BU4scdVE6gzDFzrc7")
		tx4, err := solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					2e9,
					feePayer.PublicKey(),
					testPubKey.PublicKey(),
				).Build(),
			},
			recent.Value.Blockhash,
			solana.TransactionPayer(feePayer.PublicKey()),
		)
		signers0 := []solana.PrivateKey{
			feePayer,
		}
		sig0, err := SignAndSendTx(tx4, signers0, rpcClient)
		if err != nil {
			panic(err)
		}
		spew.Dump(sig0)
	}

	// @Note: turn this off if initialized
	if false {
		fmt.Println("============ TEST INIT TC OWNER LIST =============")
		initOwnerInsts := []solana.Instruction{}
		initOwnerAccounts := []*solana.AccountMeta{
			solana.NewAccountMeta(tcOwners, true, false),
			solana.NewAccountMeta(nonceProgram, true, false),
			solana.NewAccountMeta(feePayer.PublicKey(), false, true),
			solana.NewAccountMeta(solana.SystemProgramID, false, false),
		}
		signers := []solana.PrivateKey{
			feePayer,
		}

		// list owner
		ownersPubKey := [][]byte{
			{64, 206, 253, 84, 56, 206, 63, 162, 157, 152, 148, 80, 198, 23, 66, 245, 43, 1, 207, 238, 9, 144, 161, 139, 131, 44, 146, 136, 74, 242, 22, 220, 187, 130, 145, 153, 93, 114, 117, 199, 108, 190, 233, 244, 53, 240, 247, 48, 207, 19, 94, 245, 14, 171, 207, 124, 157, 177, 173, 139, 253, 237, 36, 168},
			{175, 109, 126, 18, 52, 108, 137, 78, 38, 252, 216, 214, 224, 214, 44, 187, 2, 67, 70, 204, 196, 78, 155, 224, 72, 126, 124, 128, 134, 165, 210, 158, 138, 93, 62, 90, 76, 225, 186, 39, 215, 204, 170, 10, 127, 99, 86, 220, 107, 251, 34, 58, 235, 236, 69, 189, 235, 226, 57, 208, 106, 210, 28, 22},
			{122, 69, 179, 100, 37, 117, 17, 36, 0, 4, 211, 125, 150, 102, 106, 180, 218, 127, 238, 200, 104, 84, 250, 183, 23, 31, 209, 229, 22, 117, 248, 73, 56, 120, 112, 2, 188, 187, 152, 44, 70, 228, 25, 160, 250, 255, 40, 216, 180, 239, 183, 235, 175, 79, 66, 41, 119, 82, 195, 70, 103, 102, 135, 73},
			{24, 171, 11, 173, 118, 80, 213, 52, 20, 186, 77, 213, 182, 249, 188, 70, 15, 37, 228, 129, 102, 45, 183, 139, 139, 174, 147, 32, 130, 179, 168, 171, 36, 79, 30, 237, 44, 11, 200, 229, 108, 224, 117, 224, 206, 11, 62, 235, 127, 101, 194, 116, 209, 213, 122, 41, 77, 229, 19, 60, 199, 168, 81, 25},
		}

		// build tx
		initOwner := owners.NewOwnerInit(ownersPubKey, program, initOwnerAccounts)
		initOwnerInsts = append(initOwnerInsts, initOwner.Build())

		// create tx
		tx3, err := solana.NewTransaction(
			initOwnerInsts,
			recent.Value.Blockhash,
			solana.TransactionPayer(feePayer.PublicKey()),
		)
		if err != nil {
			panic(err)
		}
		sig, err := SignAndSendTx(tx3, signers, rpcClient)
		if err != nil {
			panic(err)
		}
		spew.Dump(sig)
	}

	vaultTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{tcOwners.Bytes()},
		program,
	)

	if false {
		fmt.Println("============ TEST DEPOSIT TOKEN TO BRIDGE =============")
		depositInsts := []solana.Instruction{}
		solAmount := uint64(1e8) // 0.1 sol
		tcAddress := "0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5"
		signers := []solana.PrivateKey{
			feePayer,
		}

		// inst create new vault sol token account
		depositNativeTokenAcc, _, _ := solana.FindAssociatedTokenAddress(
			feePayer.PublicKey(),
			solana.SolMint,
		)

		fmt.Println(depositNativeTokenAcc.String())
		fmt.Println(feePayer.PublicKey().String())

		// append inst to create account
		_, err = rpcClient.GetAccountInfo(context.TODO(), depositNativeTokenAcc)
		if err != nil {
			depositInsts = append(depositInsts, associatedtokenaccount.NewCreateInstruction(
				feePayer.PublicKey(),
				feePayer.PublicKey(),
				solana.SolMint,
			).Build())
		}

		vaultNativeTokenAcc, _, _ := solana.FindAssociatedTokenAddress(
			vaultTokenAuthority,
			solana.SolMint,
		)
		_, err = rpcClient.GetAccountInfo(context.TODO(), vaultNativeTokenAcc)
		if err != nil {
			depositInsts = append(depositInsts, associatedtokenaccount.NewCreateInstruction(
				feePayer.PublicKey(),
				vaultTokenAuthority,
				solana.SolMint,
			).Build())
		}

		fmt.Println(vaultNativeTokenAcc.String())

		// create inst to convert sol -> wsol
		depositInsts = append(depositInsts, system.NewTransferInstruction(
			solAmount,
			feePayer.PublicKey(),
			depositNativeTokenAcc,
		).Build())

		// build sync native token program
		depositInsts = append(depositInsts, solana.NewInstruction(
			solana.TokenProgramID,
			[]*solana.AccountMeta{
				solana.NewAccountMeta(depositNativeTokenAcc, true, false),
			},
			[]byte{SYNC_NATIVE_TAG},
		))

		// inst deposit
		depositAccs := []*solana.AccountMeta{
			solana.NewAccountMeta(depositNativeTokenAcc, true, false),
			solana.NewAccountMeta(vaultNativeTokenAcc, true, false),
			solana.NewAccountMeta(tcOwners, false, false),
			solana.NewAccountMeta(feePayer.PublicKey(), false, true),
			solana.NewAccountMeta(solana.TokenProgramID, false, false),
		}
		depositInst := deposit.NewDeposit(tcAddress, solAmount, program, depositAccs)
		depositInsts = append(depositInsts, depositInst.Build())

		tx3, err := solana.NewTransaction(
			depositInsts,
			recent.Value.Blockhash,
			solana.TransactionPayer(feePayer.PublicKey()),
		)
		if err != nil {
			panic(err)
		}
		sig, err := SignAndSendTx(tx3, signers, rpcClient)
		if err != nil {
			panic(err)
		}
		spew.Dump(sig)
	}

	if true {
		fmt.Println("============ TEST WITHDRAW TOKEN TO BRIDGE =============")
		withdrawInsts := []solana.Instruction{}
		signers := []solana.PrivateKey{
			feePayer,
		}
		ownerKyes := []string{
			"aad53b70ad9ed01b75238533dd6b395f4d300427da0165aafbd42ea7a606601f",
			"ca71365ceddfa8e0813cf184463bd48f0b62c9d7d5825cf95263847628816e82",
			"1e4d2244506211200640567630e3951abadbc2154cf772e4f0d2ff0770290c7c",
			"c7146b500240ed7aac9445e2532ae8bf6fc7108f6ea89fde5eebdf2fb6cefa5a",
		}
		// withdraw amounts
		amounts := []uint64{1e7, 1e7}
		// withdraw addresses
		withraw1, _ := solana.PrivateKeyFromBase58("5z6TLNJC9gyn3WBhQTtbr5gyKAKjMqHStgzHAGEXeRkW2ZLNMKLWQBWovpQRHPJsFr6s3jxB8r7nqLgoe7rTAdgj")
		withraw2, _ := solana.PrivateKeyFromBase58("3q6BFcYPFbCvtxvsk6xQoXqQjGJNm3afjBw3qMvRSkCehVbEPEJUxmc1ZzTyCSzWc4a6miaZ4jdL5wrdwhLoKN1n")
		withdrawToKeys := []solana.PrivateKey{
			withraw1, withraw2,
		}

		withdrawAddresses := [][32]byte{
			withdraw.ToByte32(withraw1.PublicKey().Bytes()), withdraw.ToByte32(withraw2.PublicKey().Bytes()),
		}

		vaultNativeTokenAcc, _, _ := solana.FindAssociatedTokenAddress(
			vaultTokenAuthority,
			solana.SolMint,
		)

		// withdraw account
		withdrawAccounts := []*solana.AccountMeta{
			solana.NewAccountMeta(vaultNativeTokenAcc, true, false),
			solana.NewAccountMeta(vaultTokenAuthority, false, false),
			solana.NewAccountMeta(nonceProgram, true, false),
			solana.NewAccountMeta(tcOwners, false, false),
			solana.NewAccountMeta(solana.TokenProgramID, false, false),
			solana.NewAccountMeta(feePayer.PublicKey(), true, false),
		}

		exemptLamport, err := rpcClient.GetMinimumBalanceForRentExemption(context.Background(), ACCCOUN_SIZE, rpc.CommitmentConfirmed)
		if err != nil {
			panic(err)
		}

		// append
		for _, v := range withdrawToKeys {
			withdrawAccounts = append(withdrawAccounts, solana.NewAccountMeta(v.PublicKey(), true, false))
		}

		// append instructions
		for _, _ = range amounts {
			tempAccount, _ := solana.NewRandomPrivateKey()
			if err != nil {
				panic(err)
			}

			// init new account
			newAccount := system.NewCreateAccountInstruction(
				exemptLamport,
				ACCCOUN_SIZE,
				solana.TokenProgramID,
				feePayer.PublicKey(),
				tempAccount.PublicKey(),
			).Build()

			// build create new token acc
			newAccTokenInst := solana.NewInstruction(
				solana.TokenProgramID,
				[]*solana.AccountMeta{
					solana.NewAccountMeta(tempAccount.PublicKey(), true, false),
					solana.NewAccountMeta(solana.SolMint, false, false),
					solana.NewAccountMeta(vaultTokenAuthority, false, false),
					solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
				},
				[]byte{NEW_TOKEN_ACC},
			)

			withdrawInsts = append(withdrawInsts, newAccount)
			withdrawInsts = append(withdrawInsts, newAccTokenInst)
			signers = append(signers, tempAccount)

			withdrawAccounts = append(withdrawAccounts, solana.NewAccountMeta(tempAccount.PublicKey(), true, false))
		}

		// get nonce from program
		nonceProgramInfo, err := rpcClient.GetAccountInfo(context.TODO(), nonceProgram)
		if err != nil {
			panic(err)
		}
		// 1 byte init + 8 bytes nonce
		nonce := binary.LittleEndian.Uint64(nonceProgramInfo.Bytes()[1:])
		fmt.Printf("nonce: %v \n", nonce)

		withdrawInst := withdraw.NewWithdraw(
			ownerKyes,
			amounts,
			withdrawAddresses,
			nonce,
			program,
			withdrawAccounts,
		)

		withdrawInsts = append(withdrawInsts, withdrawInst.Build())

		tx5, err := solana.NewTransaction(
			withdrawInsts,
			recent.Value.Blockhash,
			solana.TransactionPayer(feePayer.PublicKey()),
		)
		sig, err := SignAndSendTx(tx5, signers, rpcClient)
		if err != nil {
			panic(err)
		}
		spew.Dump(sig)

	}
}

func SignAndSendTx(tx *solana.Transaction, signers []solana.PrivateKey, rpcClient *rpc.Client) (solana.Signature, error) {
	_, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, candidate := range signers {
			if candidate.PublicKey().Equals(key) {
				return &candidate
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("unable to sign transaction: %v \n", err)
		return solana.Signature{}, err
	}
	// send tx
	signature, err := rpcClient.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Printf("unable to send transaction: %v \n", err)
		return solana.Signature{}, err
	}
	return signature, nil
}
