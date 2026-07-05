package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ayush00git/goth/grpc/goth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var cfgFile string

// rootCmd is the base command. All sub-commands are registered via init() in their own files.
var rootCmd = &cobra.Command{
	Use:   "gothtest",
	Short: "goth gRPC test client",
	Long:  "A CLI test client for the goth gRPC server, powered by Cobra & Viper.",
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.gothtest.yaml)")
	rootCmd.PersistentFlags().String("addr", "localhost:50051", "gRPC server address")

	_ = viper.BindPFlag("addr", rootCmd.PersistentFlags().Lookup("addr"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gothtest")
	}

	// GOTHTEST_* environment variables are automatically picked up.
	viper.SetEnvPrefix("GOTHTEST")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// grpcClient creates a gRPC connection using the configured server address
// and returns a GothServiceClient. Callers must defer conn.Close().
func grpcClient() (goth.GothServiceClient, *grpc.ClientConn, error) {
	addr := viper.GetString("addr")
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodecV2(goth.CodecV2{}),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	return goth.NewGothServiceClient(conn), conn, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
