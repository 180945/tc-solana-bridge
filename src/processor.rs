use solana_program::{
    account_info::{next_account_info, AccountInfo},
    entrypoint::ProgramResult,
    msg,
    program::{invoke, invoke_signed},
    program_error::ProgramError,
    program_pack::{IsInitialized, Pack},
    pubkey::Pubkey,
    sysvar::{rent::Rent, Sysvar},
    instruction::{Instruction},
    secp256k1_recover::secp256k1_recover,
    keccak::hash,
    system_instruction,
    borsh0_10::try_from_slice_unchecked
};
use std::{
    str,
};
use borsh::{BorshSerialize};
use spl_token::state::Account as TokenAccount;
use crate::{error::BridgeError, instruction::BridgeInstruction, state::{WithdrawRequest, TcOwners, Nonces}};
use spl_associated_token_account::{get_associated_token_address};
use crate::state::SignData;

pub fn process_instruction(
        program_id: &Pubkey,
        accounts: &[AccountInfo],
        instruction_data: &[u8],
) -> ProgramResult {
    let instruction = BridgeInstruction::unpack(instruction_data)?;

    match instruction {
        BridgeInstruction::Deposit { amount, tc_address } => {
            msg!("Instruction: Deposit");
            process_deposit(accounts, amount, tc_address, program_id)
        }
        BridgeInstruction::Withdraw { withdraw_info } => {
            msg!("Instruction: Withdraw");
            process_withdraw(accounts, withdraw_info, program_id)
        }
        BridgeInstruction::InitOwners { init_beacon_info } => {
            msg!("Instruction: Init owner list");
            process_init_beacon(accounts, init_beacon_info, program_id)
        }
    }
}

fn process_deposit(
    accounts: &[AccountInfo],
    amount: u64,
    tc_address: [u8; 20],
    program_id: &Pubkey,
) -> ProgramResult {
    let account_info_iter = &mut accounts.iter();
    let depositor_token_account = next_account_info(account_info_iter)?;
    let vault_token_account = next_account_info(account_info_iter)?;
    let vault_program_owners_id = next_account_info(account_info_iter)?;
    let depositor = next_account_info(account_info_iter)?;
    if !depositor.is_signer {
        return Err(ProgramError::MissingRequiredSignature);
    }
    let token_program = next_account_info(account_info_iter)?;

    if vault_program_owners_id.owner != program_id {
        msg!("Invalid tc owners program");
        return Err(ProgramError::IncorrectProgramId);
    }

    if *vault_token_account.owner != spl_token::id() {
        msg!("vault token account must be owned by spl token");
        return Err(ProgramError::IncorrectProgramId);
    }

    let token_id = _verify_vault_token_account(
        vault_program_owners_id.clone(),
        vault_token_account.clone(),
        program_id.clone())?;

    spl_token_transfer(TokenTransferParams {
        source: depositor_token_account.clone(),
        destination: vault_token_account.clone(),
        amount,
        authority: depositor.clone(),
        authority_signer_seeds: &[],
        token_program: token_program.clone(),
    })?;
    msg!("tc_owners,address,token,amount:{},{},{},{}",
        vault_program_owners_id.key,str::from_utf8(&tc_address[..]).unwrap(), token_id, amount);

    Ok(())
}

