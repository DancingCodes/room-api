param(
    [string]$BaseUrl = "http://127.0.0.1:9999",
    [string]$Account = "",
    [string]$Password = ""
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
        [hashtable]$Body = $null,
        [string]$Token = ""
    )

    $headers = @{}
    if ($Token -ne "") {
        $headers["Authorization"] = "Bearer $Token"
    }

    $params = @{
        Method = $Method
        Uri = "$BaseUrl$Path"
        Headers = $headers
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

Write-Host "Checking health..."
$health = Invoke-RoomApi -Method GET -Path "/health"
Write-Host "OK /health => $($health.data.status)"

if ($Account -eq "" -or $Password -eq "") {
    Write-Host "Account or password not provided. Authenticated smoke tests skipped."
    exit 0
}

Write-Host "Logging in..."
$login = Invoke-RoomApi -Method POST -Path "/api/v1/auth/login" -Body @{
    account = $Account
    password = $Password
}
$token = $login.data.token
if ($token -eq "") {
    throw "Login succeeded but token is empty."
}
Write-Host "OK login => user_id $($login.data.user.id)"

Write-Host "Checking current user..."
$me = Invoke-RoomApi -Method GET -Path "/api/v1/users/me" -Token $token
Write-Host "OK /users/me => $($me.data.user.account)"

Write-Host "Creating room..."
$roomResult = Invoke-RoomApi -Method POST -Path "/api/v1/rooms" -Token $token -Body @{
    max_members = 8
}
$roomID = $roomResult.data.room.id
if ($roomID -le 0) {
    throw "Room creation succeeded but room id is invalid."
}
Write-Host "OK create room => room_id $roomID"

Write-Host "Sending message..."
$messageResult = Invoke-RoomApi -Method POST -Path "/api/v1/rooms/$roomID/messages" -Token $token -Body @{
    content = "smoke test"
}
Write-Host "OK create message => message_id $($messageResult.data.message.id)"

Write-Host "Listing messages..."
$messages = Invoke-RoomApi -Method GET -Path "/api/v1/rooms/$roomID/messages?limit=20" -Token $token
Write-Host "OK list messages => count $($messages.data.list.Count)"

Write-Host "Leaving room..."
Invoke-RoomApi -Method POST -Path "/api/v1/rooms/$roomID/leave" -Token $token | Out-Null
Write-Host "OK leave room"

Write-Host "Smoke test completed."
