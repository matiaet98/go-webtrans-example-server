CERTS_DIR := ./certs

gen.cert:
	mkdir -p $(CERTS_DIR); \
	openssl req -newkey rsa:2048 -nodes -keyout $(CERTS_DIR)/certificate.key -x509 -out certificate.pem -subj '/CN=Test Certificate' -addext "subjectAltName = DNS:localhost";