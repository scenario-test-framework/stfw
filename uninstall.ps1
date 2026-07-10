[CmdletBinding()]
param(
    [string]$BinDir = $(if ($env:STFW_BINDIR) { $env:STFW_BINDIR } else { (Join-Path $HOME "bin") })
)

$ErrorActionPreference = "Stop"
$BinaryName = "stfw.exe"

if ($env:OS -ne "Windows_NT") {
    throw "uninstall.ps1: this uninstaller is for Windows PowerShell / PowerShell on Windows"
}

$destination = Join-Path $BinDir $BinaryName
if (-not (Test-Path $destination)) {
    throw "uninstall.ps1: $destination does not exist"
}

Remove-Item -Path $destination -Force
Write-Host "uninstall.ps1: removed $destination"
Write-Host "uninstall.ps1: if needed, remove $BinDir from your user PATH manually"
