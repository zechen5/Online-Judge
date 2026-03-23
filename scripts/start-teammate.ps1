param(
    [switch]$NoBrowser
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$logsDir = Join-Path $root "logs"
$stdoutPath = Join-Path $logsDir "server.log"
$stderrPath = Join-Path $logsDir "server.err.log"
$pidPath = Join-Path $logsDir "server.pid"
$archivePath = Join-Path $root "artifacts\judge-images\algojudge-judge-images.tar"
$healthzUrl = "http://127.0.0.1:8080/healthz"
$siteUrl = "http://127.0.0.1:8080/"
$requiredImages = @(
    "algojudge/judge-cpp:latest",
    "algojudge/judge-java:latest",
    "algojudge/judge-python:latest"
)

function Assert-CommandExists {
    param([string]$Name)

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Missing required command: $Name"
    }
}

function Invoke-DockerQuiet {
    param([string]$Arguments)

    & cmd /c "docker $Arguments >nul 2>nul"
    return $LASTEXITCODE
}

function Test-ImageExists {
    param([string]$Image)

    return (Invoke-DockerQuiet -Arguments "image inspect $Image") -eq 0
}

function Ensure-DockerRunning {
    if ((Invoke-DockerQuiet -Arguments "version") -ne 0) {
        throw "Docker is installed but not running. Please start Docker Desktop first."
    }
}

function Ensure-JudgeImages {
    $missing = @()
    foreach ($image in $requiredImages) {
        if (-not (Test-ImageExists -Image $image)) {
            $missing += $image
        }
    }

    if ($missing.Count -eq 0) {
        Write-Host "Judge images already exist."
        return
    }

    if (Test-Path $archivePath) {
        Write-Host "Judge image archive found. Loading local bundle..."
        & powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "load-judge-images.ps1") -ArchivePath $archivePath
    } else {
        Write-Host "Judge image archive not found. Building images from Dockerfiles..."
        & powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot "build-judge-images.ps1")
    }

    $stillMissing = @()
    foreach ($image in $requiredImages) {
        if (-not (Test-ImageExists -Image $image)) {
            $stillMissing += $image
        }
    }

    if ($stillMissing.Count -gt 0) {
        throw "Judge images are still missing: $($stillMissing -join ', ')"
    }
}

function Test-Healthz {
    try {
        $response = Invoke-RestMethod -Uri $healthzUrl -TimeoutSec 2
        return $response.status -eq "ok"
    } catch {
        return $false
    }
}

function Assert-PortAvailable {
    $listener = Get-NetTCPConnection -State Listen -LocalPort 8080 -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($listener -and -not (Test-Healthz)) {
        throw "Port 8080 is already occupied by another process (PID $($listener.OwningProcess))."
    }
}

function Start-ServerProcess {
    New-Item -ItemType Directory -Force -Path $logsDir | Out-Null

    if (Test-Path $stdoutPath) {
        Remove-Item $stdoutPath -Force
    }
    if (Test-Path $stderrPath) {
        Remove-Item $stderrPath -Force
    }
    if (Test-Path $pidPath) {
        Remove-Item $pidPath -Force
    }

    return Start-Process -FilePath "go" `
        -ArgumentList "run ./cmd/server" `
        -WorkingDirectory $root `
        -RedirectStandardOutput $stdoutPath `
        -RedirectStandardError $stderrPath `
        -WindowStyle Hidden `
        -PassThru
}

Assert-CommandExists -Name "go"
Assert-CommandExists -Name "docker"
Ensure-DockerRunning
Ensure-JudgeImages
$env:OJ_JUDGE_EXECUTOR = "docker"

if (Test-Healthz) {
    Write-Host "Server already running at $siteUrl"
    if (-not $NoBrowser) {
        Start-Process $siteUrl
    }
    return
}

Assert-PortAvailable

$process = Start-ServerProcess
Set-Content -Path $pidPath -Value $process.Id -Encoding ascii

$ready = $false
for ($i = 0; $i -lt 40; $i++) {
    Start-Sleep -Milliseconds 500
    if (Test-Healthz) {
        $ready = $true
        break
    }
    if ($process.HasExited) {
        break
    }
}

if (-not $ready) {
    if (Test-Path $pidPath) {
        Remove-Item $pidPath -Force
    }
    if ($process.HasExited) {
        throw "Server exited during startup. Check logs\server.err.log"
    }
    throw "Server did not become ready in time. Check logs\server.log and logs\server.err.log"
}

Write-Host "Online Judge is ready:"
Write-Host "  $siteUrl"
Write-Host "Default admin:"
Write-Host "  username: admin"
Write-Host "  password: admin123456"
Write-Host "Logs:"
Write-Host "  $stdoutPath"
Write-Host "  $stderrPath"
Write-Host "PID:"
Write-Host "  $pidPath"

if (-not $NoBrowser) {
    Start-Process $siteUrl
}
