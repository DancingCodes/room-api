param(
    [string]$BaseUrl = "http://127.0.0.1:9999",
    [string]$Account = "",
    [string]$Email = "",
    [string]$Password = "123456",
    [string]$Nickname = "",
    [string]$AvatarUrl = "https://example.com/avatar.png"
)

$ErrorActionPreference = "Stop"

function ConvertTo-BodyJson {
    param([hashtable]$Body)

    return $Body | ConvertTo-Json -Depth 10 -Compress
}

function Invoke-RoomApi {
    param(
        [string]$Method,
        [string]$Path,
        [hashtable]$Body = $null
    )

    $params = @{
        Method = $Method
        Uri = "$BaseUrl$Path"
    }

    if ($null -ne $Body) {
        $params["ContentType"] = "application/json"
        $params["Body"] = ConvertTo-BodyJson $Body
    }

    $response = Invoke-RestMethod @params
    if ($response.code -ne 200) {
        $json = $response | ConvertTo-Json -Depth 10
        throw "$Method $Path failed: $json"
    }

    return $response
}

if ($Account -eq "" -or $Email -eq "") {
    Write-Host "Account and email are required."
    Write-Host "Usage: .\scripts\email-smoke.ps1 -Account `"roomtest01`" -Email `"you@example.com`""
    exit 0
}

if ($Nickname -eq "") {
    $suffix = $Account
    if ($suffix.Length -gt 5) {
        $suffix = $suffix.Substring($suffix.Length - 5)
    }
    $Nickname = "R$suffix"
    if ($Nickname.Length -gt 8) {
        $Nickname = $Nickname.Substring(0, 8)
    }
}

Write-Host "Checking health..."
$health = Invoke-RoomApi -Method GET -Path "/health"
Write-Host "OK /health => $($health.data.status)"

Write-Host "Sending register email code to $Email..."
Invoke-RoomApi -Method POST -Path "/api/v1/auth/register-code" -Body @{
    email = $Email
} | Out-Null
Write-Host "OK register code sent"

$code = Read-Host "Enter the 6-digit email code"
if ($code -eq "") {
    throw "Email code is required."
}

Write-Host "Registering user..."
$register = Invoke-RoomApi -Method POST -Path "/api/v1/auth/register" -Body @{
    account = $Account
    email = $Email
    email_code = $code
    password = $Password
    nickname = $Nickname
    avatar_url = $AvatarUrl
}
Write-Host "OK register => user_id $($register.data.user.id)"

Write-Host "Logging in..."
$login = Invoke-RoomApi -Method POST -Path "/api/v1/auth/login" -Body @{
    account = $Account
    password = $Password
}
if ($login.data.token -eq "") {
    throw "Login succeeded but token is empty."
}
Write-Host "OK login => user_id $($login.data.user.id)"

Write-Host "Email smoke test completed."
