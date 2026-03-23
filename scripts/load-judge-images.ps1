param(
    [string]$ArchivePath = ""
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot

if ([string]::IsNullOrWhiteSpace($ArchivePath)) {
    $ArchivePath = Join-Path $root "artifacts\judge-images\algojudge-judge-images.tar"
}

if (-not (Test-Path $ArchivePath)) {
    throw "Archive not found: $ArchivePath"
}

docker load -i $ArchivePath

Write-Host "Loaded judge images from:"
Write-Host "  $ArchivePath"
