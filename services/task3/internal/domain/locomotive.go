package domain


type Locomotive struct {
	Key     string
	Series  string
	Number  string
	Depo    string
	Records []Record
	Trips   []Trip
}