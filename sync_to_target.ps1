$dirs = @(Get-ChildItem -Exclude "*.ps1")
$dirs.GetType().FullName
$remoteIP = '192.168.110.170'
$remotePath = 'C:\project\keen'

if ($null -eq $cred) {
  $cred = Get-Credential -UserName 'Administrator'
}

$target = New-PSSession -ComputerName $remoteIP -Credential $cred
Invoke-Command -ScriptBlock { Remove-Item -Path 'C:\project\keen\*' -Recurse } -Session $target
foreach ($dir in $dirs) {
  if ($dir.GetType() -eq [System.IO.DirectoryInfo]) {
    Copy-Item -Path ([System.IO.DirectoryInfo] $dir).BaseName -Destination $remotePath -Recurse -ToSession $target
  }
  elseif ($dir.GetType() -eq [System.IO.FileInfo]) {
    Copy-Item -Path ([System.IO.FileInfo] $dir).FullName -Destination $remotePath -ToSession $target 
  }
}

Remove-PSSession $target