// add logic to process withdraw for users
fn process_withdraw(
    accounts: &[AccountInfo],
    withdraw_info: WithdrawRequest,
    program_id: &Pubkey,
) -> ProgramResult {
    let account_info_iter = &mut accounts.iter();
    let vault_token_account = next_account_info(account_info_iter)?;
    let vault_authority_account = next_account_info(account_info_iter)?;
    let pda_account = next_account_info(account_info_iter)?;
    let tc_owners = next_account_info(account_info_iter)?;
    let token_program = next_account_info(account_info_iter)?;
    let authority_account = next_account_info(account_info_iter)?;

    if !authority_account.is_signer {
        return Err(BridgeError::InvalidAuthorityAccount.into())
    }

    let tc_owners_info = TcOwners::unpack_unchecked(&tc_owners.data.borrow())?;
    if !tc_owners_info.is_initialized() {
        return Err(BridgeError::BeaconsUnInitialized.into());
    }

    let request_withdraw_len = withdraw_info.amounts.len();
    if request_withdraw_len == 0 {
        return Err(BridgeError::EmptyWithdrawList.into());
    }

    let mut withdraw_token_accounts = Vec::with_capacity(request_withdraw_len + 1);
    let mut accounts_info = Vec::with_capacity(request_withdraw_len + 1);
    for _ in 0..request_withdraw_len {
        let account = next_account_info(account_info_iter)?;
        withdraw_token_accounts.push(account.key.clone().to_bytes());
        accounts_info.push(account);
    }

    if tc_owners.owner != program_id {
        msg!("Invalid tc owners");
        return Err(ProgramError::IncorrectProgramId);
    }

    // verify vault token account
    let token_id = _verify_vault_token_account(
        tc_owners.clone(),
        vault_token_account.clone(),
        program_id.clone())?;

    if withdraw_info.signatures.len() <= tc_owners_info.beacons.len() * 2 / 3 {
        msg!("Invalid signature input");
        return Err(BridgeError::InvalidNumberOfSignature.into());
    }

    // build sign data
    let current_nonce = _process_nonce(pda_account, program_id)?;
    let sign_data_struct = SignData{
        amounts: withdraw_info.amounts.clone(),
        accounts: withdraw_token_accounts,
        nonce: current_nonce,
    };
    let sign_data = hash(serde_json::to_string(&sign_data_struct).unwrap_or_default().as_bytes());

    for i in 0..withdraw_info.signatures.len() {
        let s_r_v = withdraw_info.signatures[i];
        let (s_r, v) = s_r_v.split_at(64);
        if v.len() != 1 {
            msg!("Invalid signature v input");
            return Err(BridgeError::InvalidBeaconSignature.into());
        }
        let beacon_key_from_signature_result = secp256k1_recover(
            &sign_data.to_bytes()[..],
            v[0],
            s_r,
        ).unwrap();
        let beacon_key = tc_owners_info.beacons[i];
        if beacon_key_from_signature_result != beacon_key {
            return Err(BridgeError::InvalidBeaconSignature.into());
        }
    }

    // prepare to transfer token to user
    let authority_signer_seeds = &[
        tc_owners.key.as_ref(),
        &[tc_owners_info.bump_seed],
    ];

    for i in 0..withdraw_info.amounts.len() {
        // transfer token
        spl_token_transfer(TokenTransferParams {
            source: vault_token_account.clone(),
            destination: accounts_info[i].clone(),
            amount: withdraw_info.amounts[i],
            authority: vault_authority_account.clone(),
            authority_signer_seeds,
            token_program: token_program.clone(),
        })?;

        if token_id == spl_token::native_mint::id() {
            // handle native token
            if *vault_token_account.key == *accounts_info[i].clone().key {
                msg!("Invalid sender and receiver in unshield request");
                return Err(BridgeError::InvalidTransferTokenData.into());
            }
            // close account
            spl_close_token_acc(TokenCloseParams {
                account: accounts_info[i].clone(),
                destination: next_account_info(account_info_iter)?.clone(),
                authority: vault_authority_account.clone(),
                authority_signer_seeds,
                token_program: token_program.clone(),
            })?;
        }
    }

    Ok(())
}

// add logic to proccess init beacon list
fn process_init_beacon(
    accounts: &[AccountInfo],
    init_beacon_info: TcOwners,
    program_id: &Pubkey,
) -> ProgramResult {
    let account_info_iter = &mut accounts.iter();
    let tc_owners = next_account_info(account_info_iter)?;
    let nonce = next_account_info(account_info_iter)?;
    let authority_account = next_account_info(account_info_iter)?;
    let system_program = next_account_info(account_info_iter)?;

    if !authority_account.is_signer {
        return Err(BridgeError::InvalidAuthorityAccount.into())
    }

    // create pda proxy program
    let (tc_owners_pda, bump) = Pubkey::find_program_address(
        &[
            [0].as_ref(),
        ],
        program_id
    );

    // check tc owner program must be empty
    if !tc_owners.data_is_empty() || *tc_owners.key != tc_owners_pda {
        return Err(BridgeError::PDAAccountCreated.into())
    }

    let rent = Rent::get()?;
    let rent_lamports = rent.minimum_balance(TcOwners::LEN);
    let create_map_ix = &system_instruction::create_account(
        authority_account.key,
        tc_owners.key,
        rent_lamports,
        TcOwners::LEN.try_into().unwrap(),
        program_id
    );
    invoke_signed(
        create_map_ix,
        &[
            authority_account.clone(),
            tc_owners.clone(),
            system_program.clone()
        ],
        &[&[
            [0].as_ref(),
            &[bump]
        ]]
    )?;
    let mut pda_owner_data = try_from_slice_unchecked::<TcOwners>(&tc_owners.try_borrow_data()?)?;
    pda_owner_data.is_initialized = true;
    pda_owner_data.bump_seed = bump;
    pda_owner_data.beacons = init_beacon_info.beacons;
    TcOwners::pack(pda_owner_data, &mut tc_owners.data.borrow_mut())?;

    // init nonce data storage
    let (nonce_pda, bump) = Pubkey::find_program_address(
        &[
            [1].as_ref(),
        ],
        program_id
    );

    // check tc owner program must be empty
    if !nonce.data_is_empty() || *nonce.key != nonce_pda {
        return Err(BridgeError::PDAAccountCreated.into())
    }

    let rent_lamports = rent.minimum_balance(Nonces::LEN);
    let create_map_ix = &system_instruction::create_account(
        authority_account.key,
        nonce.key,
        rent_lamports,
        Nonces::LEN.try_into().unwrap(),
        program_id
    );

    invoke_signed(
        create_map_ix,
        &[
            authority_account.clone(),
            nonce.clone(),
            system_program.clone()
        ],
        &[&[
            [1].as_ref(),
            &[bump]
        ]]
    )?;
    let mut nonce_data = try_from_slice_unchecked::<Nonces>(&nonce.try_borrow_data()?)?;
    nonce_data.is_initialized = true;
    nonce_data.nonce = 0;
    nonce_data.serialize(&mut &mut nonce.data.borrow_mut()[..])?;

    Ok(())
}

