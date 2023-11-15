use solana_program::{
    program_error::ProgramError,
    program_pack::{IsInitialized, Pack, Sealed},
    secp256k1_recover::{Secp256k1Pubkey, SECP256K1_PUBLIC_KEY_LENGTH},
};
use arrayref::{array_mut_ref, array_ref, array_refs, mut_array_refs};
use borsh::{BorshSerialize, BorshDeserialize};
use crate::error::BridgeError;
use serde::{Serialize, Deserialize};


/// ====== INCOGNITO PDA BURNID =======
#[derive(Clone, Default, BorshSerialize, BorshDeserialize)]
pub struct Nonces {
    pub is_initialized: bool,
    pub nonce: u64,
}

impl Nonces {
    pub const LEN: usize = 9; // 1 bool + 8 bytes uint64 for nonce
}

impl IsInitialized for Nonces {
    fn is_initialized(&self) -> bool {
        self.is_initialized
    }
}

/// ====== INCOGNITO PROXY =======
/// 
/// Max number of beacon addresses
pub const MAX_BEACON_ADDRESSES: usize = 20;

// Incognito proxy stores beacon list
#[derive(Clone, Default, PartialEq, BorshSerialize, BorshDeserialize)]
pub struct TcOwners {
    // init beacon 
    pub is_initialized: bool,
    // bump seed
    pub bump_seed: u8,
    /// beacon list
    pub beacons: Vec<Secp256k1Pubkey>, 
}

impl IsInitialized for TcOwners {
    fn is_initialized(&self) -> bool {
        self.is_initialized
    }
}

impl TcOwners {
    /// Create a new lending market
    pub fn new(params: TcOwners) -> Self {
        let mut incognito_proxy = Self::default();
        Self::init(&mut incognito_proxy, params);
        incognito_proxy
    }

    /// Initialize a lending market
    pub fn init(&mut self, params: TcOwners) {
        self.is_initialized = params.is_initialized;
        self.bump_seed = params.bump_seed;
        self.beacons = params.beacons;
    }
}

impl Sealed for TcOwners {}

impl Pack for TcOwners {
    /// 1 + 1 + 1 + 64 * 20
    const LEN: usize = 1315 - 32;
    fn unpack_from_slice(src: &[u8]) -> Result<Self, ProgramError> {
        let src = array_ref![src, 0, TcOwners::LEN];
        let (
            is_initialized,
            bump_seed,
            beacon_len,
            data_flat
        ) = array_refs![
            src, 
            1,
            1,
            1, 
            SECP256K1_PUBLIC_KEY_LENGTH * MAX_BEACON_ADDRESSES
        ];
        let is_initialized = match is_initialized {
            [0] => false,
            [1] => true,
            _ => return Err(BridgeError::InvalidBoolValue.into()),
        };

        let beacon_len = u8::from_le_bytes(*beacon_len);
        let mut beacons = Vec::with_capacity(beacon_len as usize + 1);
        let mut offset = 0;
        for _ in 0..beacon_len {
            let beacon_flat = array_ref![data_flat, offset, SECP256K1_PUBLIC_KEY_LENGTH];
            #[allow(clippy::ptr_offset_with_cast)]
            let new_beacon = Secp256k1Pubkey::new(beacon_flat);
            beacons.push(new_beacon);
            offset += SECP256K1_PUBLIC_KEY_LENGTH;
        }

        Ok(TcOwners {
            is_initialized,
            bump_seed: u8::from_le_bytes(*bump_seed),
            beacons
        })
    }

    fn pack_into_slice(&self, dst: &mut [u8]) {
        let dst = array_mut_ref![dst, 0, TcOwners::LEN];
        let (
            is_initialized,
            bump_seed,
            beacon_len,
            data_flat
        ) = mut_array_refs![
            dst, 
            1, 
            1,
            1, 
            SECP256K1_PUBLIC_KEY_LENGTH * MAX_BEACON_ADDRESSES
        ];
        *beacon_len = u8::try_from(self.beacons.len()).unwrap().to_le_bytes();
        *bump_seed = self.bump_seed.to_le_bytes();
        pack_bool(self.is_initialized, is_initialized);

        let mut offset = 0;
        // beacons
        for beacon in &self.beacons {
            let beacon_flat = array_mut_ref![data_flat, offset, SECP256K1_PUBLIC_KEY_LENGTH];
            #[allow(clippy::ptr_offset_with_cast)]
            beacon_flat.copy_from_slice(&beacon.to_bytes());
            offset += SECP256K1_PUBLIC_KEY_LENGTH;
        }

    }

}

// Init owner list
#[derive(Clone, Default)]
pub struct OwnerInit {
    /// beacon list
    pub beacons: Vec<Secp256k1Pubkey>
}

/// Reserve liquidity
#[derive(Clone, Debug, PartialEq)]
pub struct WithdrawRequest {
    // inst: list corresponding amount to transfer
    pub amounts: Vec<u64>,
    // signature 
    pub signatures: Vec<[u8; 65]>
}

fn pack_bool(boolean: bool, dst: &mut [u8; 1]) {
    *dst = (boolean as u8).to_le_bytes()
}

// sign data
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct SignData {
    // inst: list corresponding amount to transfer
    pub amounts: Vec<u64>,
    // keys
    pub accounts: Vec<[u8; 32]>,
    // nonce
    pub nonce: u64
}