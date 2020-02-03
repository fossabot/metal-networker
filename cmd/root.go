package cmd

import (
	"github.com/metal-stack/metal-networker/internal/netconf"

	"go.uber.org/zap"

	"github.com/metal-pod/v"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

const (
	flagInputName = "input"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "metal-networker",
	Short: "Configure network of bare metal servers",
	Long: `"metal-networker" is a self-sufficient tool to configure network related aspects of a bare metal server.
A bare metal server can be treated either as 'machine' or 'firewall'.`,
	Version: v.V.String(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(log *zap.SugaredLogger) {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.AddCommand(machineCmd)
	rootCmd.AddCommand(firewallCmd)

	machineCmd.AddCommand(machineConfigureCmd)
	firewallCmd.AddCommand(firewallConfigureCmd)

	// Here you will define your flags and configuration settings.
	rootCmd.PersistentFlags().StringP(flagInputName, "i", "", "Path to a YAML file containing network configuration")

	err := rootCmd.MarkPersistentFlagRequired(flagInputName)
	if err != nil {
		log.Warnf("error setting up cobra: %v", err)
	}
}

// initConfig reads in ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}

// configure configures bare metal server depending on kind.
func configure(kind netconf.BareMetalType, cmd *cobra.Command) error {
	z, err := zap.NewProduction()
	if err != nil {
		return err
	}

	logger := z.Sugar()
	logger.Infof("running app version: %s", v.V.String())

	input, err := cmd.Flags().GetString(flagInputName)

	if err != nil {
		return err
	}

	kb := netconf.NewKnowledgeBase(input, logger)

	err = kb.Validate(kind)
	if err != nil {
		logger.Panic(err)
	}

	netconf.NewConfigurator(kind, kb).Configure()
	logger.Info("completed. Exiting..")

	return nil
}
