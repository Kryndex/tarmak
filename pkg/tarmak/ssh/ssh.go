package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

var _ interfaces.SSH = &SSH{}

type SSH struct {
	tarmak interfaces.Tarmak
	log    *logrus.Entry
}

func New(tarmak interfaces.Tarmak) *SSH {
	s := &SSH{
		tarmak: tarmak,
		log:    tarmak.Log(),
	}

	return s
}

func (s *SSH) WriteConfig() error {
	ctx := s.tarmak.Context()

	hosts, err := ctx.Environment().Provider().ListHosts()
	if err != nil {
		return err
	}

	var sshConfig bytes.Buffer
	sshConfig.WriteString(fmt.Sprintf("# ssh config for tarmak context %s\n", ctx.ContextName()))

	for _, host := range hosts {
		_, err = sshConfig.WriteString(host.SSHConfig())
		if err != nil {
			return err
		}
	}

	err = utils.EnsureDirectory(filepath.Dir(ctx.SSHConfigPath()), 0700)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(ctx.SSHConfigPath(), sshConfig.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (s *SSH) args() []string {
	return []string{
		"ssh",
		"-F",
		s.tarmak.Context().SSHConfigPath(),
	}
}

// Pass through a local CLI session
func (s *SSH) PassThrough(argsAdditional []string) {
	args := append(s.args(), argsAdditional...)

	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		s.log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		s.log.Fatal(err)
	}
}
