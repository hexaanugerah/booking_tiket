# Jalankan dari root projek (Windows):
#   powershell -ExecutionPolicy Bypass -File .\scripts\run-local.ps1
# atau klik dua kali scripts\run-local.cmd
#
# Tidak perlu domain, DNS, Nginx, atau panel web. Semua berjalan lokal.

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

function Test-DockerReady {
    docker info *> $null
    return ($LASTEXITCODE -eq 0)
}

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw "Docker belum terpasang. Install Docker Desktop terlebih dahulu."
}

if (-not (Test-DockerReady)) {
    $desktop = Join-Path $env:ProgramFiles "Docker\Docker\Docker Desktop.exe"
    if (Test-Path $desktop) {
        Write-Host "Menjalankan Docker Desktop..."
        Start-Process -FilePath $desktop | Out-Null
        for ($i = 0; $i -lt 100; $i++) {
            Start-Sleep -Seconds 3
            if (Test-DockerReady) { break }
        }
    }
}

if (-not (Test-DockerReady)) {
    throw "Docker daemon belum berjalan. Buka Docker Desktop lalu jalankan ulang skrip ini."
}

docker compose version *> $null
if ($LASTEXITCODE -ne 0) {
    throw "Docker Compose plugin belum tersedia. Update Docker Desktop."
}

$envCreated = $false
if (-not (Test-Path ".env")) {
    Write-Host "Membuat .env lokal dengan akun demo..." -ForegroundColor Cyan
    $jwtAccess = [guid]::NewGuid().ToString("N") + [guid]::NewGuid().ToString("N")
    $jwtRefresh = [guid]::NewGuid().ToString("N") + [guid]::NewGuid().ToString("N")
    @"
# Dibuat otomatis oleh scripts/run-local.ps1
APP_ENV=local
FRONTEND_URL=http://localhost:8080,http://127.0.0.1:8080
BACKEND_PUBLIC_URL=http://localhost:8080
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
BIND_HOST=127.0.0.1
APP_PORT=8080

JWT_ACCESS_SECRET=$jwtAccess
JWT_REFRESH_SECRET=$jwtRefresh

SEED_SUPERADMIN_PASSWORD=superadmin123
SEED_ADMIN_PASSWORD=admin123
SEED_MANAGER_PASSWORD=manager123
SEED_PETUGAS_PASSWORD=petugas123
"@ | Set-Content -Path ".env" -Encoding ASCII
    $envCreated = $true
}

Write-Host ""
Write-Host "Membangun dan menjalankan container..." -ForegroundColor Cyan
docker compose up -d --build
if ($LASTEXITCODE -ne 0) {
    throw "docker compose gagal. Periksa pesan di atas."
}

Write-Host "Menunggu backend siap..." -ForegroundColor Cyan
$backendReady = $false
for ($i = 0; $i -lt 80; $i++) {
    try {
        $response = Invoke-WebRequest -UseBasicParsing -Uri "http://127.0.0.1:8080/healthz" -TimeoutSec 2
        if ($response.StatusCode -eq 200) {
            $backendReady = $true
            break
        }
    } catch {
        # Backend belum siap; ulangi.
    }
    Start-Sleep -Seconds 3
}

if (-not $backendReady) {
    docker compose logs --tail=120 backend
    throw "Backend tidak siap. Periksa log di atas."
}

Write-Host "Menunggu frontend siap..." -ForegroundColor Cyan
$frontendReady = $false
for ($i = 0; $i -lt 60; $i++) {
    try {
        $response = Invoke-WebRequest -UseBasicParsing -Uri "http://127.0.0.1:8080" -TimeoutSec 2
        if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 400) {
            $frontendReady = $true
            break
        }
    } catch {
        # Frontend belum siap; ulangi.
    }
    Start-Sleep -Seconds 3
}

if (-not $frontendReady) {
    docker compose logs --tail=120 frontend
    throw "Frontend tidak siap. Periksa log di atas."
}

Write-Host ""
Write-Host "Aplikasi siap dipakai:" -ForegroundColor Green
Write-Host "  Aplikasi : http://localhost:8080"
Write-Host "  API      : http://localhost:8080/api/v1"
Write-Host "  Login    : http://localhost:8080/login"
if ($envCreated) {
    Write-Host "  Superadmin: superadmin@kawahputih.id / superadmin123"
    Write-Host "  Admin     : admin@kawahputih.id / admin123"
}
Write-Host ""
Write-Host "Membuka browser..."
Start-Process "http://localhost:8080"
