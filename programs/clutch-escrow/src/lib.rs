use anchor_lang::prelude::*;

declare_id!("FdY9TYumZTvAAF5Tkpunfwg4kCzpb1bqCjke67yunoZb");

#[program]
pub mod clutch_escrow {
    use super::*;

    pub fn create_duel(ctx: Context<CreateDuel>, stake_usd: u64, deadline: i64) -> Result<()> {
        let duel = &mut ctx.accounts.duel;
        duel.creator = ctx.accounts.creator.key();
        duel.stake_usd = stake_usd;
        duel.deadline = deadline;
        duel.status = DuelStatus::PendingOpponent as u8;
        duel.bump = ctx.bumps.duel;
        Ok(())
    }

    pub fn accept_duel(ctx: Context<AcceptDuel>) -> Result<()> {
        let duel = &mut ctx.accounts.duel;
        require!(
            duel.status == DuelStatus::PendingOpponent as u8,
            ClutchError::InvalidStatus
        );
        duel.opponent = ctx.accounts.opponent.key();
        duel.status = DuelStatus::Active as u8;
        Ok(())
    }

    pub fn cancel_duel(ctx: Context<CancelDuel>) -> Result<()> {
        let duel = &mut ctx.accounts.duel;
        require!(
            duel.status == DuelStatus::PendingOpponent as u8,
            ClutchError::InvalidStatus
        );
        duel.status = DuelStatus::Cancelled as u8;
        Ok(())
    }
}

#[repr(u8)]
pub enum DuelStatus {
    PendingOpponent = 0,
    Active = 1,
    Cancelled = 9,
}

#[account]
pub struct Duel {
    pub creator: Pubkey,
    pub opponent: Pubkey,
    pub stake_usd: u64,
    pub deadline: i64,
    pub status: u8,
    pub bump: u8,
}

impl Duel {
    pub const LEN: usize = 8 + 32 + 32 + 8 + 8 + 1 + 1;
}

#[derive(Accounts)]
pub struct CreateDuel<'info> {
    #[account(mut)]
    pub creator: Signer<'info>,
    #[account(
        init,
        payer = creator,
        space = Duel::LEN,
        seeds = [b"duel", creator.key().as_ref()],
        bump
    )]
    pub duel: Account<'info, Duel>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct AcceptDuel<'info> {
    pub opponent: Signer<'info>,
    #[account(mut, has_one = creator)]
    pub duel: Account<'info, Duel>,
    /// CHECK: creator pubkey stored in duel
    pub creator: UncheckedAccount<'info>,
}

#[derive(Accounts)]
pub struct CancelDuel<'info> {
    pub creator: Signer<'info>,
    #[account(mut, has_one = creator)]
    pub duel: Account<'info, Duel>,
}

#[error_code]
pub enum ClutchError {
    #[msg("Invalid duel status")]
    InvalidStatus,
}
