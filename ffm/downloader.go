package ffm

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/golang-jwt/jwt/v4"
)

type Downloader interface {
	GenerateLink(ctx context.Context, path string) (string, error)
	ValidateLink(ctx context.Context, tokenString string) (string, error)
	DownloadProxy(ctx context.Context, path string) (*httputil.ReverseProxy, *http.Request, error)
}

func (f *FFM) secret() string {
	return f.addr + f.service
}

func (f *FFM) GenerateLink(ctx context.Context, path string) (string, error) {
	info, err := f.Info(ctx, path)
	if err != nil {
		return "", err
	}

	claims := tokenverify.BuildLinkClaims(info.Path, 15*time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(f.secret()))
	if err != nil {
		return "", ErrGenerateLink.WithError(err)
	}

	return tokenString, nil
}

func (f *FFM) ValidateLink(ctx context.Context, tokenString string) (string, error) {
	claims, err := tokenverify.GetLinkClaimFromToken(tokenString, tokenverify.Secret(f.secret()))
	if err != nil {
		return "", ErrCheckLink.WithError(err)
	}

	path := claims.Link

	info, err := f.Info(ctx, path)
	if err != nil {
		return "", err
	}

	if info.IsDir || float64(info.Size)/(1<<20) > 100 {
		return f.GenCompressDownloadPath(*info), nil
	}

	return f.GenDownloadPath(*info), nil
}

func (f *FFM) DownloadProxy(ctx context.Context, path string) (*httputil.ReverseProxy, *http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	targetURL, err := url.Parse(path)
	if err != nil {
		return nil, nil, err
	}

	uri, _ := url.Parse(targetURL.Scheme + "://" + targetURL.Host)
	proxy := httputil.NewSingleHostReverseProxy(uri)

	proxy.FlushInterval = 60 * time.Minute

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			rw.WriteHeader(http.StatusGatewayTimeout)
			_, _ = rw.Write([]byte(err.Error()))
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte(err.Error()))
		}
	}

	return proxy, req, nil
}
