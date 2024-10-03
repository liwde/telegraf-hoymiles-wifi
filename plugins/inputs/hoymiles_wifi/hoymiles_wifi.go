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

	client *hoymiles_wifi.ClientData
}

func (*HoymilesWifi) SampleConfig() string {
	return sampleConfig
}

func (s *HoymilesWifi) Init() error {
	// set up client
	s.client = hoymiles_wifi.NewClient(s.Hostname, common.DTU_PORT)
	s.Log.Debugf("Connected to %v", s.client.ConnectionInfo)
	return nil
}

func (s *HoymilesWifi) Gather(acc telegraf.Accumulator) error {
	// build request
	request := &models.RealDataNewReqDTO{}
	request.Time = int32(time.Now().Unix())
	request.TimeYmdHms = time.Now().Format("2006-01-02 15:04:05")

	result, err := s.client.GetRealDataNew(request)
	if err != nil {
		return err
	}

	// parse response
	fields := parseResponseData(result)
	tags := map[string]string{}
	timestamp := time.Unix(int64(result.Timestamp), 0)

	acc.AddFields("hoymiles_wifi", fields, tags, timestamp)

	return nil
}

func parseResponseData(data *models.RealDataNewResDTO) map[string]interface{} {
	result := map[string]interface{}{}

	dtu_data := map[string]interface{}{
		"dtu_power":        float64(data.DtuPower) / 10, // Watts
		"dtu_energy_daily": data.DtuDailyEnergy,         // Watt-hours
	}

	maps.Copy(result, dtu_data)
	maps.Copy(result, parseSgsData(data.SgsData))
	maps.Copy(result, parsePvData(data.PvData))

	return result
}

func parseSgsData(sgsData []*models.SGSMO) map[string]interface{} {
	result := map[string]interface{}{}

	for i, v := range sgsData {
		maps.Copy(result, map[string]interface{}{
			"inverter" + strconv.Itoa(i) + "_voltage":     float64(v.Voltage) / 10,     // Volts
			"inverter" + strconv.Itoa(i) + "_frequency":   float64(v.Frequency) / 100,  // Hertz
			"inverter" + strconv.Itoa(i) + "_current":     float64(v.Current) / 10,     // Amps
			"inverter" + strconv.Itoa(i) + "_power":       float64(v.ActivePower) / 10, // Watts
			"inverter" + strconv.Itoa(i) + "_temperature": float64(v.Temperature) / 10, // Degree Celsius
		})
	}

	return result
}

func parsePvData(pvData []*models.PvMO) map[string]interface{} {
	result := map[string]interface{}{}

	for i, v := range pvData {
		maps.Copy(result, map[string]interface{}{
			"pv" + strconv.Itoa(i) + "_voltage":      float64(v.Voltage) / 10,  // Volts
			"pv" + strconv.Itoa(i) + "_current":      float64(v.Current) / 100, // Amps
			"pv" + strconv.Itoa(i) + "_power":        float64(v.Power) / 10,    // Watts
			"pv" + strconv.Itoa(i) + "_energy_daily": v.EnergyDaily,            // Watt-hours
			"pv" + strconv.Itoa(i) + "_energy_total": v.EnergyTotal,            // Watt-hours
		})
	}

	return result
}

func init() {
	inputs.Add("hoymiles_wifi", func() telegraf.Input { return &HoymilesWifi{} })
}
