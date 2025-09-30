# Contributing to ValorantSecureCheck

Thanks for your interest in contributing! This project aims to provide a reliable, open‑source diagnostic tool for Valorant & Riot Vanguard on Windows.

## Code of Conduct
Be respectful and constructive. We welcome first‑time contributors. Harassment or discrimination of any kind is not tolerated.

## How to Contribute
1. **Fork** the repository and create a feature branch from `main`:
   ```bash
   git checkout -b feat/short-description
   ```
2. Make your changes with tests where possible.
3. Run formatting and linters.
4. Open a **Pull Request** with a clear description and screenshots/logs when applicable.

## Project Layout
```
scripts/           # PowerShell primary CLI scripts
cmd/vsc/           # Optional Go CLI (TUI/JSON)
pkg/system/        # Core detection logic (reusable)
installer/         # Packaging (Inno Setup/MSIX) - optional
```

## Development Guidelines

### PowerShell
- Target **Windows 10/11**, PowerShell **5.1+**.
- Keep scripts idempotent and read‑only (no registry writes).
- Use UTF‑8 (with BOM where needed); avoid inline `if` inside hashtables for PS 5.1 compatibility.
- Run **PSScriptAnalyzer**:
  ```powershell
  Install-Module PSScriptAnalyzer -Scope CurrentUser
  Invoke-ScriptAnalyzer -Path scripts -Recurse
  ```

### Go (CLI)
- Go **1.22+**. Run `go fmt ./...` and `go vet ./...`.
- Prefer explicit error handling, no panics in CLI paths.
- Unit tests welcome for `pkg/system` (mockable via interfaces).

### Commit Messages
- Conventional prefix: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `build:`
- Keep subject under ~72 chars and include context if touching multiple modules.

## Issue Triage
- **bug:** wrong detection/edge cases
- **enhancement:** UX or new features
- **docs:** README/guides updates
- Please include OS build, motherboard vendor/model, and JSON output if relevant.

## Release Process
1. Update version and CHANGELOG.
2. Build artifacts:
   - `vsc.exe` (Go CLI)
   - ZIP portable with `scripts/`
3. Tag and create a GitHub Release with checksums.
4. (Optional) Sign binaries and scripts (Authenticode).

## Security
If you discover a security issue, **do not** open a public issue. Email the maintainers or use GitHub Security Advisories.

Thanks for helping improve ValorantSecureCheck! 🎯
