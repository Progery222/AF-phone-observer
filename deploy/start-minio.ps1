# Локальный MinIO для скриншотов observer (без Docker)
$ErrorActionPreference = "Stop"
$Dir = $PSScriptRoot
$Exe = Join-Path $Dir "minio.exe"
# Путь без пробелов: MinIO на Windows не инициализируется в каталогах вроде "Mobile Farm"
$Data = if ($env:AF_MINIO_DATA) { $env:AF_MINIO_DATA } else { "C:\af-minio-data" }

if (-not (Test-Path $Exe)) {
    Write-Host "Скачивание minio.exe..."
    Invoke-WebRequest -Uri "https://dl.min.io/server/minio/release/windows-amd64/minio.exe" -OutFile $Exe -UseBasicParsing
}
New-Item -ItemType Directory -Force -Path $Data | Out-Null

$env:MINIO_ROOT_USER = "minioadmin"
$env:MINIO_ROOT_PASSWORD = "minioadmin"

Write-Host "MinIO API :9000, console :9001, data=$Data"
& $Exe server $Data --console-address ":9001"
