package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MeshRootCertificate defines the configuration for certificate issuing
// by the mesh control plane
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MeshRootCertificate struct {
	// Object's type metadata
	metav1.TypeMeta `json:",inline"`

	// Object's metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the MeshRootCertificate config specification
	// +optional
	Spec MeshRootCertificateSpec `json:"spec,omitempty"`

	// Status of the MeshRootCertificate resource
	// +optional
	Status MeshRootCertificateStatus `json:"status,omitempty"`
}

// MeshRootCertificateSpec defines the mesh root certificate specification
type MeshRootCertificateSpec struct {
	// Provider specifies the mesh certificate provider
	Provider ProviderSpec `json:"provider"`
}

// ProviderSpec defines the certificate provider used by the mesh control plane
type ProviderSpec struct {
	// CertManager specifies the cert-manager provider configuration
	// +optional
	CertManager *CertManagerProviderSpec `json:"certManager,omitempty"`

	// Vault specifies the vault provider configuration
	// +optional
	Vault *VaultProviderSpec `json:"vault,omitempty"`

	// Tresor specifies the Tresor provider configuration
	// +optional
	Tresor *TresorProviderSpec `json:"tresor,omitempty"`
}

// CertManagerProviderSpec defines the configuration of the cert-manager provider
type CertManagerProviderSpec struct {
	// SecretName specifies the name of the k8s secret containing the root certificate
	SecretName string `json:"secretName"`

	// IssuerName specifies the name of the Issuer resource
	IssuerName string `json:"issuerName"`

	// IssuerKind specifies the kind of Issuer
	IssuerKind string `json:"issuerKind"`

	// IssuerGroup specifies the group the Issuer belongs to
	IssuerGroup string `json:"issuerGroup"`
}

// VaultProviderSpec defines the configuration of the Vault provider
type VaultProviderSpec struct {
	// Host specifies the name of the Vault server
	Host string `json:"host"`

	// Role specifies the name of the role for use by mesh control plane
	Role string `json:"role"`

	// Protocol specifies the protocol for connections to Vault
	Protocol string `json:"protocol"`

	// Token specifies the name of the token to be used by mesh control plane
	// to connect to Vault
	Token string `json:"token"`
}

// TresorProviderSpec defines the configuration of the Tresor provider
type TresorProviderSpec struct {
	// SecretName specifies the name of the secret storing the root certificate
	SecretName string `json:"secretName"`
}

// MeshRootCertificateStatus defines the status of the MeshRootCertificate resource
type MeshRootCertificateStatus struct {
	// State specifies the state of the root certificate rotation
	State string `json:"state"`

	// RotationStage specifies the stage of the rotation indicating how a
	// root certificate is currently being used within the mesh. The exact
	// meaning of the RotationStage status is determined by the accompanying
	// State status
	RotationStage string `json:"rotationStage"`
}

// MeshRootCertificateList defines the list of MeshRootCertificate objects
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MeshRootCertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MeshRootCertificate `json:"items"`
}
