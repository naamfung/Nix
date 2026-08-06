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
    $portableFiles = @(Get-ChildItem -LiteralPath $extractRoot -File -Filter "*.exe")
    if ($portableFiles.Count -ne 6) {
        throw "Portable archive must contain exactly 6 executables, found $($portableFiles.Count)"
    }
    foreach ($file in $portableFiles) {
        Assert-AuthenticodeSignature -Path $file.FullName
    }

    $portableSources = @{
        "inx-desktop.exe"       = "inx-desktop.exe"
        "inx-guard.exe"         = "inx-guard.exe"
        "inx-launcher.exe"      = "inx-launcher.exe"
        "Inx.exe"               = "inx-launcher.exe"
        "inx-update-helper.exe" = "inx-update-helper.exe"
        "inx-cli.exe"           = "inx-cli.exe"
    }
    foreach ($entry in $portableSources.GetEnumerator()) {
        $portableHash = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $extractRoot $entry.Key)).Hash
        $payloadHash = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $PayloadDirectory $entry.Value)).Hash
        if ($portableHash -ne $payloadHash) {
            throw "Portable $($entry.Key) does not match signed payload $($entry.Value)"
        }
    }
}
finally {
    if (Test-Path -LiteralPath $extractRoot) {
        Remove-Item -LiteralPath $extractRoot -Recurse -Force
    }
}

Write-Host "Windows Authenticode release contract verified."
