package settings

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"

	"billionmail-core/api/settings/v1"
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) SetSSLConfig(ctx context.Context, req *v1.SetSSLConfigReq) (res *v1.SetSSLConfigRes, err error) {
	res = &v1.SetSSLConfigRes{}

	certPath := public.AbsPath(filepath.Join(consts.SSL_PATH, "cert.pem"))
	keyPath := public.AbsPath(filepath.Join(consts.SSL_PATH, "key.pem"))

	// If the certificate data is empty, delete the certificate file
	if req.CertData == "" && req.KeyData == "" {
		// Delete the certificate file (if it exists)
		if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
			res.Code = 500
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to delete certificate file: {}", err)))
			return res, nil
		}
		// Delete the key file (if it exists)
		if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
			res.Code = 500
			res.SetError(gerror.New(public.LangCtx(ctx, "Failed to delete key file: {}", err)))
			return res, nil
		}

		res.SetSuccess(public.LangCtx(ctx, "Certificate deleted successfully"))
		return res, nil
	}

	// Verify certificate data
	block, _ := pem.Decode([]byte(req.CertData))
	if block == nil {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid certificate data: {}", err)))
		return res, nil
	}

	// Verify certificate format
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid certificate format: {}", err)))
		return res, nil
	}

	// Verify private key data
	block, _ = pem.Decode([]byte(req.KeyData))
	if block == nil {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid private key data: {}", err)))
		return res, nil
	}

	// Ensure the certificate directory exists
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to create certificate directory: {}", err)))
		return res, nil
	}

	// Write the new certificate file
	if err := os.WriteFile(certPath, []byte(req.CertData), 0644); err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to save certificate file: {}", err)))
		return res, nil
	}

	// Write the new private key file
	if err := os.WriteFile(keyPath, []byte(req.KeyData), 0600); err != nil {
		// If the private key writing fails, the already written certificate file needs to be deleted
		_ = os.Remove(certPath)
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to save private key file: {}", err)))
		return res, nil
	}
	// Restart the container
	err = public.DockerApiFromCtx(ctx).RestartContainerByName(ctx, "billionmail-core-billionmail-1")
	if err != nil {
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to restart container: {}", err)))
		return res, nil
	}

	res.SetSuccess(public.LangCtx(ctx, "SSL certificate updated successfully"))
	return res, nil
}
