package domain


type Locomotive struct {
    Series  string
    Number  string
    Depo    string
    Records []Record
    Trips   []Trip
}