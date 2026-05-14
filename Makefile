.PHONY: certs
certs:
	mkdir -p certs
	openssl genrsa -out certs/ca.key 4096
	openssl req -new -x509 -key certs/ca.key -sha256 -subj "/C=US/ST=NJ/O=CA, Inc." -days 1024 -out certs/ca.cert
	openssl genrsa -out certs/service.key 4096
	openssl req -new -key certs/service.key -out certs/service.csr -config certificate.conf
	openssl x509 -req -in certs/service.csr -CA certs/ca.cert -CAkey certs/ca.key -CAcreateserial \
		-out certs/service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext
