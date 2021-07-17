package system

import (
	"fmt"

	"github.com/layer5io/meshery/mesheryctl/internal/cli/root/config"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ctx string
var viewAllTokens bool
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage Meshery user tokens",
	Long: `
	Manipulate user tokens and their context assignments in your meshconfig`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if ok := utils.IsValidSubcommand(availableSubcommands, args[0]); !ok {
			return errors.New(utils.SystemError(fmt.Sprintf("invalid command: \"%s\"", args[0])))
		}
		return nil
	},
}

var createTokenCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a token in your meshconfig",
	Long:  "Create the token with provided token name (optionally token path) to your meshconfig tokens.",
	Example: `
	mesheryctl system token add <token-name> -f <token-path>
	mesheryctl system token add <token-name> (default path is auth.json)
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenName := args[0]
		if tokenPath == "" {
			tokenPath = "auth.json"
		}

		token := config.Token{
			Name:     tokenName,
			Location: tokenPath,
		}
		if err := config.AddTokenToConfig(token, utils.DefaultConfigPath); err != nil {
			return errors.Wrap(err, "Could not create specified token to config")
		}
		log.Printf("Token %s created.", tokenName)
		return nil
	},
}
var deleteTokenCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a token from your meshconfig",
	Long:  "Delete the token with provided token name from your meshconfig tokens.",
	Example: `
	mesheryctl system token delete <token-name>
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenName := args[0]

		token := config.Token{
			Name:     tokenName,
			Location: tokenPath,
		}
		mctlCfg, err := config.ReadConfig(utils.DefaultConfigPath)
		if err != nil {
			return err
		}
		if mctlCfg, err = config.DeleteTokenFromConfig(token, mctlCfg); err != nil {
			return errors.Wrapf(err, "Could not delete token \"%s\" from config", tokenName)
		}
		err = config.WriteConfig(mctlCfg)
		if err != nil {
			return err
		}
		log.Printf("Token %s deleted.", tokenName)
		return nil
	},
}
var setTokenCmd = &cobra.Command{
	Use:   "set",
	Short: "Set token for context",
	Long:  "Set token for current context or context specified with --context flag.",
	Example: `
	mesheryctl system token set <token-name> 

	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenName := args[0]
		if ctx == "" {
			ctx = viper.GetString("current-context")

		}
		mctlCfg, err := config.ReadConfig(utils.DefaultConfigPath)
		if err != nil {
			return err
		}
		if mctlCfg, err = config.SetTokenToConfig(tokenName, mctlCfg, ctx); err != nil {
			return errors.Wrapf(err, "Could not set token \"%s\" on context %s", tokenName, ctx)

		}
		config.WriteConfig(mctlCfg)
		if err != nil {
			return err
		}
		log.Printf("Token %s set for context %s", tokenName, ctx)
		return nil
	},
}
var listTokenCmd = &cobra.Command{
	Use:   "list",
	Short: "List tokens",
	Long:  "List all the tokens in meshery config",
	Example: `
	mesheryctl system token list
	`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		mctlCfg, err := config.ReadConfig(utils.DefaultConfigPath)
		if err != nil {
			return err
		}
		log.Print("Available tokens: ")
		for _, t := range mctlCfg.Tokens {
			log.Info(t.Name)
		}
		return nil
	},
}
var viewTokenCmd = &cobra.Command{
	Use:   "view",
	Short: "View token",
	Long:  "View a specific token in meshery config",
	Example: `
	mesheryctl system token view <token-name>
	mesheryctl system token view (show token of current context)
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mctlCfg, err := config.ReadConfig(utils.DefaultConfigPath)
		if err != nil {
			return err
		}
		if viewAllTokens {
			log.Info("Listing all available tokens...\n")
			for _, t := range mctlCfg.Tokens {
				log.Info("-> token: ", t.Name)
				log.Info("   location: ", t.Location)
			}
			return nil
		}
		tokenName := ""
		if len(args) == 0 {
			token, err := mctlCfg.GetTokenForContext(viper.GetString("current-context"))
			if err != nil {
				return errors.Wrap(err, "Could not get token for the current context")
			}
			log.Warnf("Token unspecified. Displaying token for current context \"%s\"\n", viper.GetString("current-context"))
			log.Info("token: ", token.Name)
			log.Info("location: ", token.Location)
			return nil
		}
		tokenName = args[0]

		for _, t := range mctlCfg.Tokens {
			if t.Name == tokenName {
				log.Info("token: ", t.Name)
				log.Info("location: ", t.Location)
				return nil
			}
		}
		return errors.Errorf("Token %s could not be found.", tokenName)
	},
}

func init() {
	tokenCmd.AddCommand(createTokenCmd, deleteTokenCmd, setTokenCmd, listTokenCmd, viewTokenCmd)
	createTokenCmd.Flags().StringVarP(&tokenPath, "filepath", "f", "", "Add the token location")
	setTokenCmd.Flags().StringVar(&ctx, "context", "", "Pass the context")
	viewTokenCmd.Flags().BoolVar(&viewAllTokens, "all", false, "set the flag to view all the tokens.")
}
