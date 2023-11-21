package main

import (
	"context"
	"encoding/binary"
	"encoding/hex"
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
	rpcClient := rpc.New(rpc.MainNetBeta_RPC)

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
	if true {
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

		ownersStr := []string{
			"6aa595813d5d7866f8a891630b4123e718715669acb0fd6fe2f7c46dfb8a322f6dcfc1880910404760bbfa373e66cc26355fb79699316a061c4b018a6852d1ea",
			"7bb0fb6727892bde8acefb3b7e2d77d7014169385e38d2ce04694603d0432a94d518ef3def8ca1eb8e8a4cbc4673f7c809ea8a3794a855e7846febcdd5747288",
			"7a5ad8fcaafcc7defc4fdf3a82f487016ee8665f918d10540b9faf4196d02561933ea8856919e43c335b08862a059535caaa781b2902296eada8f5114fdad778",
			"59faeba475d6eea29653f50949a4b33d65fd063b3259a5d4ea9999e715bf8fb564a9ffb1656ca8638622e114dfbb2935dc7cfd1f35a477d979ee04579617ab36",
			"b2270335e1ef2a2be3565f673ec35eb5ae7c41ad132a4d30b96e870f7aaa45b32f428cf950cdafb588774f30a91f7459395df5e4f321ef6aa3c4e8a0d43fdd48",
			"f8651fb2be5886101f78a65f1d8cd838b740560b2bf40c5d794d9fca13f576493e15e79f96a8ebb6969c6a00386b7f6486f7ed3ac04ea888f9b9009f634d99b8",
			"6b5f0bc200b94dd826afba0fde6a892058def9ce5df72e091d35fbaafd37b5dcfb7a370bcee545665153d54650ee6dd2862c0ab0e1fe394648bb0fab195be57f",
		}

		// list owner
		ownersPubKey := [][]byte{}
		for i := 0; i < len(ownersStr); i++ {
			owner, _ := hex.DecodeString(ownersStr[i])
			ownersPubKey = append(ownersPubKey, owner)
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

	if false {
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
