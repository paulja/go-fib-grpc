default:
.PHONY: default

gencert:
	@cfssl gencert \
		-initca ca-csr.json | cfssljson -bare ca
	@cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=ca-config.json \
		-profile=server \
		server-csr.json | cfssljson -bare server
	@mv *.pem ../etc/certs/
	@rm *.csr
.PHONY: gencert