fn _process_nonce<'a>(
    pda_account: &AccountInfo<'a>,
    program_id: &Pubkey
) -> Result<u64, ProgramError> {
    // verify vault pda account
    let (pda, _) = Pubkey::find_program_address(
        &[
            [1].as_ref(),
        ],
        program_id
    );
    if pda != *pda_account.key {
        msg!("Mismatch pda generated and pda provided {}, {}", pda, pda_account.key);
        return Err(BridgeError::InvalidPDAAccount.into());
    }
    if !pda_account.is_writable {
        msg!("Pda account is not writable");
        return Err(BridgeError::InvalidPDAAccount.into());
    }

    let mut nonce_data = try_from_slice_unchecked::<Nonces>(&pda_account.try_borrow_data()?)?;
    let current_nonce = nonce_data.nonce;
    nonce_data.nonce = current_nonce + 1;
    nonce_data.serialize(&mut &mut pda_account.data.borrow_mut()[..])?;

    Ok(current_nonce)
}

fn _verify_vault_token_account(tc_owner_id: AccountInfo, vault_token_account: AccountInfo, program_id: Pubkey) -> Result<Pubkey, ProgramError> {
    let vault_token_account_info = TokenAccount::unpack(&vault_token_account.try_borrow_data()?)?;
    let tc_owners = TcOwners::unpack(&tc_owner_id.try_borrow_data()?)?;
    let authority_signer_seeds = &[
        tc_owner_id.key.as_ref(),
        &[tc_owners.bump_seed],
    ];
    let vault_authority_pubkey =
        Pubkey::create_program_address(authority_signer_seeds, &program_id)?;

    let tc_owners_associated_acc = get_associated_token_address(
        &vault_authority_pubkey,
        &vault_token_account_info.mint
    );

    if tc_owners_associated_acc != *vault_token_account.key {
        msg!("Only tc owners account will be accepted");
        return Err(ProgramError::IncorrectProgramId);
    }

    Ok(vault_token_account_info.mint)
}

/// Issue a spl_token `Transfer` instruction.
#[inline(always)]
fn spl_token_transfer(params: TokenTransferParams<'_, '_>) -> ProgramResult {
    let TokenTransferParams {
        source,
        destination,
        authority,
        token_program,
        amount,
        authority_signer_seeds,
    } = params;
    let result = invoke_optionally_signed(
        &spl_token::instruction::transfer(
            token_program.key,
            source.key,
            destination.key,
            authority.key,
            &[],
            amount,
        )?,
        &[source, destination, authority, token_program],
        authority_signer_seeds,
    );
    result.map_err(|_| BridgeError::TokenTransferFailed.into())
}

/// Issue a spl_token `Close` instruction.
#[inline(always)]
fn spl_close_token_acc(params: TokenCloseParams<'_, '_>) -> ProgramResult {
    let TokenCloseParams {
        account,
        destination,
        authority,
        token_program,
        authority_signer_seeds,
    } = params;
    let result = invoke_optionally_signed(
        &spl_token::instruction::close_account(
            token_program.key,
            account.key,
            destination.key,
            authority.key,
            &[],
        )?,
        &[account, destination, authority, token_program],
        authority_signer_seeds,
    );
    result.map_err(|_| BridgeError::CloseTokenAccountFailed.into())
}

/// Invoke signed unless signers seeds are empty
#[inline(always)]
fn invoke_optionally_signed(
    instruction: &Instruction,
    account_infos: &[AccountInfo],
    authority_signer_seeds: &[&[u8]],
) -> ProgramResult {
    if authority_signer_seeds.is_empty() {
        invoke(instruction, account_infos)
    } else {
        invoke_signed(instruction, account_infos, &[authority_signer_seeds])
    }
}

struct TokenTransferParams<'a: 'b, 'b> {
    source: AccountInfo<'a>,
    destination: AccountInfo<'a>,
    amount: u64,
    authority: AccountInfo<'a>,
    authority_signer_seeds: &'b [&'b [u8]],
    token_program: AccountInfo<'a>,
}

struct TokenCloseParams<'a: 'b, 'b> {
    account: AccountInfo<'a>,
    destination: AccountInfo<'a>,
    authority: AccountInfo<'a>,
    authority_signer_seeds: &'b [&'b [u8]],
    token_program: AccountInfo<'a>,
}