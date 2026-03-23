$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot

docker build -t algojudge/judge-cpp:latest "$root\docker\judger\cpp"
docker build -t algojudge/judge-java:latest "$root\docker\judger\java"
docker build -t algojudge/judge-python:latest "$root\docker\judger\python"
