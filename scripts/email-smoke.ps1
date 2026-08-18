param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [string]$Email = "",
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

if ($Email -eq "") {
    Write-Host "Email is required."
    Write-Host "Usage: .\scripts\email-smoke.ps1 -Email `"you@example.com`""
    exit 0
}

if ($Nickname -eq "") {
    $localPart = $Email.Split('@')[0]
    if ($localPart.Length -gt 7) {
        $localPart = $localPart.Substring(0, 7)
    }
    $Nickname = "R$localPart"
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

$registerCode = Read-Host "Enter the 6-digit register email code"
if ($registerCode -eq "") {
    throw "Register email code is required."
}

Write-Host "Registering user..."
$register = Invoke-RoomApi -Method POST -Path "/api/v1/auth/register" -Body @{
    email = $Email
    email_code = $registerCode
    nickname = $Nickname
    avatar_url = $AvatarUrl
}
Write-Host "OK register => user_id $($register.data.user.id)"

Write-Host "Sending login email code to $Email..."
Invoke-RoomApi -Method POST -Path "/api/v1/auth/login-code" -Body @{
    email = $Email
} | Out-Null
Write-Host "OK login code sent"

$loginCode = Read-Host "Enter the 6-digit login email code"
if ($loginCode -eq "") {
    throw "Login email code is required."
}

Write-Host "Logging in..."
$login = Invoke-RoomApi -Method POST -Path "/api/v1/auth/login" -Body @{
    email = $Email
    email_code = $loginCode
}
if ($login.data.token -eq "") {
    throw "Login succeeded but token is empty."
}
Write-Host "OK login => user_id $($login.data.user.id)"

Write-Host "Email smoke test completed."