$dirs = @(Get-ChildItem -Exclude '*.ps1', 'vendor')
$remoteIP = @('192.168.110.170', '192.168.110.175')
$remoteIP = @($remoteIP[1])
$remotePath = 'C:\project\keen'

foreach ($ip in $remoteIP) {
  $cred = Get-Credential -UserName 'administrator'
  $target = New-PSSession -ComputerName $ip -Credential $cred
  Invoke-Command -ScriptBlock { Get-ChildItem -Path $Using:remotePath | Where-Object { $_.Name -ne 'vendor' } | Remove-Item -Recurse } -Session $target -ArgumentList $remotePath
  foreach ($dir in $dirs) {
    if ($dir.GetType() -eq [System.IO.DirectoryInfo]) {
      Copy-Item -Path ([System.IO.DirectoryInfo] $dir).BaseName -Destination $remotePath -Recurse -ToSession $target
    }
    elseif ($dir.GetType() -eq [System.IO.FileInfo]) {
      Copy-Item -Path ([System.IO.FileInfo] $dir).FullName -Destination $remotePath -ToSession $target 
    }
  }

  Remove-PSSession $target
}
