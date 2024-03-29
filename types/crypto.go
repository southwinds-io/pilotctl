/*
   Pilot Control Service
   Copyright (C) 2022-Present SouthWinds Tech Ltd - www.southwinds.io

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package types

import (
	"bytes"
	"compress/gzip"
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/openpgp/s2k"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"southwinds.dev/artisan/core"
	"strconv"
	"time"
)

// Sign create a cryptographic signature for the passed-in object
func sign(obj interface{}) (string, error) {
	// only sign if we have an object
	if obj != nil {
		// load the signing key
		path, err := KeyFilePath("sign")
		if err != nil {
			return "", err
		}
		// retrieve the verification key from the specified location
		pgp, err := LoadPGP(path, "")
		if err != nil {
			return "", fmt.Errorf("sign => cannot load signing key: %s", err)
		}
		// obtain the object checksum
		cs, err := checksum(obj)
		if err != nil {
			return "", fmt.Errorf("sign => cannot create checksum: %s", err)
		}
		signature, err := pgp.Sign(cs)
		if err != nil {
			return "", fmt.Errorf("sign => cannot create signature: %s", err)
		}
		// return a base64 encoded string with the digital signature
		return base64.StdEncoding.EncodeToString(signature), nil
	}
	return "", nil
}

// checksum create a checksum of the passed-in object
func checksum(obj interface{}) ([]byte, error) {
	source, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("checksum => cannot convert object to JSON to produce checksum: %s", err)
	}
	// indent the json to make it readable
	dest := new(bytes.Buffer)
	err = json.Indent(dest, source, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("checksum => cannot indent JSON to produce checksum: %s", err)
	}
	// create a new hash
	hash := sha256.New()
	// write object bytes into hash
	_, err = hash.Write(dest.Bytes())
	if err != nil {
		return nil, fmt.Errorf("checksum => cannot write JSON bytes to hash: %s", err)
	}
	// obtain checksum
	sum := hash.Sum(nil)
	return sum, nil
}

// KeyFilePath return the path to the file where the relevant PGP key is
// keyType is either verify (public) or sign (private) PGP key
func KeyFilePath(keyType string) (string, error) {
	name := fmt.Sprintf(".pilot_%s.pgp", keyType)
	path := filepath.Join(executablePath(), name)
	_, err := os.Stat(path)
	if err != nil {
		path = filepath.Join(homePath(), name)
		_, err = os.Stat(path)
		if err != nil {
			// TODO: make path OS agnostic
			path = fmt.Sprintf("/keys/%s", name)
			_, err = os.Stat(path)
			if err != nil {
				return "", fmt.Errorf("cannot find %s key\n", keyType)
			}
		}
		return path, nil
	}
	return path, nil
}

func receiverConfigFile() string {
	filename := "ev_receive.json"
	path := filepath.Join(executablePath(), filename)
	_, err := os.Stat(path)
	if err != nil {
		path = filepath.Join(homePath(), filename)
		_, err = os.Stat(path)
		if err != nil {
			path = fmt.Sprintf("/conf/%s", filename)
			_, err = os.Stat(path)
			if err != nil {
				path, err := core.AbsPath(filename)
				if err != nil {
					return ""
				}
				return path
			}
		}
		return path
	}
	return path
}

func executablePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}

func homePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return home
}

// PGP entity for signing, verification, encryption and decryption
type PGP struct {
	entity  *openpgp.Entity
	conf    *packet.Config
	name    string
	comment string
	email   string
}

const (
	defaultDigest = crypto.SHA256
	defaultCipher = packet.CipherAES128
)

// LoadPGP load a PGP entity from file
func LoadPGP(filename, passphrase string) (*PGP, error) {
	if !filepath.IsAbs(filename) {
		abs, err := filepath.Abs(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot convert path %s to absolute path: %s", filename, err)
		}
		filename = abs
	}
	var reader io.Reader
	// if key is encrypted
	if len(passphrase) > 0 {
		c := &AesCrypto{
			CipherMode: CBC,
			Padding:    NoPadding,
		}
		bb, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		decrypted, err := c.Decrypt(string(bb), []byte(passphrase))
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader([]byte(decrypted))
	} else {
		// read the key file
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot open key file %s: %s", filename, err)
		}
		reader = f
	}
	entityList, err := openpgp.ReadArmoredKeyRing(reader)
	if err != nil {
		return nil, fmt.Errorf("cannot read PGP entity: %s", err)
	}
	if len(entityList) == 0 {
		return nil, fmt.Errorf("no PGP entities found in %s", filename)
	}
	entity := entityList[0]

	// NOTE: if this is a public key, adds the default cipher to the id self-signature so that the public key can be used to
	// encrypt messages without failing with message:
	// "cannot encrypt because no candidate hash functions are compiled in. (Wanted RIPEMD160 in this case.)
	// PGP Encrypt defaults to PreferredSymmetric=Cast5 & PreferredHash=Ripemd160
	// To avoid the error above, it has to change the required values
	// It needs to be in the list of preferred algorithms specified in the self-signature of the primary identity
	// there should only be one, but cycle over all identities for completeness
	for _, id := range entity.Identities {
		preferredHashId, _ := s2k.HashToHashId(defaultDigest)
		id.SelfSignature.PreferredHash = []uint8{preferredHashId}
		id.SelfSignature.PreferredSymmetric = []uint8{uint8(defaultCipher)}
	}

	return &PGP{
		entity: entity,
		conf: &packet.Config{
			DefaultCipher: defaultCipher,
			DefaultHash:   defaultDigest,
			RSABits:       2048,
			Time: func() time.Time {
				return time.Now()
			},
		},
	}, nil
}

func LoadPGPBytes(key []byte) (*PGP, error) {
	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(key))
	if err != nil {
		return nil, fmt.Errorf("cannot read PGP entity: %s", err)
	}
	if len(entityList) == 0 {
		return nil, fmt.Errorf("no PGP entities found in key")
	}
	entity := entityList[0]

	// NOTE: if this is a public key, adds the default cipher to the id self-signature so that the public key can be used to
	// encrypt messages without failing with message:
	// "cannot encrypt because no candidate hash functions are compiled in. (Wanted RIPEMD160 in this case.)
	// PGP Encrypt defaults to PreferredSymmetric=Cast5 & PreferredHash=Ripemd160
	// To avoid the error above, it has to change the required values
	// It needs to be in the list of preferred algorithms specified in the self-signature of the primary identity
	// there should only be one, but cycle over all identities for completeness
	for _, id := range entity.Identities {
		preferredHashId, _ := s2k.HashToHashId(defaultDigest)
		id.SelfSignature.PreferredHash = []uint8{preferredHashId}
		id.SelfSignature.PreferredSymmetric = []uint8{uint8(defaultCipher)}
	}

	return &PGP{
		entity: entity,
		conf: &packet.Config{
			DefaultCipher: defaultCipher,
			DefaultHash:   defaultDigest,
			RSABits:       2048,
			Time: func() time.Time {
				return time.Now()
			},
		},
	}, nil
}

// HasPrivate check if the PGP entity has a private key, if not an error is returned
func (p *PGP) HasPrivate() bool {
	if p.entity == nil {
		core.RaiseErr("PGP object does not contain entity")
	}
	return p.entity.PrivateKey != nil
}

// Sign signs the specified message (requires loading a private key)
func (p *PGP) Sign(message []byte) ([]byte, error) {
	writer := new(bytes.Buffer)
	reader := bytes.NewReader(message)
	err := openpgp.ArmoredDetachSign(writer, p.entity, reader, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot sign message: %s", err)
	}
	return writer.Bytes(), nil
}

// Verify verifies the message using a specified signature (requires loading a public key)
func (p *PGP) Verify(message []byte, signature []byte) error {
	sig, err := parseSignature(signature)
	if err != nil {
		return err
	}
	hash := sig.Hash.New()
	messageReader := bytes.NewReader(message)
	io.Copy(hash, messageReader)
	err = p.entity.PrimaryKey.VerifySignature(hash, sig)
	if err != nil {
		return err
	}
	return nil
}

// Encrypt encrypts the specified message
func (p *PGP) Encrypt(message []byte) ([]byte, error) {
	// create buffer to write output to
	buf := new(bytes.Buffer)
	// create armor format encoder
	encoderWriter, err := armor.Encode(buf, "Message", make(map[string]string))
	if err != nil {
		return []byte{}, fmt.Errorf("cannot create PGP armor: %v", err)
	}
	// create the encryptor with the encoder
	encryptorWriter, err := openpgp.Encrypt(encoderWriter, []*openpgp.Entity{p.entity}, nil, nil, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot create encryptor: %v", err)
	}
	// create the compressor with the encryptor
	compressorWriter, err := gzip.NewWriterLevel(encryptorWriter, gzip.BestCompression)
	if err != nil {
		return []byte{}, fmt.Errorf("invalid compression level: %v", err)
	}
	// write the message to the compressor
	messageReader := bytes.NewReader(message)
	_, err = io.Copy(compressorWriter, messageReader)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot write data to the compressor: %v", err)
	}
	compressorWriter.Close()
	encryptorWriter.Close()
	encoderWriter.Close()
	// returns an encoded, encrypted, and compressed message
	return buf.Bytes(), nil
}

// Decrypt decrypts the specified message
func (p *PGP) Decrypt(encrypted []byte) ([]byte, error) {
	// Decode message
	block, err := armor.Decode(bytes.NewReader(encrypted))
	if err != nil {
		return []byte{}, fmt.Errorf("cannot decode the PGP armor encrypted string: %v", err)
	}
	if block.Type != "Message" {
		return []byte{}, errors.New("invalid message type")
	}
	// decrypt the message
	entityList := openpgp.EntityList{
		p.entity,
	}
	messageReader, err := openpgp.ReadMessage(block.Body, entityList, nil, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot read message: %v", err)
	}
	read, err := ioutil.ReadAll(messageReader.UnverifiedBody)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot read unverified body: %v", err)
	}
	// unzip the message
	reader := bytes.NewReader(read)
	uncompressed, err := gzip.NewReader(reader)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot initialise gzip reader: %v", err)
	}
	defer uncompressed.Close()
	out, err := ioutil.ReadAll(uncompressed)
	if err != nil {
		return []byte{}, err
	}
	// return the unencoded, unencrypted, and uncompressed message
	return out, nil
}

// create PEM headers for the PGP key
func pemHeaders(version, cipher, hash string, rsaBits int, time time.Time) map[string]string {
	headers := map[string]string{
		"Version": fmt.Sprintf("golang.org/x/crypto/openpgp - %s", version),
		"Comment": fmt.Sprintf("Cipher: %s, Hash: %s, RSA Bits: %s, Created: %s", cipher, hash, strconv.Itoa(rsaBits), time.String()),
		"Hash":    fmt.Sprintf("%s/%s", cipher, hash),
	}
	return headers
}

// armor ascii encode the passed in buffer
func armorEncode(key *bytes.Buffer, keyType string, headers map[string]string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w, err := armor.Encode(buf, keyType, headers)
	if err != nil {
		return nil, fmt.Errorf("cannot encode keys in armor format: %s", err)
	}
	_, err = w.Write(key.Bytes())
	if err != nil {
		return nil, fmt.Errorf("\"error armoring serializedEntity: %s", err)
	}
	w.Close()
	return buf.Bytes(), nil
}

// returns the string representation of the passed i cipher function
func cipherToString(cipher packet.CipherFunction) string {
	switch cipher {
	case 2:
		return "3DES"
	case 3:
		return "CAST5"
	case 7:
		return "AES128"
	case 8:
		return "AES192"
	case 9:
		return "AES256"
	default:
		return "NotKnown"
	}
}

// parses a string of bytes containing a PGP signature
func parseSignature(signature []byte) (*packet.Signature, error) {
	signatureReader := bytes.NewReader(signature)
	block, err := armor.Decode(signatureReader)
	if err != nil {
		return nil, fmt.Errorf("cannot decode OpenPGP Armor: %s", err)
	}
	if block.Type != openpgp.SignatureType {
		return nil, errors.New("invalid signature file")
	}
	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		return nil, err
	}
	sig, ok := pkt.(*packet.Signature)
	if !ok {
		return nil, errors.New("cannot parse PGP signature")
	}
	return sig, nil
}
