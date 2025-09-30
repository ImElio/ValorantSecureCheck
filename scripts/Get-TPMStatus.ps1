<#
.SYNOPSIS
  Robust TPM status (Windows 10/11). Handles PS 5.1 quirks, CIM/WMI fallbacks.
.PARAMETER Json
#>

param([switch]$Json)

function New-TpmObject {
  param(
    [bool]$Present=$false,[bool]$Ready=$false,[bool]$IsV2=$false,
    [string]$Version="",[string]$Vendor="",[string]$Source="",[string]$Raw=""
  )
  [pscustomobject]@{
    Present = $Present; Ready = $Ready; IsV2 = $IsV2
    Version = $Version; Vendor = $Vendor; Source = $Source; RawJson = $Raw
  }
}

function ConvertTo-PrettyJson([object]$o){ $o | ConvertTo-Json -Depth 6 | Out-String }

function Write-Table($title, [hashtable]$rows) {
  $w = try{$host.UI.RawUI.WindowSize.Width}catch{80}
  $boxW = [Math]::Min(80, $w - 4)
  $pad = { param($s) $s.PadRight($boxW-4) }
  $sep = ('-' * $boxW)
  Write-Host $sep
  "{0}{1}{0}" -f ' ', ($title.PadRight($boxW-2)) | Write-Host
  Write-Host $sep
  foreach($k in $rows.Keys){
    $line = "{0,-22} {1}" -f $k, $rows[$k]
    "  $line" | Write-Host
  }
  Write-Host $sep
}

$PSNativeCommandUseErrorActionPreference = $true
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [Text.Encoding]::UTF8
if ($PSStyle){ $PSStyle.OutputRendering = 'PlainText' }

try {
  $t = Get-Tpm | Select-Object TpmPresent,TpmReady,SpecVersion,ManufacturerIdTxt,ManufacturerVersionFull20
  $raw = ($t | ConvertTo-Json -Depth 4 -Compress)
  $valid = ($null -ne $t.TpmPresent -or $null -ne $t.TpmReady -or $t.SpecVersion)
  if ($valid) {
    $isV2 = ($t.SpecVersion -match '2\.0') -or ([string]::IsNullOrWhiteSpace($t.ManufacturerVersionFull20) -eq $false)
    $ver  = if([string]::IsNullOrWhiteSpace($t.SpecVersion) -and $isV2){'2.0'} else {$t.SpecVersion}
    $obj  = New-TpmObject -Present:$t.TpmPresent -Ready:$t.TpmReady -IsV2:$isV2 -Version:$ver -Vendor:$t.ManufacturerIdTxt -Source 'Get-Tpm' -Raw $raw
    if ($Json) { ConvertTo-PrettyJson $obj; return } else {
      Write-Table "TPM Status" @{
        "Present / Ready" = ("{0} / {1}" -f $obj.Present,$obj.Ready)
        "Version"         = $obj.Version
        "Vendor"          = $obj.Vendor
        "Source"          = $obj.Source
      }
      return
    }
  }
} catch { $err1 = $_ }

try {
  $c = Get-CimInstance -Namespace 'root/cimv2/Security/MicrosoftTpm' -ClassName Win32_Tpm |
       Select-Object IsEnabled_InitialValue,IsActivated_InitialValue,SpecVersion,ManufacturerIdTxt,ManufacturerVersionFull20
  $raw = ($c | ConvertTo-Json -Depth 4 -Compress)
  if ($c) {
    $present = $c.IsEnabled_InitialValue -or $c.IsActivated_InitialValue
    $ready   = $c.IsEnabled_InitialValue -and $c.IsActivated_InitialValue
    $isV2    = ($c.SpecVersion -match '2\.0') -or ([string]::IsNullOrWhiteSpace($c.ManufacturerVersionFull20) -eq $false)
    $ver     = if([string]::IsNullOrWhiteSpace($c.SpecVersion) -and $isV2){'2.0'} else {$c.SpecVersion}
    $obj = New-TpmObject -Present:$present -Ready:$ready -IsV2:$isV2 -Version:$ver -Vendor:$c.ManufacturerIdTxt -Source 'CIM' -Raw $raw
    if ($Json) { ConvertTo-PrettyJson $obj; return } else {
      Write-Table "TPM Status" @{
        "Present / Ready" = ("{0} / {1}" -f $obj.Present,$obj.Ready)
        "Version"         = $obj.Version
        "Vendor"          = $obj.Vendor
        "Source"          = $obj.Source
      }
      return
    }
  }
} catch { $err2 = $_ }

try {
  $w = Get-WmiObject -Namespace 'root\CIMV2\Security\MicrosoftTpm' -Class Win32_Tpm |
       Select-Object IsEnabled_InitialValue,IsActivated_InitialValue,SpecVersion,ManufacturerIdTxt,ManufacturerVersionFull20
  $raw = ($w | ConvertTo-Json -Depth 4 -Compress)
  if ($w) {
    $present = $w.IsEnabled_InitialValue -or $w.IsActivated_InitialValue
    $ready   = $w.IsEnabled_InitialValue -and $w.IsActivated_InitialValue
    $isV2    = ($w.SpecVersion -match '2\.0') -or ([string]::IsNullOrWhiteSpace($w.ManufacturerVersionFull20) -eq $false)
    $ver     = if([string]::IsNullOrWhiteSpace($w.SpecVersion) -and $isV2){'2.0'} else {$w.SpecVersion}
    $obj = New-TpmObject -Present:$present -Ready:$ready -IsV2:$isV2 -Version:$ver -Vendor:$w.ManufacturerIdTxt -Source 'WMI' -Raw $raw
    if ($Json) { ConvertTo-PrettyJson $obj; return } else {
      Write-Table "TPM Status" @{
        "Present / Ready" = ("{0} / {1}" -f $obj.Present,$obj.Ready)
        "Version"         = $obj.Version
        "Vendor"          = $obj.Vendor
        "Source"          = $obj.Source
      }
      return
    }
  }
} catch { $err3 = $_ }

$obj = New-TpmObject -Raw ($err1,$err2,$err3 | Out-String) -Source 'none'
if ($Json) { ConvertTo-PrettyJson $obj } else { Write-Table "TPM Status" @{"Present / Ready"="false / false"; "Version"=""; "Vendor"=""; "Source"=$obj.Source} }
