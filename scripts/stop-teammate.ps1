$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$pidPath = Join-Path $root "logs\server.pid"
$stopped = $false

function Stop-ByPid {
    if (-not (Test-Path $pidPath)) {
        return $false
    }

    $pidText = (Get-Content $pidPath -ErrorAction SilentlyContinue | Select-Object -First 1).Trim()
    if (-not $pidText) {
        Remove-Item $pidPath -Force -ErrorAction SilentlyContinue
        return $false
    }

    $process = Get-Process -Id ([int]$pidText) -ErrorAction SilentlyContinue
    if ($process) {
        Stop-Process -Id $process.Id -Force
        $script:stopped = $true
    }

    Remove-Item $pidPath -Force -ErrorAction SilentlyContinue
    return $script:stopped
}

function Stop-ByPort {
    $listener = Get-NetTCPConnection -State Listen -LocalPort 8080 -ErrorAction SilentlyContinue | Select-Object -First 1
    if (-not $listener) {
        return $false
    }

    Stop-Process -Id $listener.OwningProcess -Force -ErrorAction SilentlyContinue
    $script:stopped = $true
    return $true
}

function Stop-ByProcessName {
    $targets = Get-Process server -ErrorAction SilentlyContinue
    if (-not $targets) {
        return $false
    }

    foreach ($process in $targets) {
        Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
        $script:stopped = $true
    }
    return $true
}

if (-not (Stop-ByPid)) {
    if (-not (Stop-ByPort)) {
        Stop-ByProcessName | Out-Null
    }
}

if ($stopped) {
    Write-Host "Online Judge stopped."
} else {
    Write-Host "No running Online Judge process found."
}
