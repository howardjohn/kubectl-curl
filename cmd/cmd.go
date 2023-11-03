package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/proxy"
)

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: 'kubectl curl URL [flags]'")
	}
	url := args[0]

	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	filter := &proxy.FilterServer{
		AcceptHosts: proxy.MakeRegexpArrayOrDie("kubectl-curl"),
		AcceptPaths: proxy.MakeRegexpArrayOrDie("^.*"),
	}
	server, err := proxy.NewServer("", "/", "", filter, config, 0, false)
	if err != nil {
		return err
	}
	f, err := os.MkdirTemp("", "kubectl-curl-*")
	if err != nil {
		return err
	}
	fp := filepath.Join(f, "sock")
	defer os.Remove(fp)
	defer os.Remove(f)
	l, err := server.ListenUnix(fp)
	if err != nil {
		return err
	}
	go func() {
		server.ServeOnListener(l)
	}()
	curlArgs := []string{"--unix-socket", fp, fmt.Sprintf("http:/kubectl-curl%s", url)}
	curlArgs = append(curlArgs, args[1:]...)

	c := exec.Command("curl", curlArgs...)
	c.Stdout = os.Stdout
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return err
	}
	_ = l.Close()
	return nil
}

func Execute() {
	err := run(os.Args[1:])
	var te *exec.ExitError
	if errors.As(err, &te) {
		os.Exit(te.ExitCode())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
