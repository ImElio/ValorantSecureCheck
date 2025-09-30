<#
.SYNOPSIS
  Full Valorant check: TPM 2.0 + Secure Boot + HW info.
.PARAMETER Json
#>
param([switch]$Json)

$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
if ($PSStyle) { $PSStyle.OutputRendering = 'PlainText' }

function Box($title, [hashtable]$rows) {
  $w = try { $host.UI.RawUI.WindowSize.Width } catch { 80 }
  $boxW = [Math]::Min(90, $w - 4)
  $sep = ('-' * $boxW)
  Write-Host $sep
  "{0}{1}{0}" -f ' ', ($title.PadRight($boxW - 2)) | Write-Host
  Write-Host $sep
  foreach ($k in $rows.Keys) {
    $line = "{0,-18} {1}" -f $k, $rows[$k]
    "  $line" | Write-Host
  }
  Write-Host $sep
}

function Get-Tpm-Obj {
  try {
    $t = Get-Tpm | Select-Object TpmPresent,TpmReady,SpecVersion,ManufacturerIdTxt,ManufacturerVersionFull20
    $isV2 = ($t.SpecVersion -match '2\.0') -or ([string]::IsNullOrWhiteSpace($t.ManufacturerVersionFull20) -eq $false)
    $ver  = if ([string]::IsNullOrWhiteSpace($t.SpecVersion) -and $isV2) { '2.0' } else { $t.SpecVersion }
    if ($null -ne $t.TpmPresent -or $null -ne $t.TpmReady -or $t.SpecVersion) {
      return [pscustomobject]@{
        Present = $t.TpmPresent
        Ready   = $t.TpmReady
        IsV2    = $isV2
        Version = $ver
        Vendor  = $t.ManufacturerIdTxt
        Source  = 'Get-Tpm'
      }
    }
  } catch {}

  try {
    $c = Get-CimInstance -Namespace 'root/cimv2/Security/MicrosoftTpm' -ClassName Win32_Tpm |
         Select-Object IsEnabled_InitialValue,IsActivated_InitialValue,SpecVersion,ManufacturerIdTxt,ManufacturerVersionFull20
    if ($c) {
      $present = $c.IsEnabled_InitialValue -or $c.IsActivated_InitialValue
      $ready   = $c.IsEnabled_InitialValue -and $c.IsActivated_InitialValue
      $isV2    = ($c.SpecVersion -match '2\.0') -or ([string]::IsNullOrWhiteSpace($c.ManufacturerVersionFull20) -eq $false)
      $ver     = if ([string]::IsNullOrWhiteSpace($c.SpecVersion) -and $isV2) { '2.0' } else { $c.SpecVersion }
      return [pscustomobject]@{
        Present = $present
        Ready   = $ready
        IsV2    = $isV2
        Version = $ver
        Vendor  = $c.ManufacturerIdTxt
        Source  = 'CIM'
      }
    }
  } catch {}

  try {
    $w = Get-WmiObject -Namespace 'root\CIMV2\Security\MicrosoftTpm' -Class Win32_Tpm |
         Select-Object IsEnabled_InitialValue,IsActivated_InitialValue,SpecVersion,ManufacturerIdTxt,ManufacturerVersionFull20
    if ($w) {
      $present = $w.IsEnabled_InitialValue -or $w.IsActivated_InitialValue
      $ready   = $w.IsEnabled_InitialValue -and $w.IsActivated_InitialValue
      $isV2    = ($w.SpecVersion -match '2\.0') -or ([string]::IsNullOrWhiteSpace($w.ManufacturerVersionFull20) -eq $false)
      $ver     = if ([string]::IsNullOrWhiteSpace($w.SpecVersion) -and $isV2) { '2.0' } else { $w.SpecVersion }
      return [pscustomobject]@{
        Present = $present
        Ready   = $ready
        IsV2    = $isV2
        Version = $ver
        Vendor  = $w.ManufacturerIdTxt
        Source  = 'WMI'
      }
    }
  } catch {}

  return [pscustomobject]@{ Present = $false; Ready = $false; IsV2 = $false; Version = ""; Vendor = ""; Source = 'none' }
}

function Get-SecureBoot-Obj {
  try {
    $k = 'HKLM:\System\CurrentControlSet\Control\SecureBoot\State'
    $v = (Get-ItemProperty -Path $k -Name UEFISecureBootEnabled -ErrorAction Stop).UEFISecureBootEnabled
    return [pscustomobject]@{ Enabled = ($v -eq 1); Source = 'registry' }
  } catch {}
  try {
    $v = Confirm-SecureBootUEFI
    return [pscustomobject]@{ Enabled = $v; Source = 'powershell' }
  } catch {}
  return [pscustomobject]@{ Enabled = $false; Source = 'unknown' }
}

$cpu = (Get-CimInstance Win32_Processor | Select-Object -First 1 -ExpandProperty Name)
$gpu = (Get-CimInstance Win32_VideoController | Where-Object {
  $_.Name -notmatch 'microsoft basic display|basic render|virtual|meta virtual monitor'
} | Select-Object -First 1 -ExpandProperty Name)
if (-not $gpu) { $gpu = (Get-CimInstance Win32_VideoController | Select-Object -First 1 -ExpandProperty Name) }
$memBytes = [int64]((Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory)
$ramGiB = [int]($memBytes / 1GB)
$bb = Get-CimInstance Win32_BaseBoard | Select-Object -First 1 Manufacturer,Product
$mobo = ($bb.Manufacturer + ' ' + $bb.Product).Trim()
$os = (Get-CimInstance Win32_OperatingSystem | Select-Object -First 1 -ExpandProperty Caption)

$tpm = Get-Tpm-Obj
$sb  = Get-SecureBoot-Obj

$checks = [ordered]@{
  'TPM 2.0'       = ($tpm.Present -and $tpm.Ready -and $tpm.IsV2)
  'Secure Boot'   = $sb.Enabled
  'CPU'           = -not [string]::IsNullOrWhiteSpace($cpu)
  'GPU'           = -not [string]::IsNullOrWhiteSpace($gpu)
  'RAM >= 4 GiB'  = ($ramGiB -ge 4)
  'Motherboard'   = -not [string]::IsNullOrWhiteSpace($mobo)
}

$result = [pscustomobject]@{
  tpm            = $tpm
  secureBoot     = $sb
  system         = [pscustomobject]@{ cpu = $cpu; gpu = $gpu; ramGiB = $ramGiB; motherboard = $mobo; os = $os }
  checks         = [pscustomobject]@{
                     TPM2        = $checks['TPM 2.0']
                     SecureBoot  = $checks['Secure Boot']
                     CPU         = $checks['CPU']
                     GPU         = $checks['GPU']
                     'RAM>=4GiB' = $checks['RAM >= 4 GiB']
                     Motherboard = $checks['Motherboard']
                   }
  canRunValorant = ($checks.Values -notcontains $false)
}

if ($Json) {
  $result | ConvertTo-Json -Depth 6
  return
}

$OK =  ([char]0x2714) + ' OK'
$NO =  ([char]0x2717) + ' Not OK'

Box "Checks" @{
  'TPM 2.0'      = $( if ($checks['TPM 2.0'])      { $OK } else { $NO } )
  'Secure Boot'  = $( if ($checks['Secure Boot'])  { $OK } else { $NO } )
  'CPU'          = $( if ($checks['CPU'])          { $OK } else { $NO } )
  'GPU'          = $( if ($checks['GPU'])          { $OK } else { $NO } )
  'RAM >= 4 GiB' = $( if ($checks['RAM >= 4 GiB']) { $OK } else { $NO } )
  'Motherboard'  = $( if ($checks['Motherboard'])  { $OK } else { $NO } )
}

$tpmVersionDisplay = if ($tpm.Version) { $tpm.Version } elseif ($tpm.IsV2) { '2.0' } else { '' }

Box "Details" @{
  'TPM Present/Ready' = ("{0} / {1}" -f $tpm.Present, $tpm.Ready)
  'TPM Version'       = $tpmVersionDisplay
  'TPM Vendor'        = $tpm.Vendor
  'Secure Boot'       = $sb.Enabled
  'CPU'               = $cpu
  'GPU'               = $gpu
  'RAM (GiB)'         = $ramGiB
  'Motherboard'       = $mobo
  'OS'                = $os
}


$checkChar = [char]0x2714  # ✓
$crossChar = [char]0x2717  # ✗
if ($result.canRunValorant) {
  Write-Host "`nResult: READY $checkChar" -ForegroundColor Green
} else {
  Write-Host "`nResult: NOT READY $crossChar" -ForegroundColor Red
}
