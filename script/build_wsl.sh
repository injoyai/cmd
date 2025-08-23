target='Ubuntu-22.04'

wsl -d $target bash -ic "
  export GOPROXY=https://goproxy.io
  export https_proxy=http://192.168.10.19:1081
  export http_proxy=http://192.168.10.19:1081
  go build -v
"