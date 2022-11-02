package ssh

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"git.eth4.dev/golibs/errors"

	"git.eth4.dev/golibs/iorw"

	"golang.org/x/crypto/ssh"
)

type (
	keysFilter struct {
		patterns []string
	}

	signers struct {
		signers map[string]ssh.Signer
	}
)

func newSigners(keysMask ...string) (*signers, error) {
	keysMap, err := iorw.ReadFiles(path.Join(os.Getenv("HOME"), ".ssh/"), &keysFilter{patterns: keysMask})
	if err != nil {
		return nil, errors.Wrap(err, "read ssh keys")
	}

	sig := &signers{
		signers: make(map[string]ssh.Signer),
	}

	for name, data := range keysMap {
		var signer ssh.Signer

		if signer, err = ssh.ParsePrivateKey(data); err != nil {
			return nil, errors.Ctx().Str("file", name).Wrap(err, "parse ssh key")
		}

		sig.signers[filepath.Base(name)] = signer
	}

	return sig, nil
}

func (s *signers) Signer(key string) ssh.Signer {
	return s.signers[key]
}

func (s *signers) PublicKey(key string) ssh.PublicKey {
	return s.signers[key].PublicKey()
}

func (s *signers) Sign(key string, rand io.Reader, data []byte) (*ssh.Signature, error) {
	sig, err := s.signers[key].Sign(rand, data)
	if err != nil {
		return nil, errors.Wrap(err, "try signing")
	}

	return sig, nil
}

func (k *keysFilter) Name() string {
	return "ssh keys filter"
}

func (k *keysFilter) Exclude(_, _ string, fi os.FileInfo) (bool, error) {
	result := true

	if fi.IsDir() {
		return true, nil
	}

	for pi := 0; pi < len(k.patterns); pi++ {
		result = result && !strings.Contains(fi.Name(), k.patterns[pi])
	}

	return result, nil
}
