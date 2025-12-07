#!/usr/bin/env pwsh
# Check Chronology game data in the database

param(
    [string]$ContainerName = $null
)

# Get database credentials from environment
$dbUser = $env:CARD_JUDGE_SQL_USER
$dbPassword = $env:CARD_JUDGE_SQL_PASSWORD
$dbHost = $env:CARD_JUDGE_SQL_HOST
$dbName = "card_judge"

if (-not $dbUser) { $dbUser = "root" }
if (-not $dbHost) { $dbHost = "127.0.0.1" }

# Check if running in Docker or local
$isDocker = $null -ne $ContainerName -and $ContainerName -ne ""

Write-Host "Checking Chronology game data..." -ForegroundColor Cyan

if ($isDocker) {
    Write-Host "Using Docker container: $ContainerName" -ForegroundColor Yellow
    
    # Read SQL file content
    $sqlContent = Get-Content -Path ".\infrastructure\check-chronology-data.sql" -Raw
    
    # Execute in Docker container
    $sqlContent | docker exec -i $ContainerName sh -c "mysql -u root -p$dbPassword $dbName"
} else {
    Write-Host "Using local MariaDB at $dbHost" -ForegroundColor Yellow
    
    if (-not $dbPassword) {
        $securePassword = Read-Host "Enter database password for user '$dbUser'" -AsSecureString
        $dbPassword = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
            [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
        )
    }
    
    # Run locally using mysql client
    $env:MYSQL_PWD = $dbPassword
    try {
        $process = Start-Process -FilePath "mysql" `
            -ArgumentList "-h", $dbHost, "-u", $dbUser, $dbName `
            -RedirectStandardInput ".\infrastructure\check-chronology-data.sql" `
            -NoNewWindow -Wait -PassThru
        
        if ($process.ExitCode -ne 0) {
            Write-Host "Error running diagnostic query (exit code: $($process.ExitCode))" -ForegroundColor Red
            exit 1
        }
    } finally {
        Remove-Item Env:\MYSQL_PWD -ErrorAction SilentlyContinue
    }
}

Write-Host "`nDiagnostic complete!" -ForegroundColor Green
Write-Host @"

Common issues:
1. Draw Pile Cards = 0: No decks were selected when creating the game
2. Cards With Years = 0: The selected deck cards don't have recognizable years (1000-2999)
3. Current Cards = 0: The game hasn't been started yet (click Start Game button)

To fix:
- Delete the test game and create a new one
- Make sure to SELECT at least one deck with PROMPT cards
- The cards need to have a 4-digit year (1000-2999) somewhere in the text
"@ -ForegroundColor Yellow
