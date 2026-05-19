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

# Test-only overrides used by installer regression tests. They are local-file
# inputs only and do not provide a checksum bypass for normal installs.
if ($env:NAMBA_INSTALL_TEST_VERSION) {
    $Version = $env:NAMBA_INSTALL_TEST_VERSION
}
if ($env:NAMBA_INSTALL_TEST_DIR) {
    $InstallDir = $env:NAMBA_INSTALL_TEST_DIR
}

function Copy-OrDownload {
    param(
        [string]$Source,
        [string]$Destination,
        [string]$ExpectedAssetName
    )

    if (Test-Path -LiteralPath $Source -PathType Leaf) {
        Copy-Item -LiteralPath $Source -Destination $Destination -Force
        return
    }

    try {
        Invoke-WebRequest -Uri $Source -OutFile $Destination -Headers @{ "User-Agent" = "NambaAI-Installer" }
    } catch {
        Write-DownloadError -Url $Source -ExpectedAssetName $ExpectedAssetName -ErrorRecord $_
    }
}

function Write-DownloadError {
    param(
        [string]$Url,
        [string]$ExpectedAssetName,
        [object]$ErrorRecord
    )

    $statusCode = $null
    if ($ErrorRecord.Exception.Response) {
        try {
            $statusCode = [int]$ErrorRecord.Exception.Response.StatusCode
        } catch {
        }
    }

    if ($statusCode -eq 404) {
        if ($Version -eq "latest") {
            throw "Failed to download $Url (404). No GitHub Release has been published yet, or the latest release does not contain $ExpectedAssetName. Publish a release first, or install from source with 'go install github.com/Nam-Cheol/namba-ai/cmd/namba@main'."
        }

        throw "Failed to download $Url (404). Release '$Version' was not found, or it does not contain $ExpectedAssetName."
    }

    throw "Failed to download $Url. Common causes: no published release, missing asset, repository access restrictions, or a network error. Original error: $($ErrorRecord.Exception.Message)"
}

function Get-ExpectedChecksum {
    param(
        [string]$ChecksumsFile,
        [string]$ExpectedAssetName
    )

    foreach ($line in Get-Content -LiteralPath $ChecksumsFile) {
        $trimmed = $line.Trim()
        if (-not $trimmed) {
            continue
        }
        $parts = $trimmed -split "\s+", 2
        if ($parts.Count -lt 2) {
            continue
        }
        $name = $parts[1].TrimStart("*")
        if ($name.StartsWith("./")) {
            $name = $name.Substring(2)
        }
        if ([System.IO.Path]::GetFileName($name) -eq $ExpectedAssetName -and -not [System.IO.Path]::IsPathRooted($name) -and $name -notmatch "(^|[\\/])\.\.([\\/]|$)") {
            return $parts[0]
        }
    }

    throw "checksums.txt does not contain an entry for $ExpectedAssetName."
}

function Test-ZipEntriesSafe {
    param([string]$ZipPath)

    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $zip = [System.IO.Compression.ZipFile]::OpenRead($ZipPath)
    try {
        foreach ($entry in $zip.Entries) {
            $name = $entry.FullName
            if ([System.IO.Path]::IsPathRooted($name) -or $name -match "(^|[\\/])\.\.([\\/]|$)") {
                throw "Archive contains unsafe paths and will not be extracted."
            }
        }
    } finally {
        $zip.Dispose()
    }
}

$repo = "Nam-Cheol/namba-ai"
$archSource = if ($env:NAMBA_INSTALL_TEST_ARCH) { $env:NAMBA_INSTALL_TEST_ARCH } elseif ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
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
    $checksumsUrl = "https://github.com/$repo/releases/latest/download/checksums.txt"
} else {
    $downloadUrl = "https://github.com/$repo/releases/download/$Version/$assetName"
    $checksumsUrl = "https://github.com/$repo/releases/download/$Version/checksums.txt"
}
if ($env:NAMBA_INSTALL_TEST_ASSET_PATH) {
    $downloadUrl = $env:NAMBA_INSTALL_TEST_ASSET_PATH
}
if ($env:NAMBA_INSTALL_TEST_CHECKSUMS_PATH) {
    $checksumsUrl = $env:NAMBA_INSTALL_TEST_CHECKSUMS_PATH
}

Write-Host "Installing NambaAI from $downloadUrl"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("namba-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null

try {
    $archivePath = Join-Path $tempRoot $assetName
    $checksumsPath = Join-Path $tempRoot "checksums.txt"
    Copy-OrDownload -Source $downloadUrl -Destination $archivePath -ExpectedAssetName $assetName
    Copy-OrDownload -Source $checksumsUrl -Destination $checksumsPath -ExpectedAssetName "checksums.txt"

    $expectedChecksum = (Get-ExpectedChecksum -ChecksumsFile $checksumsPath -ExpectedAssetName $assetName).ToLowerInvariant()
    $actualChecksum = (Get-FileHash -LiteralPath $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($expectedChecksum -ne $actualChecksum) {
        throw "Checksum verification failed for $assetName. Expected: $expectedChecksum Actual: $actualChecksum"
    }

    Test-ZipEntriesSafe -ZipPath $archivePath
    $extractRoot = Join-Path $tempRoot "extract"
    New-Item -ItemType Directory -Force -Path $extractRoot | Out-Null
    Expand-Archive -Path $archivePath -DestinationPath $extractRoot -Force

    $binary = Join-Path $extractRoot "namba.exe"
    if (-not (Test-Path -LiteralPath $binary -PathType Leaf)) {
        throw "namba.exe was not found in the downloaded archive."
    }

    $targetBinary = Join-Path $InstallDir "namba.exe"
    Copy-Item -LiteralPath $binary -Destination $targetBinary -Force

    if (-not $env:NAMBA_INSTALL_TEST_DIR) {
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
