package certificate

import (
	"math/rand"
	time "time"

	"github.com/rs/zerolog/log"

	"github.com/openservicemesh/osm/pkg/certificate/pem"
	"github.com/openservicemesh/osm/pkg/errcode"
)

const (
	// RenewBeforeCertExpires signifies how much earlier (before expiration) should a certificate be renewed
	RenewBeforeCertExpires = 30 * time.Second

	// So that we do not renew all certs at the same time - add noise.
	// These define the min and max of the seconds of noise to be added
	// to the early certificate renewal.
	noiseSeconds = 5
)

// GetCommonName returns the Common Name of the certificate
func (c *Certificate) GetCommonName() CommonName {
	return c.CommonName
}

// GetSerialNumber returns the serial number of the certificate
func (c *Certificate) GetSerialNumber() SerialNumber {
	return c.SerialNumber
}

// GetExpiration returns the expiration time of the certificate
func (c *Certificate) GetExpiration() time.Time {
	return c.Expiration
}

// GetCertificateChain returns the certificate chain of the certificate
func (c *Certificate) GetCertificateChain() pem.Certificate {
	return c.CertChain
}

// GetPrivateKey returns the private key of the certificate
func (c *Certificate) GetPrivateKey() pem.PrivateKey {
	return c.PrivateKey
}

// GetIssuingCA returns the issuing CA of the certificate
func (c *Certificate) GetIssuingCA() pem.RootCertificate {
	return c.IssuingCA
}

// ShouldRotate determines whether a certificate should be rotated.
func (c *Certificate) ShouldRotate() bool {
	// The certificate is going to expire at a timestamp T
	// We want to renew earlier. How much earlier is defined in renewBeforeCertExpires.
	// We add a few seconds noise to the early renew period so that certificates that may have been
	// created at the same time are not renewed at the exact same time.

	intNoise := rand.Intn(noiseSeconds) // #nosec G404
	secondsNoise := time.Duration(intNoise) * time.Second
	return time.Until(c.GetExpiration()) <= (RenewBeforeCertExpires + secondsNoise)
}

// NewFromPEM is a helper returning a *certificate.Certificate from the PEM components given.
func NewFromPEM(pemCert pem.Certificate, pemKey pem.PrivateKey) (*Certificate, error) {
	x509Cert, err := DecodePEMCertificate(pemCert)
	if err != nil {
		// TODO(#3962): metric might not be scraped before process restart resulting from this error
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrDecodingPEMCert)).
			Msg("Error converting PEM cert to x509 to obtain serial number")
		return nil, err
	}

	return &Certificate{
		CommonName:   CommonName(x509Cert.Subject.CommonName),
		SerialNumber: SerialNumber(x509Cert.SerialNumber.String()),
		CertChain:    pemCert,
		IssuingCA:    pem.RootCertificate(pemCert),
		PrivateKey:   pemKey,
		Expiration:   x509Cert.NotAfter,
	}, nil
}
