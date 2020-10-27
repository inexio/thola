package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"thola/core/device"
	"thola/core/request"
	"thola/core/utility"
)

func init() {
	addDeviceFlags(checkIdentifyCMD)
	checkCMD.AddCommand(checkIdentifyCMD)

	checkIdentifyCMD.Flags().String("os", "", "Expected os of the device")
	checkIdentifyCMD.Flags().String("vendor", "", "Expected vendor of the device")
	checkIdentifyCMD.Flags().String("model", "", "Expected model of the device")
	checkIdentifyCMD.Flags().String("serial-number", "", "Expected serial-number of the device")
	checkIdentifyCMD.Flags().String("model-series", "", "Expected model-series of the device")
	checkIdentifyCMD.Flags().String("os-version", "", "Expected os-version of the device")

	checkIdentifyCMD.Flags().Bool("os-diff-warning", false, "Use warning level if os differs to the expected value")
	checkIdentifyCMD.Flags().Bool("vendor-diff-warning", false, "Use warning level if vendor differs to the expected value")
	checkIdentifyCMD.Flags().Bool("model-diff-warning", false, "Use warning level if model differs to the expected value")
	checkIdentifyCMD.Flags().Bool("serial-number-diff-warning", false, "Use warning level if serial-number differs to the expected value")
	checkIdentifyCMD.Flags().Bool("model-series-diff-warning", false, "Use warning level if model-series differs to the expected value")
	checkIdentifyCMD.Flags().Bool("os-version-diff-warning", false, "Use warning level if os-version differs to the expected value")

	err := viper.BindPFlag("checkIdentify.os", checkIdentifyCMD.Flags().Lookup("os"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.vendor", checkIdentifyCMD.Flags().Lookup("vendor"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.model", checkIdentifyCMD.Flags().Lookup("model"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.serial-number", checkIdentifyCMD.Flags().Lookup("serial-number"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.model-series", checkIdentifyCMD.Flags().Lookup("model-series"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.os-version", checkIdentifyCMD.Flags().Lookup("os-version"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.os-diff-warning", checkIdentifyCMD.Flags().Lookup("os-diff-warning"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.vendor-diff-warning", checkIdentifyCMD.Flags().Lookup("vendor-diff-warning"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.model-diff-warning", checkIdentifyCMD.Flags().Lookup("model-diff-warning"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.serial-number-diff-warning", checkIdentifyCMD.Flags().Lookup("serial-number-diff-warning"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.model-series-diff-warning", checkIdentifyCMD.Flags().Lookup("model-series-diff-warning"))
	if err != nil {
		os.Exit(3)
	}

	err = viper.BindPFlag("checkIdentify.os-version-diff-warning", checkIdentifyCMD.Flags().Lookup("os-version-diff-warning"))
	if err != nil {
		os.Exit(3)
	}
}

var checkIdentifyCMD = &cobra.Command{
	Use:   "identify",
	Short: "Check identify properties with given expectations",
	Long: "Check identify properties with given expectations.\n\n" +
		"You can set the expectations with the flags.",
	Run: func(cmd *cobra.Command, args []string) {
		var nilString *string
		vendor := viper.GetString("checkIdentify.vendor")
		model := viper.GetString("checkIdentify.model")
		serialNumber := viper.GetString("checkIdentify.serial-number")
		modelSeries := viper.GetString("checkIdentify.model-series")
		osVersion := viper.GetString("checkIdentify.os-version")

		r := request.CheckIdentifyRequest{
			CheckDeviceRequest: getCheckDeviceRequest(),
			Expectations: device.Device{Class: viper.GetString("checkIdentify.os"), Properties: device.Properties{
				Vendor:       utility.IfThenElse(cmd.Flags().Changed("vendor"), &vendor, nilString).(*string),
				Model:        utility.IfThenElse(cmd.Flags().Changed("model"), &model, nilString).(*string),
				ModelSeries:  utility.IfThenElse(cmd.Flags().Changed("model-series"), &modelSeries, nilString).(*string),
				SerialNumber: utility.IfThenElse(cmd.Flags().Changed("serial-number"), &serialNumber, nilString).(*string),
				OSVersion:    utility.IfThenElse(cmd.Flags().Changed("os-version"), &osVersion, nilString).(*string),
			}},
			OsDiffWarning:           viper.GetBool("checkIdentify.os-diff-warning"),
			VendorDiffWarning:       viper.GetBool("checkIdentify.vendor-diff-warning"),
			ModelDiffWarning:        viper.GetBool("checkIdentify.model-diff-warning"),
			ModelSeriesDiffWarning:  viper.GetBool("checkIdentify.model-series-diff-warning"),
			OsVersionDiffWarning:    viper.GetBool("checkIdentify.os-version-diff-warning"),
			SerialNumberDiffWarning: viper.GetBool("checkIdentify.serial-number-diff-warning"),
		}
		handleRequest(&r)
	},
}
