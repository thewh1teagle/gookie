package gookie

import (
	"encoding/binary"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// some API constants
const (
	cryptprotect_ui_forbidden = 0x1
)

var (
	dllcrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32 = syscall.NewLazyDLL("Kernel32.dll")

	procEncryptData = dllcrypt32.NewProc("CryptProtectData")
	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")
)

// tDataBlob  is a structure used by Windows DPAPI Crypt32.dll::CryptProtectData(tDataBlob...)
type tDataBlob struct {
	cbData uint32
	pbData *byte
}

// newBlob creates DATA_BLOB and fills member pbData
func newBlob(d []byte) *tDataBlob {
	if len(d) == 0 {
		return &tDataBlob{}
	}
	return &tDataBlob{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

// ToByteArray creates []byte from *byte member
func (b *tDataBlob) ToByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

// encrypt calls DPAPI CryptProtectData
func encrypt(data []byte) ([]byte, error) {
	var outblob tDataBlob
	r, _, err := procEncryptData.Call(uintptr(unsafe.Pointer(newBlob(data))), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	// outblob.pbData allocated inside LSA and must be freed by us
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.ToByteArray(), nil
}

// decrypt calls Crypt32.dll::CryptUnprotectData
func decrypt(data []byte) ([]byte, error) {
	var outblob tDataBlob
	r, _, err := procDecryptData.Call(uintptr(unsafe.Pointer(newBlob(data))), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.ToByteArray(), nil
}

// convertToUTF16LittleEndianBytes , Windows is Little endian.
func convertToUTF16LittleEndianBytes(s string) []byte {
	u := utf16.Encode([]rune(s)) // encode in UTF16
	b := make([]byte, 2*len(u))
	for index, value := range u {
		binary.LittleEndian.PutUint16(b[index*2:], value) // change to LittleEndian
	}
	return b
}
