key=${1:-key.key}
cert=${2:-client_cert.crt}
server=${3:-}

curl -v -k --key ${key} --cert ${cert} https://${server}/version
