package config

var Params = map[string]interface{}{
	"Nd":        10,
	"Im":        0.5, // interval of messages per a device
	"SCoAPREQ":  417,
	"SCoAPRESP": 150,
	"mean":      0.1,
	"bind":      ":6262",
	"spAdr":     "localhost:6262",
	"exptype":   "tcp",
}

func SetParam(spAdr, bind, exptype string, Im float64, Nd int) {
	Params["spAdr"] = spAdr
	Params["bind"] = bind
	Params["Im"] = Im
	Params["Nd"] = Nd
	Params["exptype"] = exptype
}
