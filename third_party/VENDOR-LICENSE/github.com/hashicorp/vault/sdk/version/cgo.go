//go:build cgo
// +build cgo

package version

func init() {
	CgoEnabled = true
}