[CmdletBinding()]
param(
    [string]$Version = $(if ($env:STFW_VERSION) { $env:STFW_VERSION } else { "latest" }),
    [string]$BinDir = $(if ($env:STFW_BINDIR) { $env:STFW_BINDIR } else { (Join-Path $HOME "bin") })
)

$ErrorActionPreference = "Stop"

$Repo = "scenario-test-framework/stfw"
$BinaryName = "stfw.exe"

function Resolve-LatestTag {
    $response = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -UseBasicParsing
    $baseResponse = $response.BaseResponse
    $finalUri = $null
    if (($baseResponse.PSObject.Properties.Name -contains "ResponseUri") -and $baseResponse.ResponseUri) {
        # Windows PowerShell 5.1 (.NET Framework / HttpWebResponse)
        $finalUri = $baseResponse.ResponseUri.AbsolutePath
    } elseif ($baseResponse.RequestMessage -and $baseResponse.RequestMessage.RequestUri) {
        # PowerShell 7+ (.NET / HttpResponseMessage)
        $finalUri = $baseResponse.RequestMessage.RequestUri.AbsolutePath
    }
    if (-not $finalUri) {
        throw "install.ps1: failed to resolve latest release URL"
    }
    $tag = $finalUri.TrimEnd("/").Split("/")[-1]
    if (-not $tag -or $tag -eq "latest") {
        throw "install.ps1: failed to resolve latest release tag"
    }
    return $tag
}

function Resolve-Version([string]$InputVersion) {
    if ($InputVersion -eq "latest") {
        $tag = Resolve-LatestTag
        return @{
            Tag = $tag
            Version = $tag.TrimStart("v")
        }
    }

    if ($InputVersion.StartsWith("v")) {
        return @{
            Tag = $InputVersion
            Version = $InputVersion.TrimStart("v")
        }
    }

    return @{
        Tag = "v$InputVersion"
        Version = $InputVersion
    }
}

function Add-UserPathIfMissing([string]$PathToAdd) {
    $currentUserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $userEntries = @()
    if ($currentUserPath) {
        $userEntries = $currentUserPath.Split(";") | Where-Object { $_ }
    }

    foreach ($entry in $userEntries) {
        if ([StringComparer]::OrdinalIgnoreCase.Equals($entry.TrimEnd('\'), $PathToAdd.TrimEnd('\'))) {
            return
        }
    }

    $newUserPath = if ($currentUserPath) { "$currentUserPath;$PathToAdd" } else { $PathToAdd }
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    Write-Warning "$PathToAdd was added to your user PATH. Open a new shell to use it."
}

if ($env:OS -ne "Windows_NT") {
    throw "install.ps1: this installer is for Windows PowerShell / PowerShell on Windows"
}

# Windows リリースは amd64 のみ (goreleaser が windows/arm64 を ignore)。
# ARM64 Windows は x64 エミュレーションで amd64 バイナリを実行できるため amd64 を使う。
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" {
        Write-Warning "install.ps1: Windows 版は amd64 のみ提供。ARM64 では x64 エミュレーションで実行します。"
        "amd64"
    }
    default { throw "install.ps1: unsupported arch: $env:PROCESSOR_ARCHITECTURE" }
}

$resolved = Resolve-Version $Version
$tag = $resolved.Tag
$resolvedVersion = $resolved.Version

$zipName = "stfw_${resolvedVersion}_windows_${arch}.zip"
$url = "https://github.com/$Repo/releases/download/$tag/$zipName"
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("stfw-install-" + [System.Guid]::NewGuid().ToString("N"))

New-Item -ItemType Directory -Path $tempDir | Out-Null
try {
    $zipPath = Join-Path $tempDir $zipName
    Write-Host "install.ps1: downloading $url"
    Invoke-WebRequest -Uri $url -OutFile $zipPath
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

    $binary = Get-ChildItem -Path $tempDir -Filter $BinaryName -File -Recurse | Select-Object -First 1
    if (-not $binary) {
        throw "install.ps1: $BinaryName not found in $zipName"
    }

    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    $destination = Join-Path $BinDir $BinaryName
    Copy-Item -Path $binary.FullName -Destination $destination -Force
    Write-Host "install.ps1: installed $BinaryName $tag to $destination"

    Add-UserPathIfMissing $BinDir
    & $destination --version
}
finally {
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force
    }
}
