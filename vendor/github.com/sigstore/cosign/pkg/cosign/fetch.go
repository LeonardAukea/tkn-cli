//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cosign

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"knative.dev/pkg/pool"

	"github.com/sigstore/cosign/pkg/oci"
	ociremote "github.com/sigstore/cosign/pkg/oci/remote"
)

type SignedPayload struct {
	Base64Signature string
	Payload         []byte
	Cert            *x509.Certificate
	Chain           []*x509.Certificate
	Bundle          *oci.Bundle
}

type Signatures struct {
	KeyID string `json:"keyid"`
	Sig   string `json:"sig"`
}

type AttestationPayload struct {
	PayloadType string       `json:"payloadType"`
	PayLoad     string       `json:"payload"`
	Signatures  []Signatures `json:"signatures"`
}

const (
	SignatureTagSuffix   = ".sig"
	SBOMTagSuffix        = ".sbom"
	AttestationTagSuffix = ".att"
)

const (
	Signature   = "signature"
	SBOM        = "sbom"
	Attestation = "attestation"
)

func FetchSignaturesForReference(ctx context.Context, ref name.Reference, opts ...ociremote.Option) ([]SignedPayload, error) {
	simg, err := ociremote.SignedEntity(ref, opts...)
	if err != nil {
		return nil, err
	}

	sigs, err := simg.Signatures()
	if err != nil {
		return nil, errors.Wrap(err, "remote image")
	}
	l, err := sigs.Get()
	if err != nil {
		return nil, errors.Wrap(err, "fetching signatures")
	}
	if len(l) == 0 {
		return nil, fmt.Errorf("no signatures associated with %v", ref)
	}

	g := pool.New(runtime.NumCPU())
	signatures := make([]SignedPayload, len(l))
	for i, sig := range l {
		i, sig := i, sig
		g.Go(func() (err error) {
			signatures[i].Payload, err = sig.Payload()
			if err != nil {
				return err
			}
			signatures[i].Base64Signature, err = sig.Base64Signature()
			if err != nil {
				return err
			}
			signatures[i].Cert, err = sig.Cert()
			if err != nil {
				return err
			}
			signatures[i].Chain, err = sig.Chain()
			if err != nil {
				return err
			}
			signatures[i].Bundle, err = sig.Bundle()
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return signatures, nil
}

func FetchAttestationsForReference(ctx context.Context, ref name.Reference, opts ...ociremote.Option) ([]AttestationPayload, error) {
	simg, err := ociremote.SignedEntity(ref, opts...)
	if err != nil {
		return nil, err
	}

	atts, err := simg.Attestations()
	if err != nil {
		return nil, errors.Wrap(err, "remote image")
	}
	l, err := atts.Get()
	if err != nil {
		return nil, errors.Wrap(err, "fetching attestations")
	}
	if len(l) == 0 {
		return nil, fmt.Errorf("no attestations associated with %v", ref)
	}

	g := pool.New(runtime.NumCPU())
	attestations := make([]AttestationPayload, len(l))
	for i, att := range l {
		i, att := i, att
		g.Go(func() (err error) {
			attestPayload, _ := att.Payload()
			err = json.Unmarshal(attestPayload, &attestations[i])
			if err != nil {
				return err
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return attestations, nil
}