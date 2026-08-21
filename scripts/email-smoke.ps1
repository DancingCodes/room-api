param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [string]$Email = ""
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

Write-Host "Sending email code to $Email..."
Invoke-RoomApi -Method POST -Path "/api/v1/auth/email-code" -Body @{
    email = $Email
} | Out-Null
Write-Host "OK email code sent"

$emailCode = Read-Host "Enter the 6-digit email code"
if ($emailCode -eq "") {
    throw "Email code is required."
}

Write-Host "Logging in or creating user..."
$login = Invoke-RoomApi -Method POST -Path "/api/v1/auth/email-login" -Body @{
    email = $Email
    email_code = $emailCode
}
if ($login.data.token -eq "") {
    throw "Email login succeeded but token is empty."
}
Write-Host "OK email login => user_id $($login.data.user.id), nickname $($login.data.user.nickname)"
Write-Host "Email smoke test completed."