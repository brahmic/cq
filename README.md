# CQ (Codex Quota Monitor)

A TUI for monitoring Codex quota written in Go using [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Demo](demo.gif)

## Features

- Multiple accounts from local app storage, OpenCode, and Codex auth files
- OAuth authentication via browser (`n`)
- Apply active account to OpenCode auth (`o`)
- Quota display for primary and secondary windows when available

## Installation

```bash
go install github.com/deLiseLINO/codex-quota/cmd/cq@latest
```

Build from source:

```bash
git clone https://github.com/deLiseLINO/codex-quota.git
cd codex-quota
go install ./cmd/cq
```

## Usage

Run the app:

```bash
cq
```

On first launch press `n` to add an account via OAuth.

## Controls

- `n` - add new account
- `left/right` (or `h/l`) - switch between accounts
- `r` - refresh data
- `i` - toggle additional info
- `o` - apply active account to OpenCode auth
- `x` - delete current account (only locally added app accounts; external accounts are read-only)
- `q` (or `esc`) - quit
