package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"caddy-proxy-manager/internal/caddy"
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
)

func GenerateSelfSignedCert(domain string, uploadDir string) (certPath, keyPath string, err error) {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Caddy Proxy Manager Self-Signed"},
			CommonName:   domain,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	certPath = filepath.Join(uploadDir, domain+"-selfsigned.crt")
	keyPath = filepath.Join(uploadDir, domain+"-selfsigned.key")
	certOut, err := os.Create(certPath)
	if err != nil {
		return "", "", err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return "", "", err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certPath, keyPath, nil
}

func GenerateACMECert(domain, provider, email string, useDNS bool, dnsProvider, dnsToken string) (string, string, error) {
	tempHost := models.Host{
		Domain:          domain,
		Target:          "127.0.0.1",
		Type:            "redirect",
		SSL:             true,
		SSLProvider:     provider,
		UseDNSChallenge: useDNS,
		DNSProvider:     dnsProvider,
		DNSToken:        dnsToken,
		IsActive:        true,
	}

	if err := db.DB.Create(&tempHost).Error; err != nil {
		return "", "", err
	}
	if err := caddy.ApplyConfig(); err != nil {
		db.DB.Delete(&tempHost)
		return "", "", err
	}
	basePath := "/data/caddy/acme"
	var searchPaths []string
	if provider == "letsencrypt" {
		searchPaths = []string{
			filepath.Join(basePath, "acme-v02.api.letsencrypt.org-directory", domain, domain+".crt"),
		}
	} else if provider == "zerossl" {
		searchPaths = []string{
			filepath.Join(basePath, "acme.zerossl.com-v2-DV90", domain, domain+".crt"),
		}
	}

	foundCert := ""
	foundKey := ""
	for i := 0; i < 45; i++ {
		time.Sleep(2 * time.Second)
		for _, path := range searchPaths {
			if _, err := os.Stat(path); err == nil {
				foundCert = path
				foundKey = path[:len(path)-4] + ".key"
				break
			}
		}
		if foundCert != "" {
			break
		}
	}
	db.DB.Unscoped().Delete(&tempHost)
	caddy.ApplyConfig()

	if foundCert == "" {
		return "", "", fmt.Errorf("timeout waiting for certificate generation")
	}
	uploadDir := "/data/custom_ssl"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	destCert := filepath.Join(uploadDir, domain+"-"+provider+".crt")
	destKey := filepath.Join(uploadDir, domain+"-"+provider+".key")

	if err := copyFile(foundCert, destCert); err != nil {
		return "", "", err
	}
	if err := copyFile(foundKey, destKey); err != nil {
		return "", "", err
	}

	return destCert, destKey, nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

func GetCertExpiry(certPath string) (time.Time, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return time.Time{}, err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}

	return cert.NotAfter, nil
}
func fallbackToSelfSigned(host *models.Host) error {
	uploadDir := "/data/custom_ssl"
	certPath, keyPath, err := GenerateSelfSignedCert(host.Domain, uploadDir)
	if err != nil {
		return fmt.Errorf("failed to generate self-signed cert: %w", err)
	}

	expiry, _ := GetCertExpiry(certPath)

	cert := models.Certificate{
		Domain:    host.Domain,
		CertFile:  certPath,
		KeyFile:   keyPath,
		ExpiresAt: expiry,
		CreatedAt: time.Now(),
	}
	if err := db.DB.Create(&cert).Error; err != nil {
		return fmt.Errorf("failed to save cert record: %w", err)
	}

	host.CertificateID = &cert.ID
	return nil
}
