# 获取当前目录路径
$currentDirectory = Get-Location

# 构建tmp文件夹的完整路径
$tmpDirectory = Join-Path $currentDirectory "tmp"

# 检查tmp文件夹是否存在
if (Test-Path $tmpDirectory) {
    # 删除tmp文件夹及其所有子文件夹和文件
    Remove-Item -Path $tmpDirectory -Recurse -Force
}

mkdir tmp
mkdir tmp/wallets
mkdir tmp/ref_list
mkdir tmp/blocks
mkdir tmp/utxoset