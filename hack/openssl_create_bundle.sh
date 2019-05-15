openssl pkcs12 -export -clcerts \
	-inkey $1 \
	-in $2 \
	-out client.p12 \
	-passout pass:gslc \
	-name "Key pair gslc"

certutil -A -n "CA gslc" -t "TC,," -d sql:$HOME/.pki/nssdb -i ca.crt
pk12util -i path/to/bundle.p12 -d sql:$HOME/.pki/nssdb -W gslc
