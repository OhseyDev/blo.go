package https

import (
	"net/http"

	"github.com/OhseyDev/gospirit/configuration"
	"github.com/OhseyDev/gospirit/filenames"
)

func StartServer(addr string, handler http.Handler) error {
	if configuration.Config.UseLetsEncrypt {
		server := buildLetsEncryptServer(addr, handler)
		return server.ListenAndServeTLS("", "")
	} else {
		checkCertificates()
		return http.ListenAndServeTLS(addr, filenames.HttpsCertFilename, filenames.HttpsKeyFilename, handler)
	}
}
