<#
.SYNOPSIS
  Secure Boot status checker
.PARAMETER Json
#>

param([switch]$Json)

function New-SbObject([bool]$Enabled=$false,[string]$Source="",[string]$Raw=""){
  [pscustomobject]@{ Enabled=$Enabled; Source=$Source; Raw=$Raw }
}
function ConvertTo-PrettyJson([object]$o){ $o | ConvertTo-Json -Depth 4 | Out-String }
function Write-Table($title, [hashtable]$rows) {
  $w = try{$host.UI.RawUI.WindowSize.Width}catch{80}
  $boxW = [Math]::Min(80, $w - 4)
  $sep = ('-' * $boxW)
  Write-Host $sep
  "{0}{1}{0}" -f ' ', ($title.PadRight($boxW-2)) | Write-Host
  Write-Host $sep
  foreach($k in $rows.Keys){ "  {0,-22} {1}" -f $k, $rows[$k] | Write-Host }
  Write-Host $sep
}

try{
  $k = 'HKLM:\System\CurrentControlSet\Control\SecureBoot\State'
  $v = (Get-ItemProperty -Path $k -Name UEFISecureBootEnabled -ErrorAction Stop).UEFISecureBootEnabled
  $obj = New-SbObject -Enabled:($v -eq 1) -Source 'registry'
  if($Json){ ConvertTo-PrettyJson $obj } else { Write-Table "Secure Boot" @{ "Enabled"=$obj.Enabled; "Source"=$obj.Source } }
  return
}catch{ $err1 = $_ }

try{
  $v = Confirm-SecureBootUEFI
  $obj = New-SbObject -Enabled:$v -Source 'powershell'
  if($Json){ ConvertTo-PrettyJson $obj } else { Write-Table "Secure Boot" @{ "Enabled"=$obj.Enabled; "Source"=$obj.Source } }
  return
}catch{ $err2 = $_ }

$obj = New-SbObject -Enabled:$false -Source 'unknown' -Raw ($err1,$err2 | Out-String)
if($Json){ ConvertTo-PrettyJson $obj } else { Write-Table "Secure Boot" @{ "Enabled"=$obj.Enabled; "Source"=$obj.Source } }
