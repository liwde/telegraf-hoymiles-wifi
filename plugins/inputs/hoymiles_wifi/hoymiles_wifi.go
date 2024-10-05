package hoymiles_wifi

import (
	_ "embed"
	"maps"
	"strconv"
	"time"

	"github.com/BLun78/hoymiles_wifi"
	"github.com/BLun78/hoymiles_wifi/common"
	"github.com/BLun78/hoymiles_wifi/hoymiles/models"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type HoymilesWifi struct {
	Hostname string          `toml:"hostname"`
	Log      telegraf.Logger `toml:"-"`
}

func (*HoymilesWifi) SampleConfig() string {
	return sampleConfig
}

func (s *HoymilesWifi) Init() error {
	return nil
}

func (s *HoymilesWifi) Gather(acc telegraf.Accumulator) error {
	client := hoymiles_wifi.NewClient(s.Hostname, common.DTU_PORT)
	defer client.CloseConnection()

	// build request
	request := &models.RealDataNewReqDTO{}
	request.Time = int32(time.Now().Unix())
	request.TimeYmdHms = time.Now().Format("2006-01-02 15:04:05")

	result, err := client.GetRealDataNew(request)
	if err != nil {
		return err
	}

	// handle response
	processResponseData(acc, result)

	return nil
}

func processResponseData(acc telegraf.Accumulator, data *models.RealDataNewResDTO) {
	timestamp := time.Unix(int64(data.Timestamp), 0)

	dtuFields := map[string]interface{}{
		"dtu_power":        float64(data.DtuPower) / 10, // Watts
		"dtu_energy_daily": data.DtuDailyEnergy,         // Watt-hours
	}
	dtuTags := map[string]string{
		"dtu_serial_number": data.DeviceSerialNumber,
	}

	acc.AddFields("hoymiles_dtu", dtuFields, dtuTags, timestamp)

	processSgsData(acc, data.SgsData, dtuTags, timestamp)
	processPvData(acc, data.PvData, dtuTags, timestamp)

}

func processSgsData(acc telegraf.Accumulator, sgsData []*models.SGSMO, dtuTags map[string]string, timestamp time.Time) {
	for _, v := range sgsData {
		inverterFields := map[string]interface{}{
			"inverter_voltage":     float64(v.Voltage) / 10,     // Volts
			"inverter_frequency":   float64(v.Frequency) / 100,  // Hertz
			"inverter_current":     float64(v.Current) / 100,    // Amps
			"inverter_power":       float64(v.ActivePower) / 10, // Watts
			"inverter_temperature": float64(v.Temperature) / 10, // Degree Celsius
		}

		inverterTags := maps.Clone(dtuTags)
		inverterTags["inverter_serial_number"] = strconv.FormatInt(v.SerialNumber, 10)

		acc.AddFields("hoymiles_inverter", inverterFields, inverterTags, timestamp)
	}
}

func processPvData(acc telegraf.Accumulator, pvData []*models.PvMO, dtuTags map[string]string, timestamp time.Time) {
	for _, v := range pvData {
		pvFields := map[string]interface{}{
			"pv_voltage":      float64(v.Voltage) / 10,  // Volts
			"pv_current":      float64(v.Current) / 100, // Amps
			"pv_power":        float64(v.Power) / 10,    // Watts
			"pv_energy_daily": v.EnergyDaily,            // Watt-hours
			"pv_energy_total": v.EnergyTotal,            // Watt-hours
		}

		pvTags := maps.Clone(dtuTags)
		pvTags["inverter_serial_number"] = strconv.FormatInt(v.SerialNumber, 10)
		pvTags["inverter_port_number"] = strconv.FormatInt(int64(v.PortNumber), 10)

		acc.AddFields("hoymiles_pv", pvFields, pvTags, timestamp)
	}
}

func init() {
	inputs.Add("hoymiles_wifi", func() telegraf.Input { return &HoymilesWifi{} })
}
