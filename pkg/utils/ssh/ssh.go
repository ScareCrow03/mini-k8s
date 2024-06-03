package ssh

import (
	"mini-k8s/pkg/constant"

	"github.com/melbahja/goph"
)

// SSH struct
type SSH struct {
	Username string
	Password string
	Client   *goph.Client
}

// NewSSH creates a new SSH client
func NewSSH(username, password string) *SSH {
	cli, err := goph.NewUnknown(username, constant.TARGET_URL, goph.Password(password))
	if err != nil {
		panic(err)
	}
	return &SSH{
		Username: username,
		Password: password,
		Client:   cli,
	}
}

// RunCommand runs a command on the SSH client
func (s *SSH) RunCommand(command string) (string, error) {
	res, err := s.Client.Run(command)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// Close closes the SSH client
func (s *SSH) Close() {
	s.Client.Close()
}

// Upload uploads a file to the SSH client
func (s *SSH) Upload(localPath, remotePath string) error {
	return s.Client.Upload(localPath, remotePath)
}

// Download downloads a file from the SSH client
func (s *SSH) Download(remotePath, localPath string) error {
	return s.Client.Download(remotePath, localPath)
}
