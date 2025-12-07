<#
Apply Chronology schema files to the card_judge database.

Usage:
  powershell -ExecutionPolicy Bypass -File .\infrastructure\apply-chronology-schema.ps1
#>

[CmdletBinding()]
param()

function Write-ErrorAndExit([string]$msg) {
    Write-Error $msg
    exit 1
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$repoRoot = Resolve-Path -Path (Join-Path $scriptDir "..")

# Define SQL files in order
$sqlFiles = @(
    "src\static\sql\tables\LOBBY_GAME_TYPE.sql",
    "src\static\sql\tables\CHRONOLOGY_GAME.sql",
    "src\static\sql\tables\CHRONOLOGY_DRAW_PILE.sql",
    "src\static\sql\tables\CHRONOLOGY_CURRENT_CARD.sql",
    "src\static\sql\tables\CHRONOLOGY_PLAYER_TIMELINE.sql"
)

# Verify all files exist
foreach ($file in $sqlFiles) {
    $fullPath = Join-Path $repoRoot $file
    if (-not (Test-Path $fullPath)) {
        Write-ErrorAndExit "SQL file not found: $fullPath"
    }
}

# Check mysql client
if (-not (Get-Command mysql -ErrorAction SilentlyContinue)) {
    Write-ErrorAndExit "mysql client not found on PATH."
}

Write-Host "This script will apply 5 Chronology schema files to the card_judge database."
Write-Host "You will be prompted for the MySQL root password."
Write-Host ""

# Prompt for password
$rootPass = Read-Host -Prompt "Enter MySQL root password" -AsSecureString
$rootPassPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($rootPass))
$env:MYSQL_PWD = $rootPassPlain

$mysqlExe = (Get-Command mysql).Source
$mysqlArgs = @('-u', 'root', '-h', '127.0.0.1', 'card_judge')

try {
    foreach ($file in $sqlFiles) {
        $fullPath = Join-Path $repoRoot $file
        $fileName = Split-Path $file -Leaf
        Write-Host "Applying $fileName ..."
        
        $proc = Start-Process -FilePath $mysqlExe -ArgumentList $mysqlArgs -RedirectStandardInput $fullPath -NoNewWindow -Wait -PassThru
        
        if ($proc.ExitCode -ne 0) {
            Write-Error "Failed to apply $fileName (exit code $($proc.ExitCode))"
            Write-Host "You can try manually:"
            Write-Host "  mysql -u root -p -h 127.0.0.1 card_judge < $fullPath"
            exit $proc.ExitCode
        }
    }
} finally {
    Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "All Chronology schema files applied successfully!"
Write-Host "Verify with:"
Write-Host '  mysql -u root -p -h 127.0.0.1 -e "USE card_judge; SHOW TABLES LIKE ''CHRONOLOGY_%'';"'
exit 0
