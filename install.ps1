param(
    [string]$Version = "latest",
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

if (-not $PSBoundParameters.ContainsKey("Version") -and $env:NAMBA_VERSION) {
    $Version = $env:NAMBA_VERSION
}
if (-not $PSBoundParameters.ContainsKey("InstallDir") -and $env:NAMBA_INSTALL_DIR) {
    $InstallDir = $env:NAMBA_INSTALL_DIR
}

if (-not $InstallDir) {
    if ($env:LOCALAPPDATA) {
        $localAppData = $env:LOCALAPPDATA
    } else {
        $localAppData = Join-Path $env:USERPROFILE "AppData\Local"
    }
    $InstallDir = Join-Path $localAppData "Programs\NambaAI\bin"
}

$repo = "Nam-Cheol/namba-ai"
$archSource = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
$archKey = ""
if ($archSource) {
    $archKey = $archSource.ToLowerInvariant()
}
$arch = switch ($archKey) {
    "amd64" { "x86_64" }
    "x86" { "x86_64" }
    "x86_64" { "x86_64" }
    "arm64" { "arm64" }
    default { throw "Unsupported Windows architecture: $archSource" }
}

$assetName = "namba_Windows_$arch.zip"
if ($Version -eq "latest") {
    $downloadUrl = "https://github.com/$repo/releases/latest/download/$assetName"
} else {
    $downloadUrl = "https://github.com/$repo/releases/download/$Version/$assetName"
}

Write-Host "Installing NambaAI from $downloadUrl"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("namba-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null

try {
    $archivePath = Join-Path $tempRoot $assetName
    Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -Headers @{ "User-Agent" = "NambaAI-Installer" }
    Expand-Archive -Path $archivePath -DestinationPath $tempRoot -Force

    $binary = Get-ChildItem -Path $tempRoot -Filter "namba.exe" -File -Recurse | Select-Object -First 1
    if (-not $binary) {
        throw "namba.exe was not found in the downloaded archive."
    }

    $targetBinary = Join-Path $InstallDir "namba.exe"
    Copy-Item -Path $binary.FullName -Destination $targetBinary -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $parts = @()
    if ($userPath) {
        $parts = $userPath -split ";" | Where-Object { $_ }
    }
    if ($parts -notcontains $InstallDir) {
        if ($userPath) {
            $newUserPath = "$InstallDir;$userPath"
        } else {
            $newUserPath = $InstallDir
        }
        [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    }
    if (($env:Path -split ";") -notcontains $InstallDir) {
        $env:Path = "$InstallDir;$env:Path"
    }

    Write-Host ""
    Write-Host "NambaAI installed."
    Write-Host "Binary: $targetBinary"
    Write-Host "Command: namba"
    Write-Host ""
    Write-Host "If the command is not available in your current terminal, open a new terminal window."
} finally {
    if (Test-Path $tempRoot) {
        Remove-Item -Path $tempRoot -Recurse -Force
    }
}