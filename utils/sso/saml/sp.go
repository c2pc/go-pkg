package saml

import (
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	dsig "github.com/russellhaering/goxmldsig"
)

func newSamlSP(opts samlsp.Options) (*Middleware, error) {
	m := &Middleware{
		ServiceProvider: defaultServiceProvider(opts),
		Binding:         "",
		ResponseBinding: saml.HTTPPostBinding,
		OnError:         samlsp.DefaultOnError,
		Session:         samlsp.DefaultSessionProvider(opts),
	}
	m.RequestTracker = samlsp.DefaultRequestTracker(opts, &m.ServiceProvider)
	if opts.UseArtifactResponse {
		m.ResponseBinding = saml.HTTPArtifactBinding
	}

	return m, nil
}

func defaultServiceProvider(opts samlsp.Options) saml.ServiceProvider {
	host := opts.URL
	host.Path = ""

	metadataURL := host.ResolveReference(&url.URL{Path: "saml/metadata"})
	acsURL := host.ResolveReference(&url.URL{Path: "saml/acs"})
	sloURL := host.ResolveReference(&url.URL{Path: "saml/slo"})

	var forceAuthn *bool
	if opts.ForceAuthn {
		forceAuthn = &opts.ForceAuthn
	}
	signatureMethod := dsig.RSASHA1SignatureMethod
	if !opts.SignRequest {
		signatureMethod = ""
	}

	if opts.DefaultRedirectURI == "" {
		opts.DefaultRedirectURI = "/"
	}

	if len(opts.LogoutBindings) == 0 {
		opts.LogoutBindings = []string{saml.HTTPPostBinding}
	}

	return saml.ServiceProvider{
		EntityID:              opts.EntityID,
		Key:                   opts.Key,
		Certificate:           opts.Certificate,
		HTTPClient:            opts.HTTPClient,
		Intermediates:         opts.Intermediates,
		MetadataURL:           *metadataURL,
		AcsURL:                *acsURL,
		SloURL:                *sloURL,
		IDPMetadata:           opts.IDPMetadata,
		ForceAuthn:            forceAuthn,
		RequestedAuthnContext: opts.RequestedAuthnContext,
		SignatureMethod:       signatureMethod,
		AllowIDPInitiated:     opts.AllowIDPInitiated,
		DefaultRedirectURI:    opts.DefaultRedirectURI,
		LogoutBindings:        opts.LogoutBindings,
	}
}
