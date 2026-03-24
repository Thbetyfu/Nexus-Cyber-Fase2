$url = "http://localhost:8080"
$payloads = @("/?q='OR'1'='1", "/?user=<script>alert(1)</script>", "/?file=../../etc/passwd", "/?cmd=rm", "/?id=00", "/?exec=whoami")
$stopTime = (Get-Date).AddSeconds(20)
while ((Get-Date) -lt $stopTime) {
    for ($i=1; $i -le 30; $i++) {
        $p = $payloads | Get-Random
        $finalUrl = "$url$p&noise=$([guid]::NewGuid())"
        Start-Job -ScriptBlock { param($u); try { iwr -Uri $u -UseBasicParsing -TimeoutSec 1 } catch {} } -ArgumentList $finalUrl
    }
    Start-Sleep -Milliseconds 100
}
"NEXUS_ARMAGEDDON_COMPLETE"
