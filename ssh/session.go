package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"git.corout.in/golibs/errors"
	"git.corout.in/golibs/errors/errgroup"
	"golang.org/x/crypto/ssh"

	"git.corout.in/golibs/iorw"
)

type sshSession struct {
	sess           *ssh.Session
	outBuf, errBuf io.Writer
	startTime      time.Time
}

func (cli *sshClient) newSession(client *ssh.Client, o ...Option) (Session, error) {
	sess, err := client.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "open session")
	}

	session := &sshSession{
		sess: sess,
	}

	var (
		opts   = processOptions(o...)
		color  = cli.colors.NextColor()
		prefix string
	)

	if prefix = opts.prefix; prefix == "" {
		prefix = string(client.SessionID())
	}

	var (
		stdoutPipe, stderrPipe io.Reader
	)

	if session.outBuf = opts.stdout; session.outBuf == nil {
		session.outBuf = iorw.NewPrefixedWriter(formatStdoutPrefix(color, prefix), os.Stdout)
	}

	if session.errBuf = opts.stderr; session.errBuf == nil {
		session.errBuf = iorw.NewPrefixedWriter(formatStderrPrefix(color, prefix), os.Stderr)
	}

	if stdoutPipe, err = sess.StdoutPipe(); err != nil {
		return nil, errors.Wrap(err, "get session stdoutPipe pipe")
	}

	if stderrPipe, err = sess.StderrPipe(); err != nil {
		return nil, errors.Wrap(err, "get session stderrPipe pipe")
	}

	eg := errgroup.New()

	eg.Go(func() error {
		_, outErr := io.Copy(session.outBuf, stdoutPipe)
		return outErr
	})

	eg.Go(func() error {
		_, errErr := io.Copy(session.errBuf, stderrPipe)
		return errErr
	})

	go func() {
		if err = eg.Wait(); err != nil {
			if list := errors.AsChain(err); len(list) != 0 {
				for i := 0; i < len(list); i++ {
					session.logErr(list[i])
				}
			}
		}
	}()

	go func() {
		if err = sess.Wait(); err != nil {
			session.logErr(err)
		}
	}()

	return session, nil
}

func (ses *sshSession) Command(_ context.Context, cmd string, args ...any) error {
	if err := ses.sess.Run(fmt.Sprintf(cmd, args...)); err != nil {
		return errors.Wrap(err, "run command")
	}

	return nil
}

func (ses *sshSession) Shell(_ context.Context) error {
	// todo:

	if err := ses.sess.Shell(); err != nil {
		return errors.Wrap(err, "start shell")
	}

	return nil
}

func (ses *sshSession) Close() error {
	if err := ses.sess.Close(); err != nil {
		return errors.Wrap(err, "close session")
	}

	return nil
}

func (ses *sshSession) logErr(err error) {
	_, _ = fmt.Fprint(ses.errBuf, "\x1b[91mERROR:\x1b[0m "+errors.Formatted(err).Error())
}

func formatStdoutPrefix(color, name string) string {
	return fmt.Sprintf("\x1b[32m[o]\x1b[%s[%s]\x1b[0m ", color, name)
}

func formatStderrPrefix(color, name string) string {
	return fmt.Sprintf("\x1b[91m[e]\x1b[%s[%s]\x1b[0m ", color, name)
}
