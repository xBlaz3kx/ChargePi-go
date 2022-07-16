package settings

type (
	/* ------------- Hardware structs ------------*/

	Hardware struct {
		Display      Display      `fig:"display" json:"display" yaml:"display" mapstructure:"display"`
		TagReader    TagReader    `fig:"tagReader" json:"tagReader" yaml:"tagReader" mapstructure:"tagReader"`
		LedIndicator LedIndicator `fig:"ledIndicator" json:"ledIndicator" yaml:"ledIndicator" mapstructure:"ledIndicator"`
	}

	EVCC struct {
		Type string
		// If Type is "Relay"
		RelayPin     int  `fig:"RelayPin" validate:"required" json:"RelayPin,omitempty" yaml:"RelayPin" mapstructure:"RelayPin"`
		InverseLogic bool `fig:"InverseLogic" json:"InverseLogic,omitempty" yaml:"InverseLogic" mapstructure:"InverseLogic"`

		// If Type is "Phoenix"
		PhoenixAddress string `fig:"phoenixAddress" json:"phoenixAddress,omitempty" yaml:"phoenixAddress" mapstructure:"phoenixAddress"`
	}

	LedIndicator struct {
		Enabled          bool   `fig:"Enabled" json:"Enabled,omitempty" yaml:"Enabled" mapstructure:"Enabled"`
		DataPin          int    `fig:"DataPin" json:"DataPin,omitempty" yaml:"DataPin" mapstructure:"DataPin"`
		IndicateCardRead bool   `fig:"IndicateCardRead" json:"IndicateCardRead,omitempty" yaml:"IndicateCardRead" mapstructure:"IndicateCardRead"`
		Type             string `fig:"Type" json:"type,omitempty" yaml:"type" mapstructure:"type"`
		Invert           bool   `fig:"Invert" json:"invert,omitempty" yaml:"invert" mapstructure:"invert"`
	}

	TagReader struct {
		IsEnabled   bool   `fig:"IsEnabled" json:"IsEnabled,omitempty" yaml:"IsEnabled" mapstructure:"IsEnabled"`
		ReaderModel string `fig:"ReaderModel" json:"ReaderModel,omitempty" yaml:"ReaderModel" mapstructure:"ReaderModel"`
		Device      string `fig:"Device" json:"device,omitempty" yaml:"device" mapstructure:"device"`
		ResetPin    int    `fig:"ResetPin" json:"ResetPin,omitempty" yaml:"ResetPin" mapstructure:"ResetPin"`
	}

	Display struct {
		IsEnabled  bool   `fig:"IsEnabled" json:"ResetPin,omitempty" yaml:"ResetPin" mapstructure:"ResetPin"`
		Driver     string `fig:"Driver" json:"driver,omitempty" yaml:"driver" mapstructure:"driver"`
		Language   string `fig:"Language" json:"language,omitempty" yaml:"language" mapstructure:"language"`
		I2CAddress string `fig:"I2CAddress" json:"I2CAddress,omitempty" yaml:"I2CAddress" mapstructure:"I2CAddress"`
		I2CBus     int    `fig:"I2CBus" json:"I2CBus,omitempty" yaml:"I2CBus" mapstructure:"I2CBus"`
	}

	PowerMeter struct {
		Enabled              bool    `fig:"Enabled" json:"enabled,omitempty" yaml:"enabled" mapstructure:"enabled"`
		Type                 string  `fig:"Type" json:"type,omitempty" yaml:"type" mapstructure:"type"`
		PowerMeterPin        int     `fig:"PowerMeterPin" json:"PowerMeterPin,omitempty" yaml:"PowerMeterPin" mapstructure:"PowerMeterPin"`
		SpiBus               int     `fig:"SpiBus" default:"0" json:"SpiBus,omitempty" yaml:"SpiBus" mapstructure:"SpiBus"`
		Consumption          float64 `fig:"Consumption" json:"consumption,omitempty" yaml:"consumption" mapstructure:"consumption"`
		ShuntOffset          float64 `fig:"ShuntOffset" json:"ShuntOffset,omitempty" yaml:"ShuntOffset" mapstructure:"ShuntOffset"`
		VoltageDividerOffset float64 `fig:"VoltageDividerOffset" json:"VoltageDividerOffset,omitempty" yaml:"VoltageDividerOffset" mapstructure:"VoltageDividerOffset"`
	}

	PowerMeters struct {
		MinPower int `fig:"MinPower" default:"20" json:"MinPower,omitempty" yaml:"MinPower" mapstructure:"MinPower"`
		Retries  int `fig:"Retries" default:"3" json:"retries,omitempty" yaml:"retries" mapstructure:"retries"`
	}
)
