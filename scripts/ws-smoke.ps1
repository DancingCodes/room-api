param(
    [string]$BaseUrl = "http://127.0.0.1:8080",
    [string]$Account = "",
    [string]$Password = "",
    [int]$TimeoutSeconds = 10
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

function ConvertTo-WebSocketUrl {
    param([string]$HttpBaseUrl, [uint64]$RoomID)

    if ($HttpBaseUrl.StartsWith("https://")) {
        $base = "wss://" + $HttpBaseUrl.Substring(8)
    } elseif ($HttpBaseUrl.StartsWith("http://")) {
        $base = "ws://" + $HttpBaseUrl.Substring(7)
    } else {
        throw "BaseUrl must start with http:// or https://"
    }

    return "$base/api/v1/ws/rooms/$RoomID"
}

function Receive-WebSocketText {
    param(
        [System.Net.WebSockets.ClientWebSocket]$WebSocket,
        [int]$TimeoutSeconds
    )

    $buffer = New-Object byte[] 4096
    $segment = [System.ArraySegment[byte]]::new($buffer)
    $cts = [System.Threading.CancellationTokenSource]::new([TimeSpan]::FromSeconds($TimeoutSeconds))
    $chunks = New-Object System.Collections.Generic.List[byte]

    try {
        do {
            $result = $WebSocket.ReceiveAsync($segment, $cts.Token).GetAwaiter().GetResult()
            if ($result.MessageType -eq [System.Net.WebSockets.WebSocketMessageType]::Close) {
                throw "WebSocket closed before receiving text message."
            }
            if ($result.Count -gt 0) {
                for ($i = 0; $i -lt $result.Count; $i++) {
                    $chunks.Add($buffer[$i])
                }
            }
        } while (-not $result.EndOfMessage)
    } catch [System.OperationCanceledException] {
        throw "Timed out waiting for WebSocket event."
    } finally {
        $cts.Dispose()
    }

    return [System.Text.Encoding]::UTF8.GetString($chunks.ToArray())
}

if ($Account -eq "" -or $Password -eq "") {
    Write-Host "Account and password are required for WebSocket smoke test."
    Write-Host "Usage: .\scripts\ws-smoke.ps1 -Account `"your_account`" -Password `"your_password`""
    exit 0
}

Write-Host "Checking health..."
$health = Invoke-RoomApi -Method GET -Path "/health"
Write-Host "OK /health => $($health.data.status)"

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

Write-Host "Creating room..."
$roomResult = Invoke-RoomApi -Method POST -Path "/api/v1/rooms" -Token $token -Body @{
    max_members = 8
}
$roomID = $roomResult.data.room.id
if ($roomID -le 0) {
    throw "Room creation succeeded but room id is invalid."
}
Write-Host "OK create room => room_id $roomID"

$ws = [System.Net.WebSockets.ClientWebSocket]::new()
$ws.Options.SetRequestHeader("Authorization", "Bearer $token")
$wsUrl = ConvertTo-WebSocketUrl -HttpBaseUrl $BaseUrl -RoomID $roomID

try {
    Write-Host "Connecting WebSocket..."
    $connectCts = [System.Threading.CancellationTokenSource]::new([TimeSpan]::FromSeconds($TimeoutSeconds))
    try {
        $null = $ws.ConnectAsync([Uri]$wsUrl, $connectCts.Token).GetAwaiter().GetResult()
    } finally {
        $connectCts.Dispose()
    }
    if ($ws.State -ne [System.Net.WebSockets.WebSocketState]::Open) {
        throw "WebSocket state is $($ws.State), want Open."
    }
    Write-Host "OK WebSocket connected"

    Write-Host "Sending message through HTTP..."
    $messageResult = Invoke-RoomApi -Method POST -Path "/api/v1/rooms/$roomID/messages" -Token $token -Body @{
        content = "ws smoke test"
    }
    $messageID = $messageResult.data.message.id
    Write-Host "OK HTTP message => message_id $messageID"

    Write-Host "Waiting for WebSocket event..."
    $eventText = Receive-WebSocketText -WebSocket $ws -TimeoutSeconds $TimeoutSeconds
    $event = $eventText | ConvertFrom-Json
    if ($event.type -ne "message.created") {
        throw "Unexpected WebSocket event type: $($event.type)"
    }
    if ([uint64]$event.room_id -ne [uint64]$roomID) {
        throw "Unexpected WebSocket room_id: $($event.room_id), want $roomID"
    }
    if ([uint64]$event.data.message.id -ne [uint64]$messageID) {
        throw "Unexpected WebSocket message id: $($event.data.message.id), want $messageID"
    }
    Write-Host "OK WebSocket event => $($event.type)"
} finally {
    if ($ws.State -eq [System.Net.WebSockets.WebSocketState]::Open) {
        Write-Host "Closing WebSocket..."
        $closeCts = [System.Threading.CancellationTokenSource]::new([TimeSpan]::FromSeconds($TimeoutSeconds))
        try {
            $null = $ws.CloseAsync(
                [System.Net.WebSockets.WebSocketCloseStatus]::NormalClosure,
                "smoke test done",
                $closeCts.Token
            ).GetAwaiter().GetResult()
        } finally {
            $closeCts.Dispose()
        }
    }
    $ws.Dispose()
}

Start-Sleep -Milliseconds 300

Write-Host "Checking disconnect leave..."
$me = Invoke-RoomApi -Method GET -Path "/api/v1/users/me" -Token $token
if ($null -ne $me.data.user.current_room_id) {
    throw "User is still in room after WebSocket disconnect: $($me.data.user.current_room_id)"
}
Write-Host "OK disconnect leave"

Write-Host "WebSocket smoke test completed."
