param(
    [Parameter(Mandatory = $true)]
    [string]$PayloadDirectory,

    [Parameter(Mandatory = $true)]
    [string]$InstallerPath,

    [Parameter(Mandatory = $true)]
    [string]$PortableArchivePath,

    [switch]$RequireTrusted
)

$ErrorActionPreference = "Stop"

$expectedPayload = @(
    "inx-desktop.exe",
    "inx-guard.exe",
    "inx-launcher.exe",
    "inx-update-helper.exe",
    "inx-cli.exe",
    "inx-uninstall.exe"
)

function Assert-AuthenticodeSignature {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) {
        throw "Signed Windows artifact is missing: $Path"
    }
    $signature = Get-AuthenticodeSignature -LiteralPath $Path
    if ($null -eq $signature.SignerCertificate -or $signature.SignatureType -eq "None") {
        throw "Authenticode signature is missing: $Path"
    }
    if ($RequireTrusted -and $signature.Status -ne "Valid") {
        throw "Authenticode signature is not trusted for $Path`: $($signature.Status) $($signature.StatusMessage)"
    }
    Write-Host "Authenticode $($signature.Status): $Path"
}

$payloadFiles = @(Get-ChildItem -LiteralPath $PayloadDirectory -File -Filter "*.exe")
if ($payloadFiles.Count -ne $expectedPayload.Count) {
    throw "Payload must contain exactly $($expectedPayload.Count) executables, found $($payloadFiles.Count)"
}
foreach ($name in $expectedPayload) {
    Assert-AuthenticodeSignature -Path (Join-Path $PayloadDirectory $name)
}
Assert-AuthenticodeSignature -Path $InstallerPath

$extractRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("inx-authenticode-" + [guid]::NewGuid().ToString("N"))
try {
    Expand-Archive -LiteralPath $PortableArchivePath -DestinationPath $extractRoot

    # Legacy portable releases kept all six executables at InstallRoot. The
    # versioned-v1 layout deliberately keeps only the launcher aliases and CLI
    # at the root, while the active Desktop, update helper, and CLI live under
    # versions/vX.Y.Z/. Verify the exact layout selected by current.json instead
    # of treating the three versioned executables as missing.
    $currentPath = Join-Path $extractRoot "current.json"
    if (Test-Path -LiteralPath $currentPath -PathType Leaf) {
        $current = Get-Content -LiteralPath $currentPath -Raw | ConvertFrom-Json
        if ($current.schemaVersion -ne 1) {
            throw "Portable current.json schemaVersion must be 1"
        }
        $activeVersion = [string]$current.activeVersion
        $activeDir = [string]$current.activeDir
        if ($activeVersion -notmatch '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?$' -or
            [string]::IsNullOrWhiteSpace($activeDir) -or
            $activeDir.Replace("\", "/") -ne "versions/$activeVersion") {
            throw "Portable current.json must bind activeVersion to versions/<activeVersion>"
        }

        $activePath = [System.IO.Path]::GetFullPath((Join-Path $extractRoot $activeDir))
        $extractPrefix = [System.IO.Path]::GetFullPath($extractRoot).TrimEnd([char[]]@('\', '/')) + [System.IO.Path]::DirectorySeparatorChar
        if (-not $activePath.StartsWith($extractPrefix, [System.StringComparison]::OrdinalIgnoreCase) -or
            -not (Test-Path -LiteralPath $activePath -PathType Container)) {
            throw "Portable current.json activeDir escapes or is missing: $activeDir"
        }

        $portableSources = @(
            [pscustomobject]@{ Portable = "inx-launcher.exe"; Payload = "inx-launcher.exe" },
            [pscustomobject]@{ Portable = "Inx.exe"; Payload = "inx-launcher.exe" },
            [pscustomobject]@{ Portable = "inx-cli.exe"; Payload = "inx-cli.exe" },
            [pscustomobject]@{ Portable = (Join-Path $activeDir "inx-desktop.exe"); Payload = "inx-desktop.exe" },
            [pscustomobject]@{ Portable = (Join-Path $activeDir "inx-update-helper.exe"); Payload = "inx-update-helper.exe" },
            [pscustomobject]@{ Portable = (Join-Path $activeDir "inx-cli.exe"); Payload = "inx-cli.exe" }
        )
    }
    else {
        $portableSources = @(
            [pscustomobject]@{ Portable = "inx-desktop.exe"; Payload = "inx-desktop.exe" },
            [pscustomobject]@{ Portable = "inx-guard.exe"; Payload = "inx-guard.exe" },
            [pscustomobject]@{ Portable = "inx-launcher.exe"; Payload = "inx-launcher.exe" },
            [pscustomobject]@{ Portable = "Inx.exe"; Payload = "inx-launcher.exe" },
            [pscustomobject]@{ Portable = "inx-update-helper.exe"; Payload = "inx-update-helper.exe" },
            [pscustomobject]@{ Portable = "inx-cli.exe"; Payload = "inx-cli.exe" }
        )
    }

    $portableFiles = @(Get-ChildItem -LiteralPath $extractRoot -Recurse -File -Filter "*.exe")
    if ($portableFiles.Count -ne 6) {
        throw "Portable archive must contain exactly 6 executables, found $($portableFiles.Count)"
    }

    foreach ($entry in $portableSources) {
        $portablePath = Join-Path $extractRoot $entry.Portable
        Assert-AuthenticodeSignature -Path $portablePath
        $portableHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $portablePath).Hash
        $payloadHash = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $PayloadDirectory $entry.Payload)).Hash
        if ($portableHash -ne $payloadHash) {
            throw "Portable $($entry.Portable) does not match signed payload $($entry.Payload)"
        }
    }
}
finally {
    if (Test-Path -LiteralPath $extractRoot) {
        Remove-Item -LiteralPath $extractRoot -Recurse -Force
    }
}

Write-Host "Windows Authenticode release contract verified."
