package tide

var Speeds = map[string]float64{
	"M2": 28.9841042,
	"S2": 30.0000000,
	"N2": 28.4397295,
	"K2": 30.0821373,
	"K1": 15.0410686,
	"O1": 13.9430356,
	"P1": 14.9589314,
	"M4": 57.9682084,
}

type Constituent struct {
	Name      string
	Amplitude float64
	Phase     float64
	Speed     float64
}
