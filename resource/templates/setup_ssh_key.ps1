param(
    [string]$Server,
    [string]$User = "root",
    [int]$Port = 22
)

$sshDir = "$env:USERPROFILE\.ssh"
$keyPath = "$sshDir\id_rsa"
$pubKey = "$sshDir\id_rsa.pub"

Write-Host "====== SSH Key Setup ======"

if (!(Test-Path $sshDir)) {
    New-Item -ItemType Directory -Path $sshDir | Out-Null
}

if (!(Test-Path $keyPath)) {
    Write-Host "No SSH key found, generating..."
    ssh-keygen -t rsa -b 4096 -f $keyPath -N ""
} else {
    Write-Host "An SSH key already exists"
}

$key = Get-Content $pubKey

Write-Host "Upload the public key to the server..."

ssh -p $Port "$User@$Server" "mkdir -p ~/.ssh && chmod 700 ~/.ssh"
ssh -p $Port "$User@$Server" "echo '$key' >> ~/.ssh/authorized_keys"
ssh -p $Port "$User@$Server" "chmod 600 ~/.ssh/authorized_keys"

Write-Host "Test password-free login..."

ssh -p $Port "$User@$Server" "echo SSH Key Login Success"

Write-Host "done"