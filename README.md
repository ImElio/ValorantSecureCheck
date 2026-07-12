# ValorantSecureCheck

### This project is archived.

### Riot Vanguard now performs these compatibility checks internally, making this utility largely unnecessary.

### The source code remains available for educational purposes.

[![Release](https://img.shields.io/github/v/release/ImElio/ValorantSecureCheck?style=for-the-badge)](https://github.com/ImElio/ValorantSecureCheck/releases)
[![Issues](https://img.shields.io/github/issues/ImElio/ValorantSecureCheck?style=for-the-badge)](https://github.com/ImElio/ValorantSecureCheck/issues)
[![License](https://img.shields.io/github/license/ImElio/ValorantSecureCheck?style=for-the-badge)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![PowerShell](https://img.shields.io/badge/PowerShell-5.1+-5391FE?style=for-the-badge&logo=powershell)](https://learn.microsoft.com/powershell/)
[![Windows](https://img.shields.io/badge/Windows-10%2F11-0078D6?style=for-the-badge&logo=windows)](https://www.microsoft.com/windows)

---

**ValorantSecureCheck** is an open‑source **CLI** diagnostic tool for Windows that verifies whether a PC meets the **security requirements** to run **Valorant** and **Riot Vanguard**:
- **TPM 2.0**
- **Secure Boot**
- Basic hardware info (CPU, GPU, RAM, motherboard, OS)

> **Disclaimer**  
> ValorantSecureCheck **does not enable or modify** TPM or Secure Boot. It is **not a repair tool** and **not affiliated** with Riot Games, Valorant or Riot Vanguard. It only reads system state and presents it clearly.

---

## 🔗 Quick Links

- 📥 **Download**: latest builds on [Releases](https://github.com/ImElio/ValorantSecureCheck/releases)
- 🐞 **Issues / Support**: <https://github.com/ImElio/ValorantSecureCheck/issues>
- 🤝 **Contributing**: see [CONTRIBUTING.md](CONTRIBUTING.md)
- 📄 **License**: [MIT](LICENSE)

---

## ✨ Features

- ✅ Check **TPM 2.0** status (presence, readiness, vendor, version)
- ✅ Check **Secure Boot** status (UEFI state / registry / PowerShell API)
- ✅ Collect system info (CPU, GPU, RAM GiB, motherboard, OS)
- ✅ Output as **pretty table** or **JSON** for automation
- ✅ **PowerShell** scripts for quick one‑liners (TPM‑only, SecureBoot‑only, full)
- 🧩 Shared detection logic reusable across scripts and the Go CLI

---


### Table view
```text
------------------------------------------------------------------------------------------
 Checks
------------------------------------------------------------------------------------------
  Secure Boot        ✔ OK
  GPU                ✔ OK
  TPM 2.0            ✔ OK
  RAM >= 4 GiB       ✔ OK
  Motherboard        ✔ OK
  CPU                ✔ OK
------------------------------------------------------------------------------------------
```

### JSON output
```json
{
  "tpm": { "present": true, "ready": true, "isV2": true, "version": "2.0", "vendor": "INTC" },
  "secureBoot": { "enabled": true, "source": "registry" },
  "system": { "cpu": "Intel Core i7-13700K", "gpu": "NVIDIA GeForce RTX 3060 Ti", "ramGiB": 32, "motherboard": "MSI PRO Z790-P WIFI", "os": "Windows 11 Pro" },
  "checks": { "TPM2": true, "SecureBoot": true, "CPU": true, "GPU": true, "RAM>=4GiB": true, "Motherboard": true },
  "canRunValorant": true
}
```

---

### PowerShell scripts (In Administrator Mode)
```powershell
# Full check
iwr -useb https://raw.githubusercontent.com/ImElio/ValorantSecureCheck/main/scripts/ValorantSecureCheck.ps1 | iex

# TPM only
iwr -useb https://raw.githubusercontent.com/ImElio/ValorantSecureCheck/main/scripts/Get-TPMStatus.ps1 | iex

# Secure Boot only
iwr -useb https://raw.githubusercontent.com/ImElio/ValorantSecureCheck/main/scripts/Get-SecureBootStatus.ps1 | iex
```

## 📦 Distribution

- Portable **CLI**: `vsc.exe`
- **PowerShell** scripts: `scripts\*.ps1`
- (Roadmap) Desktop GUI will be added later.

---

## ❓ FAQ

**Does this tool fix TPM or Secure Boot?**  
No. It only **reads** system state and provides guidance links.

**Is this affiliated with Riot Games?**  
No. ValorantSecureCheck is independent and open source.

**Why JSON output?**  
For automation, support tickets, and telemetry (opt‑in).

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, coding standards, and PR process.

---

## 🧾 License & Notice

Licensed under [MIT](LICENSE).  
“Valorant” and “Riot Vanguard” are trademarks of Riot Games, Inc. This project is **not affiliated** with or endorsed by Riot Games.

---
