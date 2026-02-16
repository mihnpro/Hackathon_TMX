package requests
// internal/transport/models/requests/task2.go

// LocomotiveDirectionRequest запрос для получения направления конкретного локомотива
type LocomotiveDirectionRequest struct {
	Series string `uri:"series" binding:"required"`
	Number string `uri:"number" binding:"required"`
}

// DepoBranchesRequest запрос для получения веток депо (задача 3)
type DepoBranchesRequest struct {
	DepoCode string `uri:"depoCode" binding:"required"`
}

// LocomotiveBranchesRequest запрос для получения веток конкретного локомотива (задача 3)
type LocomotiveBranchesRequest struct {
	Series string `uri:"series" binding:"required"`
	Number string `uri:"number" binding:"required"`
}