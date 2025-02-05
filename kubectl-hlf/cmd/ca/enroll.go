package ca

import (
	log "github.com/kfsoftware/hlf-operator/internal/github.com/hyperledger/fabric-ca/sdkpatch/logbridge"
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/kfsoftware/hlf-operator/controllers/certs"
	"github.com/kfsoftware/hlf-operator/controllers/utils"
	"github.com/kfsoftware/hlf-operator/kubectl-hlf/cmd/helpers"
	"github.com/spf13/cobra"
)

type EnrollOptions struct {
	Name    string
	NS      string
	User    string
	Secret  string
	Type    string
	MspID   string
	CAName  string
	Profile string
	Hosts   []string
	CN      string
}

func (o EnrollOptions) Validate() error {
	return nil
}

type enrollCmd struct {
	out        io.Writer
	errOut     io.Writer
	enrollOpts EnrollOptions
	fileOutput string
}

func (c *enrollCmd) validate() error {
	return c.enrollOpts.Validate()
}
func (c *enrollCmd) run(args []string) error {
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	certAuth, err := helpers.GetCertAuthByName(oclient, c.enrollOpts.Name, c.enrollOpts.NS)
	if err != nil {
		return err
	}
	url, err := helpers.GetURLForCA(certAuth)
	if err != nil {
		return err
	}
	log.Debugf("CA URL=%s", url)
	crt, pk, _, err := certs.EnrollUser(certs.EnrollUserRequest{
		TLSCert:    certAuth.Status.TlsCert,
		URL:        url,
		Name:       c.enrollOpts.CAName,
		MSPID:      c.enrollOpts.MspID,
		User:       c.enrollOpts.User,
		Secret:     c.enrollOpts.Secret,
		Hosts:      c.enrollOpts.Hosts,
		CN:         c.enrollOpts.CN,
		Profile:    c.enrollOpts.Profile,
		Attributes: nil,
	})
	if err != nil {
		return err
	}
	crtPem := utils.EncodeX509Certificate(crt)
	pkPem, err := utils.EncodePrivateKey(pk)
	if err != nil {
		return err
	}
	userYaml, err := yaml.Marshal(map[string]interface{}{
		"key": map[string]interface{}{
			"pem": string(pkPem),
		},
		"cert": map[string]interface{}{
			"pem": string(crtPem),
		},
	})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(c.fileOutput, userYaml, 0644)
	if err != nil {
		return err
	}

	return nil
}
func newCAEnrollCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := enrollCmd{out: out, errOut: errOut}
	cmd := &cobra.Command{
		Use:   "enroll",
		Short: "Enroll a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(args)
		},
	}
	f := cmd.Flags()
	f.StringVar(&c.enrollOpts.Name, "name", "", "name of the Certificate Authority in the cluster, e.g ca.default")
	f.StringVarP(&c.enrollOpts.NS, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	f.StringVarP(&c.enrollOpts.CAName, "ca-name", "", "", "ca name to enroll this user")
	f.StringVarP(&c.enrollOpts.User, "user", "", "", "namespace scope for this request")
	f.StringVarP(&c.enrollOpts.Secret, "secret", "", "", "namespace scope for this request")
	f.StringVarP(&c.enrollOpts.Type, "type", "", "", "namespace scope for this request")
	f.StringVarP(&c.enrollOpts.MspID, "mspid", "", "", "namespace scope for this request")
	f.StringVarP(&c.enrollOpts.Profile, "profile", "", "", "profile")
	f.StringVarP(&c.enrollOpts.CN, "cn", "", "", "cn")
	f.StringSliceVarP(&c.enrollOpts.Hosts, "hosts", "", []string{}, "hosts")

	f.StringVar(&c.fileOutput, "output", "", "output file")

	return cmd
}
