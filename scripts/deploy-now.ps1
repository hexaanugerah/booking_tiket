<#
.SYNOPSIS
  Deploy Kawah Putih ke server dari Windows dalam satu langkah.

.DESCRIPTION
  Skrip ini SSH ke server, lalu menjalankan scripts/bootstrap.sh di sana.
  Bootstrap akan clone/update kode, membuat .env, build, dan menyalakan
  seluruh stack Docker. Tidak perlu membuka dashboard/Coolify.

.EXAMPLE
  .\scripts\deploy-now.cmd
  .\scripts\deploy-now.cmd -Server 103.xx.xx.xx -User root -Port 22
#>
param(
    [string]$Server,
    [string]$User = "root",
    [int]$Port = 22
)

$ErrorActionPreference = "Stop"
$BootstrapUrl = "https://raw.githubusercontent.com/hexaanugerah/booking_tiket/main/scripts/bootstrap.sh"

if (-not (Get-Command ssh -ErrorAction SilentlyContinue)) {
    throw "OpenSSH client tidak ditemukan. Install 'OpenSSH Client' dari Windows Optional Features."
}

if ([string]::IsNullOrWhiteSpace($Server)) {
    $Server = Read-Host "Host/IP server (IP asli VPS, bukan domain Cloudflare)"
}
if ([string]::IsNullOrWhiteSpace($Server)) {
    throw "Host/IP server wajib diisi."
}

if ([string]::IsNullOrWhiteSpace($User)) {
    $User = "root"
}

$target = "$User@$Server"
Write-Host ""
Write-Host "==> Deploy ke $target (port $Port)" -ForegroundColor Cyan
Write-Host "    Perintah ini akan meminta password SSH/sudo bila diperlukan." -ForegroundColor DarkGray
Write-Host ""

# $() sengaja literal agar dieksekusi di server, bukan di PowerShell.
$remoteCommand = 'curl -fsSL "{0}" -o /tmp/kawahputih-bootstrap.sh && if [ "$(id -u)" -eq 0 ]; then bash /tmp/kawahputih-bootstrap.sh; else sudo bash /tmp/kawahputih-bootstrap.sh; fi' -f $BootstrapUrl

& ssh -t -p $Port $target $remoteCommand
if ($LASTEXITCODE -ne 0) {
    throw "Deploy gagal (exit code $LASTEXITCODE). Periksa pesan di atas."
}

Write-Host ""
Write-Host "[ok] Deploy selesai. URL aplikasi dan password superadmin ada di output di atas." -ForegroundColor Green
