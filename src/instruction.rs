use solana_program::{
    program_error::ProgramError,
    msg,
    secp256k1_recover::{Secp256k1Pubkey},
};
use crate::error::BridgeError::{
    InvalidInstruction,
    InstructionUnpackError
};
use crate::state::{TcOwners, WithdrawRequest, OwnerInit};
use std::{convert::TryInto};

pub enum BridgeInstruction {

    ///   Request new deposit to move token from Solana -> TC.
    ///
    ///   0. `[writable]` Token account to make deposit request
    ///   1. `[writable]` Vault token account to receive token from asker
    ///   2. `[]` Incognito proxy which stores beacon list and bump seed to retrieve vault token account
    ///   3. `[signer]` Deposit maker address
    ///   4. `[]` Spl Token program id
    Deposit {
        /// deposit info
        amount: u64,
        tc_address: [u8; 20],
    },

    ///   Request new withdraw to move token from TC -> Solana.
    ///
    ///   0. `[writable]` Vault token account to transfer tokens to withdraw maker
    ///   1. `[]` Withdraw maker address
    ///   2. `[]` $vault_authority derived from `create_program_address(&[incognito proxy account])`
    ///   3. `[writable]` pda account to mark burn id already used
    ///   4. `[]` Incognito proxy which stores beacon list and bump seed to retrieve vault token account
    ///   5. `[]` Spl Token program id
    ///   6. `[writable]` Associated token account of withdraw maker
    ///   7. `[signer]` Fee payer for pda account which stores tx id
    ///   8. `[]` System program to create pda account
    Withdraw {
        /// withdraw info
        withdraw_info: WithdrawRequest,
    },

    ///   Initializes a new Incognito proxy account.
    ///
    ///   0. `[]` $SYSVAR_RENT_PUBKEY to check account rent exempt
    ///   1. `[writable]` Incognito proxy account
    ///   2. `[writable]` Vault account
    InitOwners {
        /// beacon info
        init_beacon_info: OwnerInit,
    }
}

impl BridgeInstruction {
    /// Unpacks a byte buffer into a [BridgeInstruction](enum.BridgeInstruction.html).
    pub fn unpack(input: &[u8]) -> Result<Self, ProgramError> {
        let (tag, rest) = input.split_first().ok_or(InvalidInstruction)?;
        Ok(match tag {
            0 => {
                let (amount, rest) = Self::unpack_u64(rest)?;
                let (tc_address, _) = Self::unpack_bytes20(rest)?;
                Self::Deposit {
                    amount,
                    tc_address: tc_address.clone()
                }
            },
            1 => {
                // unpack amounts
                let (amount_len,mut rest) = Self::unpack_u8(rest)?;
                let mut amounts = Vec::with_capacity(amount_len as usize + 1);
                for _ in 0..amount_len {
                    let (amount, rest_) = Self::unpack_u64(rest)?;
                    rest = rest_;
                    amounts.push(amount);
                }

                // signatures
                let (signature_len,mut rest) = Self::unpack_u8(rest)?;
                let mut signatures = Vec::with_capacity(signature_len as usize + 1);
                for _ in 0..signature_len {
                    let (signature, rest_) = Self::unpack_bytes65(rest)?;
                    rest = rest_;
                    signatures.push(*signature);
                }

                let withdraw_info = WithdrawRequest {
                    amounts,
                    signatures,
                };

                Self::Withdraw {
                    withdraw_info
                }
            },
            2 => {
                let (beacon_list_len, mut rest) =  Self::unpack_u8(rest)?;
                let mut beacons = Vec::with_capacity(beacon_list_len as usize + 1);
                for _ in 0..beacon_list_len {
                    let (beacon, rest_) = Self::unpack_bytes64(rest)?;
                    rest = rest_;
                    let new_beacon = Secp256k1Pubkey::new(beacon);
                    beacons.push(new_beacon);
                }
                Self::InitOwners {
                    init_beacon_info: OwnerInit{
                        beacons
                    }   
                }
            }
            _ => return Err(InvalidInstruction.into()),
        })
    }

    fn unpack_u64(input: &[u8]) -> Result<(u64, &[u8]), ProgramError> {
        if input.len() < 8 {
            msg!("u64 cannot be unpacked");
            return Err(InstructionUnpackError.into());
        }
        let (bytes, rest) = input.split_at(8);
        let value = bytes
            .get(..8)
            .and_then(|slice| slice.try_into().ok())
            .map(u64::from_le_bytes)
            .ok_or(InstructionUnpackError)?;
        Ok((value, rest))
    }

    fn unpack_bytes20(input: &[u8]) -> Result<(&[u8; 20], &[u8]), ProgramError> {
        if input.len() < 20 {
            msg!("20 bytes cannot be unpacked");
            return Err(InstructionUnpackError.into());
        }
        let (bytes, rest) = input.split_at(20);
        Ok((
            bytes
                .try_into()
                .map_err(|_| InstructionUnpackError)?,
            rest,
        ))
    }

    fn unpack_u8(input: &[u8]) -> Result<(u8, &[u8]), ProgramError> {
        if input.is_empty() {
            msg!("u8 cannot be unpacked");
            return Err(InstructionUnpackError.into());
        }
        let (bytes, rest) = input.split_at(1);
        let value = bytes
            .get(..1)
            .and_then(|slice| slice.try_into().ok())
            .map(u8::from_le_bytes)
            .ok_or(InstructionUnpackError)?;
        Ok((value, rest))
    }

    fn unpack_bytes65(input: &[u8]) -> Result<(&[u8; 65], &[u8]), ProgramError> {
        if input.len() < 65 {
            msg!("65 bytes cannot be unpacked");
            return Err(InstructionUnpackError.into());
        }
        let (bytes, rest) = input.split_at(65);
        Ok((
            bytes
                .try_into()
                .map_err(|_| InstructionUnpackError)?,
            rest,
        ))
    }

    fn unpack_bytes64(input: &[u8]) -> Result<(&[u8; 64], &[u8]), ProgramError> {
        if input.len() < 64 {
            msg!("64 bytes cannot be unpacked {:?}", input);
            return Err(InstructionUnpackError.into());
        }
        let (bytes, rest) = input.split_at(64);
        Ok((
            bytes
                .try_into()
                .map_err(|_| InstructionUnpackError)?,
            rest,
        ))
    }
}
