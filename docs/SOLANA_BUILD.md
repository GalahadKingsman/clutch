# Solana program — сборка на Mac (devnet)

> Без Helius: `SOLANA_RPC_URL=https://api.devnet.solana.com`

## Частые ошибки

| Ошибка | Причина | Решение |
|--------|---------|---------|
| `Found argument '#'` | В команде есть комментарий `# ...` | Только команда, без `#` |
| `lock file version 4` | Новый Cargo vs Anchor SBF | В репо уже `Cargo.lock` v3 с пинами — не удаляй |
| `edition2024` при build | crates.io обновился, SBF cargo старый | Используй закоммиченный `Cargo.lock` |
| `anchor build` 0.12s, нет `.so` | Не вызван SBF output | См. команды ниже |
| `idl-build` / `source_file` | IDL на новом proc-macro2 | `anchor build --no-idl` |

## Сборка (из корня репо)

```bash
export PATH="$HOME/.local/share/solana/install/active_release/bin:$HOME/.avm/bin:$HOME/.cargo/bin:$PATH"
source "$HOME/.cargo/env" 2>/dev/null

solana config set --url devnet
solana balance   # нужно ~1+ SOL

cd /path/to/clutch
avm use 0.30.1

# Сборка .so (обязательно sbf-out-dir)
cargo build-sbf --manifest-path programs/clutch-escrow/Cargo.toml --sbf-out-dir target/deploy

ls -la target/deploy/clutch_escrow.so
```

## Деплой devnet

```bash
solana program deploy target/deploy/clutch_escrow.so \
  --program-id target/deploy/clutch_escrow-keypair.json
```

Program ID печатается в конце (тот же, что в `Anchor.toml` после sync).

## .env на VPS

```env
SOLANA_RPC_URL=https://api.devnet.solana.com
CLUTCH_PROGRAM_ID=FdY9TYumZTvAAF5Tkpunfwg4kCzpb1bqCjke67yunoZb
CLUTCH_TREASURY_PUBKEY=<твой solana address>
```

## Текущий devnet Program ID

`FdY9TYumZTvAAF5Tkpunfwg4kCzpb1bqCjke67yunoZb`

Explorer: https://explorer.solana.com/address/FdY9TYumZTvAAF5Tkpunfwg4kCzpb1bqCjke67yunoZb?cluster=devnet
