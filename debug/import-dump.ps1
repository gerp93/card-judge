<#
Import a SQL dump into local MariaDB or into a temporary Docker MariaDB container.

Usage examples:
  # Default: pick the newest .sql in infrastructure/backups
  powershell -ExecutionPolicy Bypass -File .\infrastructure\import-dump.ps1

  # Specify a particular dump file
  powershell -ExecutionPolicy Bypass -File .\infrastructure\import-dump.ps1 -DumpPath ".\infrastructure\backups\20251121233133_backup_card_judge.sql"

  # Use Docker instead of local mariadb
  powershell -ExecutionPolicy Bypass -File .\infrastructure\import-dump.ps1 -UseDocker

Notes:
- For local import the script will invoke the Windows cmd shell to preserve the standard input redirection so MySQL can prompt for a password.
- The script does not store any passwords; you will be prompted.
- Run PowerShell as Administrator only if you need to manage services; the import itself does not require elevation.
#>

[CmdletBinding()]
param(
    [string]$DumpPath = '',
    [switch]$UseDocker
)

function Write-ErrorAndExit([string]$msg) {
    Write-Error $msg
    exit 1
}

# Resolve repository root (script is inside infrastructure folder)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$repoRoot = Resolve-Path -Path (Join-Path $scriptDir "..")

if ([string]::IsNullOrWhiteSpace($DumpPath)) {
    # find newest .sql in backups
    $backupDir = Join-Path $scriptDir "backups"
    if (-not (Test-Path $backupDir)) { Write-ErrorAndExit "Backups folder not found: $backupDir" }
    $candidate = Get-ChildItem -Path $backupDir -Filter "*.sql" -File | Sort-Object LastWriteTime | Select-Object -Last 1
    if (-not $candidate) { Write-ErrorAndExit "No .sql files found in $backupDir" }
    $DumpPath = $candidate.FullName
}

# Resolve full path
try { $DumpPath = Resolve-Path -Path $DumpPath -ErrorAction Stop | Select-Object -ExpandProperty Path } catch { Write-ErrorAndExit "Dump file not found: $DumpPath" }

Write-Host "Using dump: $DumpPath"

if ($UseDocker) {
    # Ensure docker available
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) { Write-ErrorAndExit "Docker not found on PATH. Install Docker Desktop or Docker Engine to use -UseDocker." }

    $containerName = "card-judge-import-$(Get-Random)"
    $rootPass = Read-Host -Prompt "Choose a temporary root password for the Docker MariaDB instance" -AsSecureString
    $rootPassPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($rootPass))

    Write-Host "Starting temporary MariaDB container named $containerName..."
    # Pass the root password safely as a single argument
    $dockerRunEnv = "MYSQL_ROOT_PASSWORD=$rootPassPlain"
    & docker run --name $containerName -e $dockerRunEnv -e MYSQL_DATABASE=card_judge -d -p 3307:3306 mariadb:10.6 | Out-Null

    Write-Host "Waiting for the database to accept connections (this can take 10-20s)..."
    Start-Sleep -Seconds 12

    Write-Host "Copying dump into the container..."
    # Use braced variable expansion to avoid PowerShell parsing the ':' after the variable
    & docker cp $DumpPath "${containerName}:/tmp/restore.sql"

    Write-Host "Importing dump into container (you will see progress or errors)..."
    # Build the command that runs inside the container; use the plain password we generated
    $innerCmd = "mysql -u root -p'$rootPassPlain' card_judge < /tmp/restore.sql"
    # Run the import inside the container
    $execArgs = @('exec','-i',$containerName,'sh','-c',$innerCmd)
    & docker @execArgs 2>&1 | Write-Host

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Docker import command returned exit code $LASTEXITCODE"
        Write-Host "Container logs:"
        docker logs $containerName | Select-Object -Last 200 | ForEach-Object { Write-Host $_ }
        Write-Host "You can remove the container with: docker rm -f $containerName"
        exit 1
    }

    Write-Host "Import completed into Docker container $containerName. Connect with:"
    Write-Host "  mysql -u root -p -h 127.0.0.1 -P 3307 card_judge"
    Write-Host "When finished, remove the container: docker rm -f $containerName"
    exit 0
}

# Local import path
# Ensure mysql client exists
if (-not (Get-Command mysql -ErrorAction SilentlyContinue)) {
    Write-ErrorAndExit "mysql client not found on PATH. Install MySQL/MariaDB client or add its bin folder to PATH." 
}

# Ensure target DB exists (prompt for root if needed)
Write-Host "Ensuring database 'card_judge' exists..."
# We will call mysql to create DB if needed. Use cmd.exe to preserve < redirection later.
$checkCreateCmd = "mysql -u root -p -h 127.0.0.1 -e `"CREATE DATABASE IF NOT EXISTS card_judge;`""
Write-Host "You will be prompted for the root password to create the DB (if required)."
$process = Start-Process -FilePath cmd.exe -ArgumentList '/c', $checkCreateCmd -NoNewWindow -Wait -PassThru
if ($process.ExitCode -ne 0) { Write-ErrorAndExit "Failed to ensure database existence (exit code $($process.ExitCode))." }

# Local import using PowerShell-native process with temporary env var for password
Write-Host "Starting local import (you will be prompted for the MySQL root password here)..."

# Prompt securely for root password and set it only for this process
$rootPass = Read-Host -Prompt "Enter MySQL root password (will not be echoed)" -AsSecureString
$rootPassPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($rootPass))
$env:MYSQL_PWD = $rootPassPlain

try {
    $mysqlExe = (Get-Command mysql -ErrorAction Stop).Source
} catch {
    Write-ErrorAndExit "mysql client not found on PATH."
}

# Build argument list for mysql (no -p so it reads password from MYSQL_PWD)
$args = @('-u','root','-h','127.0.0.1','card_judge')

Write-Host "Importing $DumpPath -> card_judge using mysql client..."

# Start mysql with stdin redirected from the dump file
$proc = Start-Process -FilePath $mysqlExe -ArgumentList $args -RedirectStandardInput $DumpPath -NoNewWindow -Wait -PassThru

# Clear sensitive env var
Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue

if ($proc.ExitCode -ne 0) {
    Write-Error "Import failed with exit code $($proc.ExitCode)."
    Write-Host "Try running the command manually to see interactive errors:" 
    Write-Host "    mysql -u root -h 127.0.0.1 card_judge < <dump-file>"
    exit $proc.ExitCode
}

Write-Host "Import finished successfully. You can verify with:"
Write-Host '  mysql -u root -p -h 127.0.0.1 -e "USE card_judge; SHOW TABLES LIKE ''CHRONOLOGY_%'';"'
exit 0
