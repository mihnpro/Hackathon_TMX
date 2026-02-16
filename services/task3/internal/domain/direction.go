package domain

type Direction struct {
    ID          string
    Name        string          // название (например, "Маршрут на Ростов")
    Depo        string          // депо отправления
    Terminal    string          // конечная станция
    TerminalName string         // название конечной станции
    Route       []string        // типичный маршрут
    RouteNames  []string        // названия станций маршрута
    Frequency   int             // как часто используется
    Locomotives map[string]bool // локомотивы, использующие это направление
}