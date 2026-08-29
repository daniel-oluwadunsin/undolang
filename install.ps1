param(
    [string]$SourceBinary = ".\undo.exe",
    [string]$InstallDir = "$env:LOCALAPPDATA\UndoLang\bin"
)

$ErrorActionPreference = "Stop"
$temporaryBinary = $null

try {
    if (-not (Test-Path -LiteralPath $SourceBinary -PathType Leaf)) {
        if ($SourceBinary -ne ".\undo.exe" -or -not (Test-Path -LiteralPath ".\go.mod" -PathType Leaf)) {
            throw "UndoLang binary not found: $SourceBinary"
        }
        $temporaryBinary = Join-Path ([System.IO.Path]::GetTempPath()) ("undolang-install-" + [System.Guid]::NewGuid() + ".exe")
        Write-Host "Building UndoLang from the local source tree..."
        $env:GOPROXY = "off"
        $env:GOTOOLCHAIN = "local"
        $env:CGO_ENABLED = "0"
        & go build -trimpath -buildvcs=false -o $temporaryBinary ./cmd/undo
        if ($LASTEXITCODE -ne 0) { throw "Go build failed" }
        $SourceBinary = $temporaryBinary
    }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $destination = Join-Path $InstallDir "undo.exe"
    Copy-Item -LiteralPath $SourceBinary -Destination $destination -Force
    Write-Host "Installed UndoLang at $destination"
    if (($env:PATH -split ";") -notcontains $InstallDir) {
        Write-Host "Add $InstallDir to your user PATH to invoke 'undo' directly."
    }
}
finally {
    if ($temporaryBinary -and (Test-Path -LiteralPath $temporaryBinary)) {
        Remove-Item -LiteralPath $temporaryBinary -Force
    }
}
