$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$outputDir = Join-Path $root "artifacts\judge-images"
$archivePath = Join-Path $outputDir "algojudge-judge-images.tar"
$hashPath = Join-Path $outputDir "algojudge-judge-images.sha256.txt"

$images = @(
    "algojudge/judge-cpp:latest",
    "algojudge/judge-java:latest",
    "algojudge/judge-python:latest"
)

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

foreach ($image in $images) {
    docker image inspect $image | Out-Null
}

if (Test-Path $archivePath) {
    Remove-Item $archivePath -Force
}

docker save -o $archivePath $images

$hash = Get-FileHash -Algorithm SHA256 $archivePath
"$($hash.Hash)  $(Split-Path -Leaf $archivePath)" | Set-Content -Path $hashPath -Encoding ascii

Write-Host "Packaged judge images:"
Write-Host "  $archivePath"
Write-Host "SHA256:"
Write-Host "  $hashPath"
