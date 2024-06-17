[CmdletBinding()]
param (
    [Parameter(Mandatory=$true)]
    [string]$ComputerName
)

$ErrorActionPreference = "Stop"